import {useEffect, useState} from 'react';
import type {ReactNode} from 'react';
import {Activity, CheckCircle2, Copy, Plug, RefreshCw, Square, Terminal, Trash2, XCircle} from 'lucide-react';
import {GetPlaybackStatus, InstallShellIntegration, ListHandledEventLog, ListIntegrationStatuses, RunIntegrationDoctor, StopPlayback, UninstallShellIntegration} from '../../wailsjs/go/main/App';
import {Badge} from '../components/Badge';
import {Button} from '../components/Button';
import {Card} from '../components/Card';
import {CollapsibleSection} from '../components/CollapsibleSection';
import {Toast, ToastState} from '../components/Toast';
import {AppConfig, DoctorResult, IntegrationStatus, PlaybackStatus, RecentEventRecord} from '../types/app';

export function IntegrationsPage({config}: {config: AppConfig}) {
  const [playback, setPlayback] = useState<PlaybackStatus | null>(null);
  const [integrations, setIntegrations] = useState<IntegrationStatus[]>([]);
  const [doctor, setDoctor] = useState<DoctorResult | null>(null);
  const [recentEvents, setRecentEvents] = useState<RecentEventRecord[]>([]);
  const [busyAction, setBusyAction] = useState<string | null>(null);
  const [copied, setCopied] = useState<string | null>(null);
  const [toast, setToast] = useState<ToastState | null>(null);

  useEffect(() => { void refreshStatus(); }, []);

  async function refreshStatus() {
    try {
      const [nextPlayback, nextIntegrations, events] = await Promise.all([GetPlaybackStatus(), ListIntegrationStatuses(), ListHandledEventLog()]);
      setPlayback(nextPlayback as PlaybackStatus);
      setIntegrations(nextIntegrations as IntegrationStatus[]);
      setRecentEvents(events as RecentEventRecord[]);
    } catch (error) {
      setToast({message: String(error), tone: 'rose'});
    }
  }

  async function runShellAction(shell: string, action: 'install' | 'uninstall') {
    const key = `${shell}-${action}`;
    setBusyAction(key);
    try {
      const result = action === 'install' ? await InstallShellIntegration(shell) : await UninstallShellIntegration(shell);
      const next = result as DoctorResult;
      setDoctor(next);
      setIntegrations(next.integrations);
      setPlayback(next.playback);
      setToast({message: action === 'install' ? `${shellLabel(shell)} installed. Open a new terminal to activate it.` : `${shellLabel(shell)} profile integration removed.`, tone: 'green'});
    } catch (error) {
      setToast({message: String(error), tone: 'rose'});
    } finally {
      setBusyAction(null);
    }
  }

  async function runDoctor() {
    setBusyAction('doctor');
    try {
      const result = await RunIntegrationDoctor() as DoctorResult;
      setDoctor(result); setIntegrations(result.integrations); setPlayback(result.playback);
      setToast({message: result.ok ? 'Doctor found a healthy runtime.' : 'Doctor found setup items that need attention.', tone: result.ok ? 'green' : 'amber'});
    } catch (error) { setToast({message: String(error), tone: 'rose'}); }
    finally { setBusyAction(null); }
  }

  async function stopPlayback() {
    setBusyAction('stop');
    try { setPlayback(await StopPlayback() as PlaybackStatus); }
    finally { setBusyAction(null); }
  }

  async function copyCommand(key: string, value?: string) {
    if (!value) return;
    try { await navigator.clipboard.writeText(value); setCopied(key); setToast({message: 'Command copied.', tone: 'green'}); window.setTimeout(() => setCopied(null), 1600); }
    catch { setToast({message: 'Could not copy the command.', tone: 'rose'}); }
  }

  return (
    <div className="space-y-6">
      {toast && <Toast toast={toast} onDismiss={() => setToast(null)} />}
      <Card>
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Shell integrations</h2>
            <p className="mt-1 text-sm text-neutral-500">Install a profile integration for new terminals or copy a command for the current session.</p>
          </div>
          <Button isLoading={busyAction === 'doctor'} leftIcon={<RefreshCw size={15} />} onClick={runDoctor} variant="secondary">Run doctor</Button>
        </div>
        <div className="mt-5 grid gap-4">
          {integrations.map((integration) => (
            <IntegrationCard busyAction={busyAction} copied={copied} integration={integration} key={integration.shell} onAction={runShellAction} onCopy={copyCommand} />
          ))}
        </div>
      </Card>

      <Card>
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div><h2 className="text-lg font-semibold">Runtime and playback</h2><p className="mt-1 text-sm text-neutral-500">The helper uses the same matcher, config, event log, and playback service as the desktop app.</p></div>
          <Badge tone={playback?.supported ? 'green' : 'amber'}>{playback?.supported ? playback.method : 'Unavailable'}</Badge>
        </div>
        <div className="mt-5 grid gap-3 md:grid-cols-3">
          <StatusTile icon={<Activity size={17} />} label="Event engine" tone={config.eventEngineEnabled ? 'green' : 'amber'} value={config.eventEngineEnabled ? 'Enabled' : 'Disabled'} />
          <StatusTile icon={playback?.supported ? <CheckCircle2 size={17} /> : <XCircle size={17} />} label="Playback" tone={playback?.supported && config.playbackEnabled ? 'green' : 'amber'} value={config.playbackEnabled ? playback?.method || 'Checking...' : 'Disabled'} />
          <StatusTile icon={<Plug size={17} />} label="Detected shells" tone="neutral" value={`${integrations.filter((item) => item.shellExecutableFound).length} of 3`} />
        </div>
        {playback && <div className="mt-4 rounded-lg bg-neutral-50 p-3"><div className="flex flex-wrap items-center justify-between gap-3"><div><div className="text-sm font-semibold">{playback.platform} / {playback.method}</div><p className="mt-1 text-sm text-neutral-500">{playback.message}</p>{playback.alternatives?.length > 1 && <p className="mt-1 text-xs text-neutral-500">Detected: {playback.alternatives.join(', ')}</p>}</div><Button disabled={!playback.playing} isLoading={busyAction === 'stop'} leftIcon={<Square size={15} />} onClick={stopPlayback} variant="secondary">Stop playback</Button></div></div>}
        {!!playback?.suggestions?.length && !playback.supported && <div className="mt-3 grid gap-2">{playback.suggestions.map((suggestion) => <SetupCommand command={suggestion.command || suggestion.detail} copied={copied === `dependency-${suggestion.platform}`} key={`${suggestion.platform}-${suggestion.label}`} label={`${suggestion.platform}: ${suggestion.label}`} onCopy={() => copyCommand(`dependency-${suggestion.platform}`, suggestion.command)} />)}</div>}
      </Card>

      {doctor && <CollapsibleSection action={<Badge tone={doctor.ok ? 'green' : 'amber'}>{doctor.ok ? 'Healthy' : 'Review'}</Badge>} description="Runtime settings, matcher cache, playback, shell profiles, and installed assets." title="Doctor summary"><div className="grid gap-2 md:grid-cols-2"><DoctorLine label="Matcher cache" ok={doctor.matcherCacheValid} value={doctor.matcherCacheValid ? 'Valid' : 'Invalid'} /><DoctorLine label="Rules" ok={doctor.ruleCount > 0} value={`${doctor.ruleCount} configured`} /><DoctorLine label="Listening / engine" ok={doctor.listening && doctor.eventEngineEnabled} value={`${doctor.listening ? 'Listening' : 'Paused'} / ${doctor.eventEngineEnabled ? 'enabled' : 'disabled'}`} /><DoctorLine label="Playback controls" ok={doctor.playbackEnabled && !doctor.muted} value={`${doctor.playbackEnabled ? 'Enabled' : 'Disabled'} / ${doctor.muted ? 'muted' : 'audible'}`} /></div></CollapsibleSection>}

      <CollapsibleSection action={<Badge>{recentEvents.length}</Badge>} description="The newest retained helper and in-app events." title="Recent handled events"><div className="grid gap-2">{recentEvents.length === 0 ? <p className="rounded-lg bg-neutral-50 px-3 py-2 text-sm text-neutral-500">No events handled yet.</p> : recentEvents.slice(0, 6).map((event) => <div className="rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-2" key={event.id}><div className="flex items-center justify-between gap-2"><div className="min-w-0"><div className="text-sm font-semibold">{event.event.eventType}</div><div className="truncate text-xs text-neutral-500" title={event.event.command}>{event.event.command || 'No command'}</div></div><Badge tone={event.playbackError ? 'rose' : event.matched ? 'green' : 'neutral'}>{event.playbackError ? 'Error' : event.matched ? 'Matched' : 'No match'}</Badge></div></div>)}</div></CollapsibleSection>
    </div>
  );
}

