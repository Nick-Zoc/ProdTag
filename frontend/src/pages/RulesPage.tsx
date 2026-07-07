import {useEffect, useMemo, useRef, useState} from 'react';
import {BellRing, CheckCircle2, Edit3, Play, Plus, RotateCcw, Save, SlidersHorizontal, Trash2, X, XCircle} from 'lucide-react';
import {
  ClearRecentEvents,
  CreateRule,
  DeleteRule,
  GetSoundPreviewDataURL,
  HandleTerminalEvent,
  ListRecentEvents,
  SimulateEvent,
  TestRuleSound,
  ToggleRule,
  UpdateRule,
} from '../../wailsjs/go/main/App';
import {Badge} from '../components/Badge';
import {Button} from '../components/Button';
import {Card} from '../components/Card';
import {CollapsibleSection} from '../components/CollapsibleSection';
import {CompactPath} from '../components/CompactPath';
import {ConfirmDialog} from '../components/ConfirmDialog';
import {EmptyState} from '../components/EmptyState';
import {Toast, ToastState} from '../components/Toast';
import {AppConfig, ConfigSnapshot, RecentEventRecord, RuleMatchResult, RuleRecord, SoundRecord} from '../types/app';

type RulesPageProps = {
  config: AppConfig;
  onConfigUpdated: (snapshot: ConfigSnapshot) => void;
};

type RuleFormState = {
  id: string | null;
  name: string;
  enabled: boolean;
  eventType: string;
  soundId: string;
  matchMode: string;
  commandPattern: string;
  exitCode: string;
};

type SimulatorFormState = {
  eventType: string;
  command: string;
  exitCode: string;
  cwd: string;
  durationMs: string;
};

const emptyForm: RuleFormState = {
  id: null,
  name: '',
  enabled: true,
  eventType: 'command_success',
  soundId: '',
  matchMode: 'any',
  commandPattern: '',
  exitCode: '',
};

const emptySimulator: SimulatorFormState = {
  eventType: 'command_success',
  command: 'npm test',
  exitCode: '0',
  cwd: '',
  durationMs: '',
};

const eventTypes = [
  ['command_success', 'Command success', 'Any matching command exits with code 0.'],
  ['command_failure', 'Command failure', 'Any matching command exits with a non-zero code.'],
  ['git_commit_success', 'Git commit success', 'A git commit command succeeds.'],
  ['git_push_success', 'Git push success', 'A git push command succeeds.'],
  ['test_success', 'Test success', 'A test command succeeds.'],
  ['test_failure', 'Test failure', 'A test command fails.'],
  ['build_success', 'Build success', 'A build command succeeds.'],
  ['build_failure', 'Build failure', 'A build command fails.'],
] as const;

const matchModes = [
  ['any', 'Any command'],
  ['exact', 'Exact'],
  ['startsWith', 'Starts with'],
  ['endsWith', 'Ends with'],
  ['contains', 'Contains'],
  ['regex', 'Regex'],
] as const;

