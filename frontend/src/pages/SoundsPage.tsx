import {useEffect, useMemo, useRef, useState} from 'react';
import {
  DeleteSound,
  DeleteSounds,
  GetSoundPreviewDataURL,
  ImportSoundPaths,
  RenameSound,
  SelectSoundFiles,
} from '../../wailsjs/go/main/App';
import {OnFileDrop, OnFileDropOff} from '../../wailsjs/runtime/runtime';
import {Badge} from '../components/Badge';
import {Button} from '../components/Button';
import {Card} from '../components/Card';
import {ConfirmDialog} from '../components/ConfirmDialog';
import {EmptyState} from '../components/EmptyState';
import {Toast, ToastState} from '../components/Toast';
import {AppConfig, ConfigSnapshot, SoundRecord} from '../types/app';
import {classNames} from '../utils/classNames';

type SoundsPageProps = {
  config: AppConfig;
  onConfigUpdated: (snapshot: ConfigSnapshot) => void;
};

type ConfirmState =
  | {kind: 'single'; sound: SoundRecord}
  | {kind: 'bulk'; ids: string[]}
  | null;

export function SoundsPage({config, onConfigUpdated}: SoundsPageProps) {
  const [toast, setToast] = useState<ToastState | null>(null);
  const [isImporting, setIsImporting] = useState(false);
  const [isDropActive, setIsDropActive] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');
  const [playingId, setPlayingId] = useState<string | null>(null);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [confirmState, setConfirmState] = useState<ConfirmState>(null);
  const dragDepth = useRef(0);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  const selectedCount = selectedIds.length;
  const allVisibleSelected = useMemo(() => {
    return config.sounds.length > 0 && selectedIds.length === config.sounds.length;
  }, [config.sounds.length, selectedIds.length]);

  useEffect(() => {
    OnFileDrop((_x, _y, paths) => {
      dragDepth.current = 0;
      setIsDropActive(false);
      if (paths.length > 0) {
        void importPaths(paths);
      }
    }, true);

    return () => {
      OnFileDropOff();
    };
  }, []);

  useEffect(() => {
    const soundIds = new Set(config.sounds.map((sound) => sound.id));
    setSelectedIds((current) => current.filter((id) => soundIds.has(id)));
  }, [config.sounds]);

  async function importFromPicker() {
    setIsImporting(true);
    setToast({message: 'Selecting file...', tone: 'neutral'});
    try {
      const paths = await SelectSoundFiles();
      if (paths.length === 0) {
        setToast(null);
        return;
      }
      await importPaths(paths);
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    } finally {
      setIsImporting(false);
    }
  }

  async function importPaths(paths: string[]) {
    setIsImporting(true);
    setToast({message: 'Copying to library...', tone: 'amber'});
    try {
      const snapshot = await ImportSoundPaths(paths);
      setToast({message: 'Reading metadata...', tone: 'amber'});
      onConfigUpdated(snapshot as ConfigSnapshot);
      setToast({message: `${paths.length} sound${paths.length === 1 ? '' : 's'} added to library`, tone: 'green'});
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    } finally {
      setIsImporting(false);
    }
  }

  function handleDragEnter(event: React.DragEvent<HTMLDivElement>) {
    event.preventDefault();
    dragDepth.current += 1;
    setIsDropActive(true);
  }

  function handleDragOver(event: React.DragEvent<HTMLDivElement>) {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
    setIsDropActive(true);
  }

  function handleDragLeave(event: React.DragEvent<HTMLDivElement>) {
    event.preventDefault();
    dragDepth.current = Math.max(0, dragDepth.current - 1);
    if (dragDepth.current === 0) {
      setIsDropActive(false);
    }
  }

  function handleDrop(event: React.DragEvent<HTMLDivElement>) {
    event.preventDefault();
    dragDepth.current = 0;
    setIsDropActive(false);
    setToast({message: 'Drop received. Preparing import...', tone: 'neutral'});
  }

  function startRename(sound: SoundRecord) {
    setEditingId(sound.id);
    setEditingName(sound.name);
  }

  async function saveRename(sound: SoundRecord) {
    const name = editingName.trim();
    if (!name) {
      setToast({message: 'Sound name cannot be empty', tone: 'rose'});
      return;
    }

    try {
      const snapshot = await RenameSound({id: sound.id, name});
      onConfigUpdated(snapshot as ConfigSnapshot);
      setEditingId(null);
      setEditingName('');
      setToast({message: 'Sound renamed', tone: 'green'});
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    }
  }

  function requestDelete(sound: SoundRecord) {
    setConfirmState({kind: 'single', sound});
  }

  function requestBulkDelete() {
    if (selectedIds.length > 0) {
      setConfirmState({kind: 'bulk', ids: selectedIds});
    }
  }

  async function confirmDelete() {
    if (!confirmState) {
      return;
    }

    try {
      if (confirmState.kind === 'single') {
        if (playingId === confirmState.sound.id) {
          stopPreview();
        }
        const snapshot = await DeleteSound(confirmState.sound.id);
        onConfigUpdated(snapshot as ConfigSnapshot);
        setSelectedIds((current) => current.filter((id) => id !== confirmState.sound.id));
        setToast({message: 'Sound deleted', tone: 'green'});
      } else {
        if (playingId && confirmState.ids.includes(playingId)) {
          stopPreview();
        }
        const snapshot = await DeleteSounds(confirmState.ids);
        onConfigUpdated(snapshot as ConfigSnapshot);
        setSelectedIds([]);
        setToast({message: `${confirmState.ids.length} sound${confirmState.ids.length === 1 ? '' : 's'} deleted`, tone: 'green'});
      }
    } catch (error) {
      setToast({message: failureMessage(error), tone: 'rose'});
    } finally {
      setConfirmState(null);
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
      setToast({message: failureMessage(error), tone: 'rose'});
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

  function toggleSelected(id: string, checked: boolean) {
    setSelectedIds((current) => {
      if (checked) {
        return current.includes(id) ? current : [...current, id];
      }
      return current.filter((currentId) => currentId !== id);
    });
  }

  function toggleAllSelected(checked: boolean) {
    setSelectedIds(checked ? config.sounds.map((sound) => sound.id) : []);
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

        <div
          className={classNames('drop-zone mt-5', isDropActive && 'drop-zone-active')}
          onDragEnter={handleDragEnter}
          onDragLeave={handleDragLeave}
          onDragOver={handleDragOver}
          onDrop={handleDrop}
        >
          <div className="text-sm font-semibold text-neutral-800">
            {isDropActive ? 'Drop to copy into ProdTag' : 'Drag audio files here'}
          </div>
          <p className="mt-1 text-sm text-neutral-500">MP3, WAV, M4A, OGG, and FLAC are accepted.</p>
        </div>

        {toast && <Toast toast={toast} />}
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
          <div className="flex flex-wrap items-center justify-between gap-3">
            <label className="flex items-center gap-2 text-sm font-medium text-neutral-600">
              <input
                checked={allVisibleSelected}
                className="h-4 w-4 accent-neutral-950"
                onChange={(event) => toggleAllSelected(event.target.checked)}
                type="checkbox"
              />
              Select all
            </label>
            {selectedCount > 0 && (
              <Button onClick={requestBulkDelete} variant="danger">
                Delete selected ({selectedCount})
              </Button>
            )}
          </div>

          {config.sounds.map((sound) => (
            <SoundCard
              editingId={editingId}
              editingName={editingName}
              isSelected={selectedIds.includes(sound.id)}
              key={sound.id}
              playingId={playingId}
              sound={sound}
              onCancelRename={() => setEditingId(null)}
              onDelete={() => requestDelete(sound)}
              onEditNameChange={setEditingName}
              onPlay={() => playPreview(sound)}
              onRename={() => startRename(sound)}
              onSaveRename={() => saveRename(sound)}
              onSelectedChange={(checked) => toggleSelected(sound.id, checked)}
              onStop={stopPreview}
            />
          ))}
        </section>
      )}

      {confirmState && (
        <ConfirmDialog
          body={confirmationBody(confirmState)}
          confirmLabel={confirmState.kind === 'bulk' ? 'Delete selected' : 'Delete sound'}
          onCancel={() => setConfirmState(null)}
          onConfirm={confirmDelete}
          title="Delete from library?"
        />
      )}
    </div>
  );
}