function IntegrationCard({busyAction, copied, integration, onAction, onCopy}: {busyAction: string | null; copied: string | null; integration: IntegrationStatus; onAction: (shell: string, action: 'install' | 'uninstall') => void; onCopy: (key: string, value?: string) => void}) {
  const partial = integration.profileState === 'partial';
  return <section className="rounded-lg border border-neutral-200 bg-neutral-50 p-4">
    <div className="flex flex-wrap items-start justify-between gap-3"><div><div className="flex items-center gap-2"><Terminal size={17} /><h3 className="font-semibold">{integration.displayName}</h3>{integration.platformRelevant && <Badge tone="green">Recommended here</Badge>}</div><p className="mt-1 text-sm text-neutral-500">{integration.supported ? 'Available for current-user profile integration.' : integration.problems[0] || 'Unavailable on this machine.'}</p></div><Badge tone={integration.profileConfigured ? 'green' : partial ? 'rose' : integration.supported ? 'amber' : 'neutral'}>{integration.profileConfigured ? 'Installed' : partial ? 'Needs repair' : integration.supported ? 'Not installed' : 'Unavailable'}</Badge></div>
    <div className="mt-4 grid gap-2 sm:grid-cols-3"><MiniStatus label="Shell" ok={integration.shellExecutableFound} value={integration.shellExecutableFound ? 'Available' : 'Missing'} /><MiniStatus label="Helper" ok={integration.helperInstalled && integration.helperExecutable} value={integration.helperExecutable ? 'Ready' : 'Missing'} /><MiniStatus label="Script / profile" ok={integration.scriptInstalled && integration.profileConfigured} value={`${integration.scriptInstalled ? 'Script ready' : 'Script missing'} / ${integration.profileState.replace('_', ' ')}`} /></div>
    <div className="mt-4 flex flex-wrap gap-2"><Button disabled={!integration.supported || partial} isLoading={busyAction === `${integration.shell}-install`} leftIcon={<Plug size={14} />} onClick={() => onAction(integration.shell, 'install')}>Install</Button><Button disabled={!integration.profileConfigured} isLoading={busyAction === `${integration.shell}-uninstall`} leftIcon={<Trash2 size={14} />} onClick={() => onAction(integration.shell, 'uninstall')} variant="secondary">Uninstall</Button></div>
    <details className="mt-4 rounded-lg border border-neutral-200 bg-white"><summary className="cursor-pointer px-3 py-2 text-sm font-semibold text-neutral-700">Session and debug commands</summary><div className="grid gap-2 border-t border-neutral-200 p-3"><SetupCommand command={integration.currentSessionCommand} copied={copied === `${integration.shell}-enable`} disabled={!integration.scriptInstalled || !integration.helperExecutable} label="Enable current session" onCopy={() => onCopy(`${integration.shell}-enable`, integration.currentSessionCommand)} /><SetupCommand command={integration.disableSessionCommand} copied={copied === `${integration.shell}-disable`} label="Disable current session" onCopy={() => onCopy(`${integration.shell}-disable`, integration.disableSessionCommand)} /><SetupCommand command={integration.debugEnableCommand} copied={copied === `${integration.shell}-debug`} label="Enable debug" onCopy={() => onCopy(`${integration.shell}-debug`, integration.debugEnableCommand)} /><SetupCommand command={integration.debugLogCommand} copied={copied === `${integration.shell}-log`} label="Inspect debug log" onCopy={() => onCopy(`${integration.shell}-log`, integration.debugLogCommand)} /></div></details>
  </section>;
}