export function RulesPage({config, onConfigUpdated}: RulesPageProps) {
  const [form, setForm] = useState<RuleFormState>(() => ({...emptyForm}));
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [toast, setToast] = useState<ToastState | null>(null);
  const [confirmRule, setConfirmRule] = useState<RuleRecord | null>(null);
  const [savingRuleId, setSavingRuleId] = useState<string | null>(null);
  const [deletingRuleId, setDeletingRuleId] = useState<string | null>(null);
  const [testingRuleId, setTestingRuleId] = useState<string | null>(null);
  const [simulator, setSimulator] = useState<SimulatorFormState>(emptySimulator);
  const [simulationResult, setSimulationResult] = useState<RuleMatchResult | null>(null);
  const [recentEvents, setRecentEvents] = useState<RecentEventRecord[]>([]);
  const [isSimulating, setIsSimulating] = useState(false);
  const [isHandlingEvent, setIsHandlingEvent] = useState(false);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  const soundsById = useMemo(() => {
    return new Map(config.sounds.map((sound) => [sound.id, sound]));
  }, [config.sounds]);

  const selectedSound = form.soundId ? soundsById.get(form.soundId) ?? null : null;
  const hasSounds = config.sounds.length > 0;
  const enabledRules = config.rules.filter((rule) => rule.enabled).length;

  useEffect(() => {
    if (!toast || toast.tone === 'rose') {
      return;
    }
    const timeout = window.setTimeout(() => setToast(null), 2600);
    return () => window.clearTimeout(timeout);
  }, [toast]);

  useEffect(() => {
    void refreshRecentEvents();
  }, []);

  async function refreshRecentEvents() {
    try {
      const events = await ListRecentEvents();
      setRecentEvents(events as RecentEventRecord[]);
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    }
  }

  function openCreateForm() {
    setForm({...emptyForm});
    setIsFormOpen(true);
  }

  function openEditForm(rule: RuleRecord) {
    setForm({
      id: rule.id,
      name: rule.name,
      enabled: rule.enabled,
      eventType: rule.eventType,
      soundId: rule.soundId,
      matchMode: rule.matchMode ?? '',
      commandPattern: rule.commandPattern ?? '',
      exitCode: rule.exitCode === null || rule.exitCode === undefined ? '' : String(rule.exitCode),
    });
    setIsFormOpen(true);
  }

  function closeForm() {
    setIsFormOpen(false);
    setForm({...emptyForm});
  }

  async function saveRule() {
    const validation = validateForm(form, hasSounds);
    if (validation) {
      setToast({message: validation, tone: 'rose'});
      return;
    }

    const request = {
      id: form.id ?? '',
      name: form.name.trim(),
      enabled: form.enabled,
      eventType: form.eventType,
      soundId: form.soundId,
      matchMode: form.matchMode,
      commandPattern: form.commandPattern.trim(),
      exitCode: form.exitCode.trim() === '' ? undefined : Number(form.exitCode),
    };

    setSavingRuleId(form.id ?? 'new');
    try {
      const snapshot = form.id ? await UpdateRule(request) : await CreateRule(request);
      onConfigUpdated(snapshot as ConfigSnapshot);
      setToast({message: form.id ? 'Rule updated' : 'Rule created', tone: 'green'});
      closeForm();
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    } finally {
      setSavingRuleId(null);
    }
  }

  async function toggleRule(rule: RuleRecord, enabled: boolean) {
    setSavingRuleId(rule.id);
    try {
      const snapshot = await ToggleRule(rule.id, enabled);
      onConfigUpdated(snapshot as ConfigSnapshot);
      setToast({message: enabled ? 'Rule enabled' : 'Rule disabled', tone: 'green'});
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    } finally {
      setSavingRuleId(null);
    }
  }

  async function deleteRule() {
    if (!confirmRule) {
      return;
    }

    setDeletingRuleId(confirmRule.id);
    try {
      const snapshot = await DeleteRule(confirmRule.id);
      onConfigUpdated(snapshot as ConfigSnapshot);
      setToast({message: 'Rule deleted', tone: 'green'});
      setConfirmRule(null);
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    } finally {
      setDeletingRuleId(null);
    }
  }

  async function playRule(rule: RuleRecord) {
    setTestingRuleId(rule.id);
    try {
      stopPreview();
      const dataURL = await TestRuleSound(rule.id);
      const audio = new Audio(dataURL);
      audioRef.current = audio;
      audio.addEventListener('ended', () => setTestingRuleId(null), {once: true});
      await audio.play();
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
      setTestingRuleId(null);
    }
  }

  async function playSelectedSound() {
    if (!selectedSound) {
      setToast({message: 'Choose a sound before testing playback', tone: 'rose'});
      return;
    }

    setTestingRuleId('form');
    try {
      stopPreview();
      const dataURL = await GetSoundPreviewDataURL(selectedSound.id);
      const audio = new Audio(dataURL);
      audioRef.current = audio;
      audio.addEventListener('ended', () => setTestingRuleId(null), {once: true});
      await audio.play();
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
      setTestingRuleId(null);
    }
  }

  async function simulateEvent() {
    const event = simulatorEvent(simulator);
    if (typeof event === 'string') {
      setToast({message: event, tone: 'rose'});
      return;
    }

    setIsSimulating(true);
    try {
      const result = await SimulateEvent(event);
      setSimulationResult(result as RuleMatchResult);
      if ((result as RuleMatchResult).matched) {
        setToast({message: 'Event matched a rule', tone: 'green'});
      }
      await refreshRecentEvents();
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    } finally {
      setIsSimulating(false);
    }
  }

  async function handleEventWithPlayback() {
    const event = simulatorEvent(simulator);
    if (typeof event === 'string') {
      setToast({message: event, tone: 'rose'});
      return;
    }

    setIsHandlingEvent(true);
    try {
      const result = await HandleTerminalEvent(event);
      setSimulationResult(result as RuleMatchResult);
      const handled = result as RuleMatchResult;
      if (handled.playbackError || handled.matched) {
        setToast({
          message: handled.playbackError || handled.message,
          tone: handled.playbackError ? 'rose' : 'green',
        });
      }
      await refreshRecentEvents();
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    } finally {
      setIsHandlingEvent(false);
    }
  }

  async function clearRecentEvents() {
    try {
      const events = await ClearRecentEvents();
      setRecentEvents(events as RecentEventRecord[]);
      setToast({message: 'Recent events cleared', tone: 'green'});
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    }
  }

  async function playSimulationSound() {
    if (!simulationResult?.sound) {
      setToast({message: 'No matched sound to play', tone: 'rose'});
      return;
    }

    setTestingRuleId('simulation');
    try {
      stopPreview();
      const dataURL = await GetSoundPreviewDataURL(simulationResult.sound.id);
      const audio = new Audio(dataURL);
      audioRef.current = audio;
      audio.addEventListener('ended', () => setTestingRuleId(null), {once: true});
      await audio.play();
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
      setTestingRuleId(null);
    }
  }

  function stopPreview() {
    if (audioRef.current) {
      audioRef.current.pause();
      audioRef.current.currentTime = 0;
      audioRef.current = null;
    }
  }

  return (
    <div className="space-y-5">
      <Card>
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Rules / Events</h2>
            <p className="mt-1 text-sm text-neutral-500">
              Map future terminal events to sounds. Shell detection arrives in a later phase.
            </p>
          </div>
          <Button disabled={!hasSounds} leftIcon={<Plus size={16} />} onClick={openCreateForm} variant="secondary">
            New rule
          </Button>
        </div>

        <div className="mt-4 grid gap-3 md:grid-cols-3">
          <RuleStat label="Total rules" value={config.rules.length} />
          <RuleStat label="Enabled" value={enabledRules} />
          <RuleStat label="Available sounds" value={config.sounds.length} />
        </div>

        {!hasSounds && (
          <EmptyState
            icon={<BellRing size={20} />}
            title="Import sounds before creating rules"
            body="Rules need a sound assignment. You can still view existing rules here if a sound is removed later."
          />
        )}

        {toast && <Toast toast={toast} onDismiss={() => setToast(null)} />}
      </Card>

      {config.rules.length === 0 ? (
        <Card>
          <EmptyState
            icon={<SlidersHorizontal size={20} />}
            title="No rules configured"
            body="Create rules for command success, command failure, Git, tests, and builds once you have sounds in the library."
          />
        </Card>
      ) : (
        <section className="grid gap-4">
          {config.rules.map((rule) => (
            <RuleCard
              key={rule.id}
              rule={rule}
              sound={soundsById.get(rule.soundId) ?? null}
              isBusy={savingRuleId === rule.id}
              isTesting={testingRuleId === rule.id}
              onDelete={() => setConfirmRule(rule)}
              onEdit={() => openEditForm(rule)}
              onTest={() => playRule(rule)}
              onToggle={(enabled) => toggleRule(rule, enabled)}
            />
          ))}
        </section>
      )}

      <CollapsibleSection
        title="Event simulator"
        description="Test matching and the full backend playback path before shell setup exists."
      >
        <EventSimulator
          form={simulator}
          isPlaying={testingRuleId === 'simulation'}
          isHandlingEvent={isHandlingEvent}
          isSimulating={isSimulating}
          recentEvents={recentEvents}
          result={simulationResult}
          onChange={setSimulator}
          onClearRecent={clearRecentEvents}
          onHandleEvent={handleEventWithPlayback}
          onPlayMatchedSound={playSimulationSound}
          onSimulate={simulateEvent}
        />
      </CollapsibleSection>

      {isFormOpen && (
        <RuleEditorModal
          form={form}
          selectedSound={selectedSound}
          sounds={config.sounds}
          isSaving={savingRuleId === (form.id ?? 'new')}
          isTesting={testingRuleId === 'form'}
          onCancel={closeForm}
          onChange={setForm}
          onSave={saveRule}
          onTest={playSelectedSound}
        />
      )}

      {confirmRule && (
        <ConfirmDialog
          body={`This removes "${confirmRule.name}" from ProdTag. It does not delete any sound files.`}
          confirmLabel="Delete rule"
          isConfirming={deletingRuleId === confirmRule.id}
          onCancel={() => setConfirmRule(null)}
          onConfirm={deleteRule}
          title="Delete rule?"
        />
      )}
    </div>
  );
}

