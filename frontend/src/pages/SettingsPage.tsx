import {Badge} from '../components/Badge';
import {Card} from '../components/Card';
import {PathList} from '../components/PathList';
import {ToggleCard} from '../components/ToggleCard';
import {AppConfig, AppDataPaths} from '../types/app';

type SettingsPageProps = {
  config: AppConfig;
  paths: AppDataPaths | null;
  updateDraft: (patch: Partial<AppConfig>) => void;
};

export function SettingsPage({config, paths, updateDraft}: SettingsPageProps) {
  return (
    <div className="space-y-6">
      <Card>
        <h2 className="text-lg font-semibold">Runtime state</h2>
        <p className="mt-1 text-sm text-neutral-500">
          Listening controls whether terminal events are accepted. Muted controls only audio playback.
        </p>
        <div className="mt-5 grid gap-4 md:grid-cols-3">
          <ToggleCard
            label="Listening"
            description="Off means ProdTag ignores terminal events entirely."
            checked={config.listening}
            onChange={(listening) => updateDraft({listening})}
          />
          <ToggleCard
            label="Event engine"
            description="Off means simulated or local events are logged but not matched."
            checked={config.eventEngineEnabled}
            onChange={(eventEngineEnabled) => updateDraft({eventEngineEnabled})}
          />
          <ToggleCard
            label="Muted"
            description="On means events can still be received and logged, but no sound plays."
            checked={config.muted}
            onChange={(muted) => updateDraft({muted})}
          />
          <ToggleCard
            label="Backend playback"
            description="Off means matching events will not start local audio playback."
            checked={config.playbackEnabled}
            onChange={(playbackEnabled) => updateDraft({playbackEnabled})}
          />
          <ToggleCard
            label="Stop previous sound"
            description="On means a new event stops current backend playback before starting another sound."
            checked={config.stopPreviousSoundOnNewEvent}
            onChange={(stopPreviousSoundOnNewEvent) => updateDraft({stopPreviousSoundOnNewEvent})}
          />
          <ToggleCard
            label="Start helper at login"
            description="Reserved for the background helper startup setting."
            checked={config.launchHelperAtStartup}
            onChange={(launchHelperAtStartup) => updateDraft({launchHelperAtStartup})}
          />
        </div>
      </Card>

      <Card>
        <div className="flex items-center justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Config file</h2>
            <p className="mt-1 text-sm text-neutral-500">Last saved: {formatDate(config.updatedAt)}</p>
          </div>
          <Badge>v{config.version}</Badge>
        </div>
        {paths && <PathList paths={paths} />}
      </Card>
    </div>
  );
}

function formatDate(value: string) {
  if (!value) {
    return 'Not saved yet';
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString();
}