type SoundCardProps = {
  sound: SoundRecord;
  playingId: string | null;
  editingId: string | null;
  editingName: string;
  isSelected: boolean;
  onSelectedChange: (checked: boolean) => void;
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
  isSelected,
  onSelectedChange,
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
        <div className="flex min-w-0 flex-1 gap-3">
          <input
            checked={isSelected}
            className="mt-1 h-4 w-4 shrink-0 accent-neutral-950"
            onChange={(event) => onSelectedChange(event.target.checked)}
            type="checkbox"
          />
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
        </div>

        <div className="flex flex-wrap justify-end gap-2">
          {isPlaying ? (
            <Button onClick={onStop} variant="secondary">
              Stop
            </Button>
          ) : (
            <Button onClick={onPlay} variant="success">
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
          <Button onClick={onDelete} variant="danger">
            Delete
          </Button>
        </div>
      </div>
    </Card>
  );
}

function confirmationBody(confirmState: ConfirmState) {
  if (!confirmState) {
    return '';
  }
  if (confirmState.kind === 'single') {
    return `This removes "${confirmState.sound.name}" from ProdTag and deletes its copied library file. Your original selected file is not touched.`;
  }
  return `This removes ${confirmState.ids.length} selected sound${confirmState.ids.length === 1 ? '' : 's'} from ProdTag and deletes their copied library files. Original selected files are not touched.`;
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