function RuleEditorModal(props: {
  form: RuleFormState;
  sounds: SoundRecord[];
  selectedSound: SoundRecord | null;
  isSaving: boolean;
  isTesting: boolean;
  onCancel: () => void;
  onChange: (form: RuleFormState) => void;
  onSave: () => void;
  onTest: () => void;
}) {
  return (
    <div className="modal-backdrop">
      <section aria-modal="true" className="rule-modal-panel" role="dialog">
        <RuleForm {...props} />
      </section>
    </div>
  );
}

function RuleForm({
  form,
  isSaving,
  isTesting,
  onCancel,
  onChange,
  onSave,
  onTest,
  selectedSound,
  sounds,
}: {
  form: RuleFormState;
  sounds: SoundRecord[];
  selectedSound: SoundRecord | null;
  isSaving: boolean;
  isTesting: boolean;
  onCancel: () => void;
  onChange: (form: RuleFormState) => void;
  onSave: () => void;
  onTest: () => void;
}) {
  return (
    <div>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <h2 className="text-lg font-semibold">{form.id ? 'Edit rule' : 'Create rule'}</h2>
          <p className="mt-1 text-sm text-neutral-500">Choose the event and the sound ProdTag should play later.</p>
        </div>
        <Button leftIcon={<X size={16} />} onClick={onCancel} variant="ghost">
          Cancel
        </Button>
      </div>

      <div className="mt-5 grid gap-4 lg:grid-cols-2">
        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Rule name</span>
          <input
            autoFocus
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            onChange={(event) => onChange({...form, name: event.target.value})}
            placeholder="Tests passed"
            value={form.name}
          />
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Event type</span>
          <select
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            onChange={(event) => onChange({...form, eventType: event.target.value})}
            value={form.eventType}
          >
            {eventTypes.map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Sound</span>
          <select
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            onChange={(event) => onChange({...form, soundId: event.target.value})}
            value={form.soundId}
          >
            <option value="">Choose a sound</option>
            {form.soundId && !selectedSound && (
              <option value={form.soundId}>Missing sound ({form.soundId})</option>
            )}
            {sounds.map((sound) => (
              <option key={sound.id} value={sound.id}>{sound.name}</option>
            ))}
          </select>
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Command filter</span>
          <select
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            onChange={(event) => onChange({...form, matchMode: event.target.value})}
            value={form.matchMode}
          >
            {matchModes.map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Command pattern</span>
          <input
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            disabled={form.matchMode === 'any'}
            onChange={(event) => onChange({...form, commandPattern: event.target.value})}
            placeholder="npm test"
            value={form.commandPattern}
          />
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Exit code override</span>
          <input
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            inputMode="numeric"
            onChange={(event) => onChange({...form, exitCode: event.target.value})}
            placeholder="Optional"
            value={form.exitCode}
          />
        </label>
      </div>

      <div className="mt-4 flex items-center gap-2">
        <input
          checked={form.enabled}
          className="h-4 w-4 accent-neutral-950"
          id="rule-enabled"
          onChange={(event) => onChange({...form, enabled: event.target.checked})}
          type="checkbox"
        />
        <label className="text-sm font-semibold text-neutral-700" htmlFor="rule-enabled">Enable this rule</label>
      </div>

      <SelectedSound missingSoundId={form.soundId && !selectedSound ? form.soundId : ''} sound={selectedSound} />

      <div className="mt-5 flex flex-wrap justify-end gap-2">
        <Button disabled={!selectedSound} isLoading={isTesting} leftIcon={<Play size={16} />} onClick={onTest} variant="success">
          Test sound
        </Button>
        <Button disabled={!selectedSound} isLoading={isSaving} leftIcon={<Save size={16} />} onClick={onSave}>
          {form.id ? 'Save rule' : 'Create rule'}
        </Button>
      </div>
    </div>
  );
}

function EventSimulator({
  form,
  isHandlingEvent,
  isPlaying,
  isSimulating,
  onChange,
  onClearRecent,
  onHandleEvent,
  onPlayMatchedSound,
  onSimulate,
  recentEvents,
  result,
}: {
  form: SimulatorFormState;
  result: RuleMatchResult | null;
  recentEvents: RecentEventRecord[];
  isHandlingEvent: boolean;
  isSimulating: boolean;
  isPlaying: boolean;
  onChange: (form: SimulatorFormState) => void;
  onSimulate: () => void;
  onHandleEvent: () => void;
  onPlayMatchedSound: () => void;
  onClearRecent: () => void;
}) {
  return (
    <div>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 className="text-base font-semibold">Test an event</h2>
          <p className="mt-1 text-sm text-neutral-500">
            Use simulated command data to check what would match, or run the full app playback path.
          </p>
        </div>
        <div className="flex flex-wrap justify-end gap-2">
          <Button isLoading={isSimulating} leftIcon={<SlidersHorizontal size={16} />} onClick={onSimulate} variant="secondary">
            Simulate
          </Button>
          <Button isLoading={isHandlingEvent} leftIcon={<Play size={16} />} onClick={onHandleEvent}>
            Test full event flow
          </Button>
        </div>
      </div>

      <div className="mt-5 grid gap-4 lg:grid-cols-2">
        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Event type</span>
          <select
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            onChange={(event) => onChange({...form, eventType: event.target.value})}
            value={form.eventType}
          >
            {eventTypes.map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Command</span>
          <input
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            onChange={(event) => onChange({...form, command: event.target.value})}
            placeholder="npm test"
            value={form.command}
          />
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Exit code</span>
          <input
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            inputMode="numeric"
            onChange={(event) => onChange({...form, exitCode: event.target.value})}
            placeholder="0"
            value={form.exitCode}
          />
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-semibold text-neutral-700">Duration ms</span>
          <input
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            inputMode="numeric"
            onChange={(event) => onChange({...form, durationMs: event.target.value})}
            placeholder="Optional"
            value={form.durationMs}
          />
        </label>

        <label className="grid gap-2 lg:col-span-2">
          <span className="text-sm font-semibold text-neutral-700">Working directory</span>
          <input
            className="rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-950"
            onChange={(event) => onChange({...form, cwd: event.target.value})}
            placeholder="Optional"
            value={form.cwd}
          />
        </label>
      </div>

      <SimulationResult result={result} isPlaying={isPlaying} onPlayMatchedSound={onPlayMatchedSound} />

      <div className="mt-5 flex items-center justify-between gap-3">
        <h3 className="text-sm font-semibold text-neutral-700">Recent simulated events</h3>
        <Button disabled={recentEvents.length === 0} leftIcon={<RotateCcw size={15} />} onClick={onClearRecent} variant="ghost">
          Clear
        </Button>
      </div>
      {recentEvents.length === 0 ? (
        <p className="mt-3 rounded-lg bg-neutral-50 px-3 py-2 text-sm text-neutral-500">No simulated events yet.</p>
      ) : (
        <div className="mt-3 grid gap-2">
          {recentEvents.slice(0, 5).map((event) => (
            <div key={event.id} className="rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-2">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="min-w-0">
                  <div className="text-sm font-semibold text-neutral-800">{eventLabel(event.event.eventType)}</div>
                  <div className="mt-1 truncate text-xs text-neutral-500">{event.event.command || 'No command'}{event.event.exitCode !== null && event.event.exitCode !== undefined ? ` - exit ${event.event.exitCode}` : ''}</div>
                </div>
                <Badge tone={event.matched ? event.missingSound ? 'rose' : 'green' : 'neutral'}>
                  {event.matched ? event.ruleName || 'Matched' : 'No match'}
                </Badge>
              </div>
              {event.soundName && <div className="mt-1 text-xs text-neutral-500">Sound: {event.soundName}</div>}
              {(event.playbackStarted || event.playbackError) && (
                <div className="mt-1 text-xs text-neutral-500">
                  {event.playbackError ? `Playback error: ${event.playbackError}` : 'Playback started'}
                </div>
              )}
            </div>
          ))}
          {recentEvents.length > 5 && (
            <p className="text-xs text-neutral-500">Showing the 5 most recent simulated events.</p>
          )}
        </div>
      )}
    </div>
  );
}

function SimulationResult({
  isPlaying,
  onPlayMatchedSound,
  result,
}: {
  result: RuleMatchResult | null;
  isPlaying: boolean;
  onPlayMatchedSound: () => void;
}) {
  if (!result) {
    return (
      <div className="mt-4 rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-3 text-sm text-neutral-500">
        Simulate an event to see which rule would fire.
      </div>
    );
  }

  if (!result.matched) {
    return (
      <div className="mt-4 rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-3">
        <div className="flex items-center gap-2 text-sm font-semibold text-neutral-800">
          <XCircle size={16} />
          No matching rule
        </div>
        <p className="mt-1 text-sm text-neutral-500">{result.message}</p>
      </div>
    );
  }

  return (
    <div className={result.missingSound ? 'mt-4 rounded-lg border border-rose-200 bg-rose-50 px-3 py-3' : 'mt-4 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-3'}>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
            {result.missingSound ? <XCircle size={16} /> : <CheckCircle2 size={16} />}
            {result.rule?.name || 'Matched rule'}
          </div>
          <p className="mt-1 text-sm text-neutral-600">
            {result.missingSound ? 'The matched rule points to a missing sound.' : `Sound: ${result.sound?.name || 'Unknown sound'}`}
          </p>
          {(result.playbackStarted || result.playbackError || result.playbackAttempted) && (
            <p className="mt-1 text-sm text-neutral-600">
              {result.playbackError ? `Playback error: ${result.playbackError}` : result.playbackStarted ? 'Backend playback started.' : result.message}
            </p>
          )}
          {result.soundPath && (
            <div className="mt-3">
              <CompactPath label="Matched sound file" path={result.soundPath} tone="green" />
            </div>
          )}
        </div>
        <Button disabled={!result.sound} isLoading={isPlaying} leftIcon={<Play size={15} />} onClick={onPlayMatchedSound} variant="success">
          Play
        </Button>
      </div>
    </div>
  );
}

function RuleCard({
  isBusy,
  isTesting,
  onDelete,
  onEdit,
  onTest,
  onToggle,
  rule,
  sound,
}: {
  rule: RuleRecord;
  sound: SoundRecord | null;
  isBusy: boolean;
  isTesting: boolean;
  onDelete: () => void;
  onEdit: () => void;
  onTest: () => void;
  onToggle: (enabled: boolean) => void;
}) {
  return (
    <Card className="p-4">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="truncate text-base font-semibold">{rule.name}</h3>
            <Badge tone={rule.enabled ? 'green' : 'neutral'}>
              <span className="flex items-center gap-1.5">
                {rule.enabled ? <CheckCircle2 size={14} /> : <XCircle size={14} />}
                {rule.enabled ? 'Enabled' : 'Disabled'}
              </span>
            </Badge>
            <Badge tone={sound ? 'neutral' : 'rose'}>{eventLabel(rule.eventType)}</Badge>
          </div>
          <div className="mt-2 flex flex-wrap items-center gap-2 text-sm text-neutral-500">
            <span>{sound ? sound.name : 'Missing sound'}</span>
            {sound && <span>- {sound.status}</span>}
            {rule.matchMode && <span>- {rule.matchMode}: {rule.commandPattern || 'No pattern yet'}</span>}
            {rule.exitCode !== null && rule.exitCode !== undefined && <span>- exit {rule.exitCode}</span>}
          </div>
          {sound ? (
            <div className="mt-3">
              <CompactPath label="Assigned sound" path={sound.processedPath || sound.originalPath} tone={sound.processedPath ? 'green' : 'neutral'} />
            </div>
          ) : (
            <p className="mt-3 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
              This rule points to a sound that is no longer in the library.
            </p>
          )}
        </div>

        <div className="flex w-full shrink-0 flex-wrap justify-end gap-2 sm:w-[360px]">
          <Button disabled={!sound} isLoading={isTesting} leftIcon={<Play size={15} />} onClick={onTest} variant="success">
            Test
          </Button>
          <Button disabled={isBusy} leftIcon={<Edit3 size={15} />} onClick={onEdit} variant="ghost">
            Edit
          </Button>
          <Button disabled={isBusy} onClick={() => onToggle(!rule.enabled)} variant="secondary">
            {rule.enabled ? 'Disable' : 'Enable'}
          </Button>
          <Button disabled={isBusy} leftIcon={<Trash2 size={15} />} onClick={onDelete} variant="danger">
            Delete
          </Button>
        </div>
      </div>
    </Card>
  );
}

function RuleStat({label, value}: {label: string; value: number}) {
  return (
    <div className="rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-2">
      <div className="text-xs font-semibold uppercase text-neutral-500">{label}</div>
      <div className="mt-1 text-lg font-semibold text-neutral-900">{value}</div>
    </div>
  );
}

function SelectedSound({missingSoundId, sound}: {sound: SoundRecord | null; missingSoundId?: string}) {
  if (!sound) {
    if (missingSoundId) {
      return (
        <div className="mt-4 rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-800">
          The previously assigned sound is missing. Choose another sound before saving.
        </div>
      );
    }

    return (
      <div className="mt-4 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
        Choose a sound before saving this rule.
      </div>
    );
  }

  return (
    <div className="mt-4 rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-neutral-800">{sound.name}</div>
          <div className="mt-1 text-xs text-neutral-500">{sound.format.toUpperCase()} - {sound.status}</div>
        </div>
        <Badge tone={sound.status === 'ready' ? 'green' : sound.status === 'failed' ? 'rose' : 'neutral'}>{sound.status}</Badge>
      </div>
    </div>
  );
}

function validateForm(form: RuleFormState, hasSounds: boolean) {
  if (!hasSounds) {
    return 'Import a sound before creating a rule';
  }
  if (!form.name.trim()) {
    return 'Rule name is required';
  }
  if (!form.eventType) {
    return 'Event type is required';
  }
  if (!form.soundId) {
    return 'Choose a sound for this rule';
  }
  if (form.exitCode.trim() !== '' && Number.isNaN(Number(form.exitCode))) {
    return 'Exit code must be a number';
  }
  return '';
}

function simulatorEvent(form: SimulatorFormState) {
  const exitCode = form.exitCode.trim() === '' ? undefined : Number(form.exitCode);
  const durationMs = form.durationMs.trim() === '' ? undefined : Number(form.durationMs);
  if (exitCode !== undefined && Number.isNaN(exitCode)) {
    return 'Exit code must be a number';
  }
  if (durationMs !== undefined && Number.isNaN(durationMs)) {
    return 'Duration must be a number';
  }

  return {
    eventType: form.eventType,
    command: form.command.trim(),
    exitCode,
    cwd: form.cwd.trim(),
    timestamp: new Date().toISOString(),
    durationMs,
  };
}

function eventLabel(value: string) {
  return eventTypes.find(([eventType]) => eventType === value)?.[1] ?? value;
}

function failureMessage(error: unknown) {
  if (error instanceof Error) {
    return `Failed: ${error.message}`;
  }
  if (typeof error === 'string') {
    return `Failed: ${error}`;
  }
  return 'Failed: something went wrong';
}
