import {useEffect, useRef, useState} from 'react';
import {DeleteSound, GetSoundPreviewDataURL, ImportSoundPaths, RenameSound, SelectSoundFiles} from '../../wailsjs/go/main/App';
import {OnFileDrop, OnFileDropOff} from '../../wailsjs/runtime/runtime';
import {Badge} from '../components/Badge';
import {Button} from '../components/Button';
import {Card} from '../components/Card';
import {EmptyState} from '../components/EmptyState';
import {AppConfig, ConfigSnapshot, SoundRecord} from '../types/app';

type ImportProgressTone = 'neutral' | 'green' | 'amber' | 'rose';

type ImportProgress = {
  message: string;
  tone: ImportProgressTone;
};

type SoundsPageProps = {
  config: AppConfig;
  onConfigUpdated: (snapshot: ConfigSnapshot) => void;
};

export function SoundsPage({config, onConfigUpdated}: SoundsPageProps) {
  const [progress, setProgress] = useState<ImportProgress | null>(null);
  const [isImporting, setIsImporting] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');
  const [playingId, setPlayingId] = useState<string | null>(null);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  useEffect(() => {
    OnFileDrop((_x, _y, paths) => {
      if (paths.length > 0) {
        void importPaths(paths);
      }
    }, false);

    return () => {
      OnFileDropOff();
    };
  }, []);

  async function importFromPicker() {
    setIsImporting(true);
    setProgress({message: 'Selecting file...', tone: 'neutral'});
    try {
      const paths = await SelectSoundFiles();
      if (paths.length === 0) {
        setProgress(null);
        return;
      }
      setProgress({message: 'Copying to library...', tone: 'amber'});
      const snapshot = await ImportSoundPaths(paths);
      setProgress({message: 'Reading metadata...', tone: 'amber'});
      onConfigUpdated(snapshot as ConfigSnapshot);
      setProgress({message: 'Added to library', tone: 'green'});
    } catch (error) {
      setProgress({message: failureMessage(error), tone: 'rose'});
    } finally {
      setIsImporting(false);
    }
  }

  async function importPaths(paths: string[]) {
    setIsImporting(true);
    setProgress({message: 'Copying to library...', tone: 'amber'});
    try {
      const snapshot = await ImportSoundPaths(paths);
      setProgress({message: 'Reading metadata...', tone: 'amber'});
      onConfigUpdated(snapshot as ConfigSnapshot);
      setProgress({message: 'Added to library', tone: 'green'});
    } catch (error) {
      setProgress({message: failureMessage(error), tone: 'rose'});
    } finally {
      setIsImporting(false);
    }
  }

  function startRename(sound: SoundRecord) {
    setEditingId(sound.id);
    setEditingName(sound.name);
  }

  async function saveRename(sound: SoundRecord) {
    const name = editingName.trim();
    if (!name) {
      setProgress({message: 'Sound name cannot be empty', tone: 'rose'});
      return;
    }

    try {
      const snapshot = await RenameSound({id: sound.id, name});
      onConfigUpdated(snapshot as ConfigSnapshot);
      setEditingId(null);
      setEditingName('');
      setProgress({message: 'Sound renamed', tone: 'green'});
    } catch (error) {
      setProgress({message: failureMessage(error), tone: 'rose'});
    }
  }

  async function deleteSound(sound: SoundRecord) {
    if (!window.confirm(`Delete "${sound.name}" from the library?`)) {
      return;
    }

    try {
      stopPreview();
      const snapshot = await DeleteSound(sound.id);
      onConfigUpdated(snapshot as ConfigSnapshot);
      setProgress({message: 'Sound deleted', tone: 'green'});
    } catch (error) {
      setProgress({message: failureMessage(error), tone: 'rose'});
    }
  }

  async function playPreview(sound: SoundRecord) {
    try {
      stopPreview();
      const dataURL = await GetSoundPreviewDataURL(sound.id);
      const audio = new Audio(dataURL);
      audioRef.current = audio;
      setPlayingId(sound.id);
      audio.addEventListener('ended', () => setPlayingId(null), {once: true});
      await audio.play();
    } catch (error) {
      setPlayingId(null);
      setProgress({message: failureMessage(error), tone: 'rose'});
    }
  }

  function stopPreview() {
    if (audioRef.current) {
      audioRef.current.pause();
      audioRef.current.currentTime = 0;
      audioRef.current = null;
    }
    setPlayingId(null);
  }

  return (
    <div className="space-y-5">
      <Card>
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Sound library</h2>
            <p className="mt-1 text-sm text-neutral-500">
              Import short producer tags and keep local copies inside ProdTag.
            </p>
          </div>
          <Button disabled={isImporting} onClick={importFromPicker} variant="secondary">
            {isImporting ? 'Importing...' : 'Import sound'}
          </Button>
        </div>

        <div className="drop-zone mt-5">
          <div className="text-sm font-semibold text-neutral-800">Drag audio files here</div>
          <p className="mt-1 text-sm text-neutral-500">MP3, WAV, M4A, OGG, and FLAC are accepted.</p>
        </div>

        {progress && (
          <div className="mt-4 flex items-center justify-between gap-3 rounded-lg bg-neutral-50 px-3 py-2">
            <span className="text-sm text-neutral-700">{progress.message}</span>
            <Badge tone={progress.tone}>{progress.tone === 'rose' ? 'Failed' : 'Import'}</Badge>
          </div>
        )}
      </Card>

      {config.sounds.length === 0 ? (
        <Card>
          <EmptyState
            title="No sounds yet"
            body="Use the import button or drag audio files onto the Sounds page to add your first producer tag."
          />
        </Card>
      ) : (
        <section className="grid gap-4">
          {config.sounds.map((sound) => (
            <SoundCard
              editingId={editingId}
              editingName={editingName}
              key={sound.id}
              playingId={playingId}
              sound={sound}
              onCancelRename={() => setEditingId(null)}
              onDelete={() => deleteSound(sound)}
              onEditNameChange={setEditingName}
              onPlay={() => playPreview(sound)}
              onRename={() => startRename(sound)}
              onSaveRename={() => saveRename(sound)}
              onStop={stopPreview}
            />
          ))}
        </section>
      )}
    </div>
  );
}

