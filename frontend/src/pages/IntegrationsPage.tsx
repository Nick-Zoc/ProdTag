import {useEffect, useState} from 'react';
import type {ReactNode} from 'react';
import {Activity, CheckCircle2, Plug, Square, XCircle} from 'lucide-react';
import {GetPlaybackStatus, ListRecentEvents, StopPlayback} from '../../wailsjs/go/main/App';
import {Badge} from '../components/Badge';
import {Button} from '../components/Button';
import {Card} from '../components/Card';
import {AppConfig, PlaybackStatus, RecentEventRecord} from '../types/app';

export function IntegrationsPage({config}: {config: AppConfig}) {
  const [playbackStatus, setPlaybackStatus] = useState<PlaybackStatus | null>(null);
  const [recentEvents, setRecentEvents] = useState<RecentEventRecord[]>([]);
  const [isStopping, setIsStopping] = useState(false);

  const integrations = [
    ['zsh', config.integrations.zsh],
    ['bash', config.integrations.bash],
    ['PowerShell', config.integrations.powerShell],
  ] as const;

  useEffect(() => {
    void refreshStatus();
  }, []);

  async function refreshStatus() {
    const [status, events] = await Promise.all([GetPlaybackStatus(), ListRecentEvents()]);
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

  return (
    <div className="space-y-6">
      <Card>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Event engine</h2>
            <p className="mt-1 text-sm text-neutral-500">
              Local event handling is ready for future shell integrations. No shell files are installed yet.
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
            label="Local intake"
            tone="neutral"
            value="Deferred to Phase 4.1"
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

      <Card>
        <h2 className="text-lg font-semibold">Recent handled events</h2>
        <div className="mt-4 grid gap-2">
          {recentEvents.length === 0 ? (
            <p className="rounded-lg bg-neutral-50 px-3 py-2 text-sm text-neutral-500">
              No events handled yet. Use Handle + play in the Rules simulator to exercise the backend path.
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
      </Card>

      <Card>
        <h2 className="text-lg font-semibold">Shell integrations</h2>
        <div className="mt-5 grid gap-3 md:grid-cols-3">
          {integrations.map(([name, integration]) => (
            <div key={name} className="rounded-lg border border-neutral-200 p-4">
              <div className="flex items-center justify-between gap-3">
                <div className="text-base font-semibold">{name}</div>
                <Badge tone={integration.installed ? 'green' : 'neutral'}>
                  {integration.installed ? 'Installed' : 'Not installed'}
                </Badge>
              </div>
              <div className="mt-3 text-sm text-neutral-500">
                {integration.scope || 'Install and doctor checks will appear here.'}
              </div>
            </div>
          ))}
        </div>
      </Card>
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
