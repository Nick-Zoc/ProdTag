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

export type PlaylistRecord = {
  id: string;
  name: string;
  soundIds: string[];
};

export type RuleRecord = {
  id: string;
  name: string;
  enabled: boolean;
  pattern: string;
  matchType: string;
  condition: string;
  soundId?: string;
  playlistId?: string;
  cooldownMs?: number;
  probability?: number;
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