type SoundCardProps = {
  sound: SoundRecord;
  playingId: string | null;
  editingId: string | null;
  editingName: string;
  onPlay: () => void;
  onStop: () => void;
  onRename: () => void;
  onSaveRename: () => void;
  onCancelRename: () => void;
  onEditNameChange: (value: string) => void;
  onDelete: () => void;
};

function SoundCard({
  sound,
  playingId,
  editingId,
  editingName,
  onPlay,
  onStop,
  onRename,
  onSaveRename,
  onCancelRename,
  onEditNameChange,
  onDelete,
}: SoundCardProps) {
  const isEditing = editingId === sound.id;
  const isPlaying = playingId === sound.id;

  return (
    <Card className="p-4">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          {isEditing ? (
            <input
              autoFocus
              className="w-full rounded-lg border border-neutral-300 bg-white px-3 py-2 text-sm font-semibold outline-none focus:border-neutral-950"
              onChange={(event) => onEditNameChange(event.target.value)}
              value={editingName}
            />
          ) : (
            <h3 className="truncate text-base font-semibold">{sound.name}</h3>
          )}
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <Badge tone={statusTone(sound.status)}>{sound.status}</Badge>
            <span className="rounded-full bg-neutral-100 px-3 py-1 text-sm font-semibold uppercase text-neutral-600">
              {sound.format || fileExtension(sound.originalPath)}
            </span>
            <span className="text-sm text-neutral-500">Imported {formatDate(sound.createdAt)}</span>
          </div>
          <p className="mt-3 break-all font-mono text-xs leading-5 text-neutral-500">{sound.originalPath}</p>
          {sound.error && <p className="mt-2 text-sm text-rose-700">{sound.error}</p>}
        </div>

        <div className="flex flex-wrap justify-end gap-2">
          {isPlaying ? (
            <Button onClick={onStop} variant="secondary">
              Stop
            </Button>
          ) : (
            <Button onClick={onPlay} variant="secondary">
              Preview
            </Button>
          )}
          {isEditing ? (
            <>
              <Button onClick={onSaveRename}>Save</Button>
              <Button onClick={onCancelRename} variant="ghost">
                Cancel
              </Button>
            </>
          ) : (
            <Button onClick={onRename} variant="ghost">
              Rename
            </Button>
          )}
          <Button onClick={onDelete} variant="ghost">
            Delete
          </Button>
        </div>
      </div>
    </Card>
  );
}

function statusTone(status: SoundRecord['status']) {
  if (status === 'ready') {
    return 'green';
  }
  if (status === 'processing') {
    return 'amber';
  }
  if (status === 'failed') {
    return 'rose';
  }
  return 'neutral';
}

function fileExtension(path: string) {
  const extension = path.split('.').pop();
  return extension ? extension.toUpperCase() : 'AUDIO';
}

function formatDate(value: string) {
  if (!value) {
    return 'just now';
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString();
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
