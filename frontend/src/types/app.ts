export type PageKey = 'dashboard' | 'sounds' | 'rules' | 'hotkeys' | 'integrations' | 'settings';

export type SoundRecord = {
  id: string;
  name: string;
  originalPath: string;
  processedPath?: string | null;
  durationMs?: number | null;
  format: string;
  createdAt: string;
  status: string;
  error?: string | null;
};

export type AudioToolsStatus = {
  ffmpegAvailable: boolean;
  ffprobeAvailable: boolean;
  ffmpegPath: string;
  ffprobePath: string;
  message: string;
  error?: string | null;
};

export type PlaylistRecord = {
  id: string;
  name: string;
  soundIds: string[];
};

export type RuleRecord = {
  id: string;
  name: string;
  enabled: boolean;
  eventType: string;
  soundId: string;
  matchMode?: string;
  commandPattern?: string;
  exitCode?: number | null;
  createdAt: string;
  updatedAt: string;
  playlistId?: string;
  cooldownMs?: number;
  probability?: number;
};

export type TerminalEvent = {
  eventType: string;
  command?: string;
  exitCode?: number | null;
  cwd?: string;
  timestamp: string;
  durationMs?: number | null;
};

export type RuleMatchResult = {
  matched: boolean;
  rule?: RuleRecord | null;
  sound?: SoundRecord | null;
  soundPath?: string;
  missingSound: boolean;
  playbackAttempted: boolean;
  playbackStarted: boolean;
  playbackError?: string;
  eventEngineEnabled: boolean;
  playbackEnabled: boolean;
  message: string;
  event: TerminalEvent;
};

export type RecentEventRecord = {
  id: string;
  event: TerminalEvent;
  matched: boolean;
  ruleId?: string;
  ruleName?: string;
  soundId?: string;
  soundName?: string;
  missingSound: boolean;
  playbackStarted: boolean;
  playbackError?: string;
  message: string;
  timestamp: string;
};

export type PlaybackStatus = {
  supported: boolean;
  platform: string;
  method: string;
  playing: boolean;
  message: string;
};

export type ShellIntegrationState = {
  installed: boolean;
  scope: string;
  lastCheck: string;
};

export type AppConfig = {
  version: number;
  listening: boolean;
  muted: boolean;
  eventEngineEnabled: boolean;
  playbackEnabled: boolean;
  stopPreviousSoundOnNewEvent: boolean;
  localEventPort?: number;
  launchHelperAtStartup: boolean;
  sounds: SoundRecord[];
  playlists: PlaylistRecord[];
  rules: RuleRecord[];
  hotkeys: {
    stopAudio: string;
    toggleListening: string;
    toggleMute: string;
    openApp: string;
  };
  integrations: {
    zsh: ShellIntegrationState;
    bash: ShellIntegrationState;
    powerShell: ShellIntegrationState;
  };
  updatedAt: string;
};

export type AppDataPaths = {
  configDir: string;
  dataDir: string;
  configFile: string;
  matcherCacheFile: string;
  originalSoundsDir: string;
  processedSoundsDir: string;
  logsDir: string;
};

export type ConfigSnapshot = {
  config: AppConfig;
  paths: AppDataPaths;
};

export type LoadState = 'loading' | 'ready' | 'saving' | 'error';

export type PageNavItem = {
  key: PageKey;
  label: string;
  hint: string;
};

export const pages: PageNavItem[] = [
  {key: 'dashboard', label: 'Dashboard', hint: 'Status'},
  {key: 'sounds', label: 'Sounds', hint: 'Library'},
  {key: 'rules', label: 'Rules', hint: 'Events'},
  {key: 'hotkeys', label: 'Hotkeys', hint: 'Shortcuts'},
  {key: 'integrations', label: 'Integrations', hint: 'Shells'},
  {key: 'settings', label: 'Settings', hint: 'Config'},
];