function MiniStatus({label, ok, value}: {label: string; ok: boolean; value: string}) { return <div className="rounded-md bg-white px-3 py-2"><div className="flex items-center gap-2 text-xs font-semibold text-neutral-500">{ok ? <CheckCircle2 className="text-emerald-600" size={13} /> : <XCircle className="text-neutral-400" size={13} />}{label}</div><div className="mt-1 truncate text-sm text-neutral-700" title={value}>{value}</div></div>; }
function DoctorLine({label, ok, value}: {label: string; ok: boolean; value: string}) { return <div className="rounded-lg bg-neutral-50 px-3 py-2"><div className="flex items-center gap-2 text-sm font-semibold">{ok ? <CheckCircle2 className="text-emerald-600" size={15} /> : <XCircle className="text-amber-600" size={15} />}{label}</div><div className="mt-1 text-xs text-neutral-500">{value}</div></div>; }
function StatusTile({icon, label, tone, value}: {icon: ReactNode; label: string; tone: 'green' | 'amber' | 'rose' | 'neutral'; value: string}) { return <div className="rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-3"><div className="flex items-center justify-between gap-3"><div className="flex items-center gap-2 text-sm font-semibold text-neutral-700">{icon}{label}</div><Badge tone={tone}>{value}</Badge></div></div>; }
function SetupCommand({command, copied, disabled = false, label, onCopy}: {command: string; copied: boolean; disabled?: boolean; label: string; onCopy: () => void}) { return <div className="grid gap-2 rounded-lg bg-neutral-50 px-3 py-2 md:grid-cols-[170px_minmax(0,1fr)_auto] md:items-center"><div className="text-xs font-semibold uppercase text-neutral-500">{label}</div><code className="min-w-0 truncate rounded-md bg-white px-2 py-1 font-mono text-xs text-neutral-700" title={command}>{command}</code><Button className="h-8 px-3 text-xs" disabled={disabled} leftIcon={<Copy size={13} />} onClick={onCopy} variant="ghost">{copied ? 'Copied' : 'Copy'}</Button></div>; }
function shellLabel(shell: string) { return shell === 'powershell' ? 'PowerShell' : shell; }
