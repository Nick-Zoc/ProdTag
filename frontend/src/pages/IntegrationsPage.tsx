import {useEffect, useState} from 'react';
import type {ReactNode} from 'react';
import {Activity, CheckCircle2, Copy, Plug, Square, Terminal, XCircle} from 'lucide-react';
import {GetPlaybackStatus, ListHandledEventLog, StopPlayback} from '../../wailsjs/go/main/App';
import {Badge} from '../components/Badge';
import {Button} from '../components/Button';
import {Card} from '../components/Card';
import {CollapsibleSection} from '../components/CollapsibleSection';
import {AppConfig, PlaybackStatus, RecentEventRecord} from '../types/app';

const setupCommands = {
  buildHelper: 'go build -tags prodtaghelper -o build/bin/prodtag-helper .',
  exportHelper: 'export PRODTAG_HELPER="$PWD/build/bin/prodtag-helper"',
  sourceScript: 'source "$PWD/scripts/prodtag.zsh"',
  testEmit: 'build/bin/prodtag-helper emit --event-type command_success --command "ls -a" --exit-code 0 --cwd "$PWD"',
};

export function IntegrationsPage({config}: {config: AppConfig}) {
  const [playbackStatus, setPlaybackStatus] = useState<PlaybackStatus | null>(null);
  const [recentEvents, setRecentEvents] = useState<RecentEventRecord[]>([]);
  const [isStopping, setIsStopping] = useState(false);
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null);

  const integrations = [
    ['zsh', config.integrations.zsh],
    ['bash', config.integrations.bash],
    ['PowerShell', config.integrations.powerShell],
  ] as const;

  useEffect(() => {
    void refreshStatus();
  }, []);

  async function refreshStatus() {
    const [status, events] = await Promise.all([GetPlaybackStatus(), ListHandledEventLog()]);
    setPlaybackStatus(status as PlaybackStatus);
    setRecentEvents(events as RecentEventRecord[]);
  }

  async function stopPlayback() {
    setIsStopping(true);
    try {
      const status = await StopPlayback();
      setPlaybackStatus(status as PlaybackStatus);
    } finally {
      setIsStopping(false);
    }
  }

  async function copyCommand(key: string, value: string) {
    try {
      await navigator.clipboard.writeText(value);
      setCopiedCommand(key);
      window.setTimeout(() => setCopiedCommand(null), 1600);
    } catch {
      setCopiedCommand(null);
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Shell integrations</h2>
            <p className="mt-1 text-sm text-neutral-500">
              zsh can be wired manually on macOS. Automatic install and doctor checks arrive in a later phase.
            </p>
          </div>
          <Badge tone="amber">Manual setup</Badge>
        </div>
        <div className="mt-5 grid gap-3 md:grid-cols-3">
          {integrations.map(([name, integration]) => (
            <div key={name} className="rounded-lg border border-neutral-200 bg-neutral-50 p-4">
              <div className="flex items-center justify-between gap-3">
                <div className="text-base font-semibold">{name}</div>
                <Badge tone={integration.installed ? 'green' : 'neutral'}>
                  {integration.installed ? 'Installed' : 'Not installed'}
                </Badge>
              </div>
              <div className="mt-3 text-sm text-neutral-500">
                {integration.scope || 'Install, uninstall, and doctor checks will appear here.'}
              </div>
            </div>
          ))}
        </div>

        <div className="mt-5 rounded-lg border border-amber-200 bg-amber-50 p-4">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <div className="flex items-center gap-2 text-sm font-semibold text-amber-950">
                <Terminal size={16} />
                macOS zsh MVP available
              </div>
              <p className="mt-1 text-sm text-amber-900">
                Build the helper and source the script manually. Automatic install/uninstall is coming later.
              </p>
            </div>
            <Badge tone="amber">Manual setup</Badge>
          </div>
          <div className="mt-4 grid gap-2">
            <SetupCommand
              command={setupCommands.buildHelper}
              copied={copiedCommand === 'build-helper'}
              label="Build helper"
              onCopy={() => copyCommand('build-helper', setupCommands.buildHelper)}
            />
            <SetupCommand
              command={setupCommands.exportHelper}
              copied={copiedCommand === 'export-helper'}
              label="Set helper path"
              onCopy={() => copyCommand('export-helper', setupCommands.exportHelper)}
            />
            <SetupCommand
              command={setupCommands.sourceScript}
              copied={copiedCommand === 'source-script'}
              label="Source script"
              onCopy={() => copyCommand('source-script', setupCommands.sourceScript)}
            />
            <SetupCommand
              command={setupCommands.testEmit}
              copied={copiedCommand === 'test-emit'}
              label="Test emit"
              onCopy={() => copyCommand('test-emit', setupCommands.testEmit)}
            />
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Event engine and playback</h2>
            <p className="mt-1 text-sm text-neutral-500">
              The helper can send local zsh events through the same matcher and backend playback path.
            </p>
          </div>
          <Badge tone={config.eventEngineEnabled ? 'green' : 'amber'}>
            {config.eventEngineEnabled ? 'Enabled' : 'Paused'}
          </Badge>
        </div>

        <div className="mt-5 grid gap-3 md:grid-cols-3">
          <StatusTile
            icon={<Activity size={17} />}
            label="Event engine"
            tone={config.eventEngineEnabled ? 'green' : 'amber'}
            value={config.eventEngineEnabled ? 'Enabled' : 'Disabled'}
          />
          <StatusTile
            icon={playbackStatus?.supported ? <CheckCircle2 size={17} /> : <XCircle size={17} />}
            label="Backend playback"
            tone={playbackStatus?.supported && config.playbackEnabled ? 'green' : 'amber'}
            value={config.playbackEnabled ? playbackStatus?.method || 'Checking...' : 'Disabled'}
          />
          <StatusTile
            icon={<Plug size={17} />}
            label="CLI intake"
            tone="amber"
            value="Manual zsh"
          />
        </div>

        {playbackStatus && (
          <div className="mt-4 rounded-lg bg-neutral-50 px-3 py-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <div className="text-sm font-semibold text-neutral-800">
                  {playbackStatus.platform} / {playbackStatus.method}
                </div>
                <p className="mt-1 text-sm text-neutral-500">{playbackStatus.message}</p>
              </div>
              <Button disabled={!playbackStatus.playing} isLoading={isStopping} leftIcon={<Square size={15} />} onClick={stopPlayback} variant="secondary">
                Stop playback
              </Button>
            </div>
          </div>
        )}
      </Card>

      <CollapsibleSection
        action={<Badge>{recentEvents.length}</Badge>}
        title="Recent handled events"
        description="Diagnostics from the in-app event handling path."
      >
        <div className="grid gap-2">
          {recentEvents.length === 0 ? (
            <p className="rounded-lg bg-neutral-50 px-3 py-2 text-sm text-neutral-500">
              No events handled yet. Use Test full event flow in the Rules simulator to exercise the backend path.
            </p>
          ) : (
            recentEvents.slice(0, 6).map((event) => (
              <div key={event.id} className="rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-2">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <div className="text-sm font-semibold text-neutral-800">{event.event.eventType}</div>
                    <div className="mt-1 text-xs text-neutral-500">{event.event.command || 'No command'}</div>
                  </div>
                  <Badge tone={event.playbackError ? 'rose' : event.matched ? 'green' : 'neutral'}>
                    {event.playbackError ? 'Playback error' : event.matched ? 'Matched' : 'No match'}
                  </Badge>
                </div>
                <p className="mt-1 text-xs text-neutral-500">
                  {event.ruleName || event.message}{event.soundName ? ` - ${event.soundName}` : ''}
                </p>
              </div>
            ))
          )}
        </div>
      </CollapsibleSection>
    </div>
  );
}

function StatusTile({
  icon,
  label,
  tone,
  value,
}: {
  icon: ReactNode;
  label: string;
  tone: 'green' | 'amber' | 'rose' | 'neutral';
  value: string;
}) {
  return (
    <div className="rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-3">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-sm font-semibold text-neutral-700">
          {icon}
          {label}
        </div>
        <Badge tone={tone}>{value}</Badge>
      </div>
    </div>
  );
}

function SetupCommand({
  command,
  copied,
  label,
  onCopy,
}: {
  command: string;
  copied: boolean;
  label: string;
  onCopy: () => void;
}) {
  return (
    <div className="grid gap-2 rounded-lg bg-white px-3 py-2 md:grid-cols-[150px_minmax(0,1fr)_auto] md:items-center">
      <div className="text-xs font-semibold uppercase text-neutral-500">{label}</div>
      <code className="min-w-0 truncate rounded-md bg-neutral-100 px-2 py-1 font-mono text-xs text-neutral-700" title={command}>
        {command}
      </code>
      <Button className="h-8 px-3 text-xs" leftIcon={<Copy size={13} />} onClick={onCopy} variant="ghost">
        {copied ? 'Copied' : 'Copy'}
      </Button>
    </div>
  );
}
