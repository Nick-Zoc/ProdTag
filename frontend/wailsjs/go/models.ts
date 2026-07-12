export namespace core {

	export class ShellIntegrationState {
	    installed: boolean;
	    scope: string;
	    lastCheck: string;

	    static createFrom(source: any = {}) {
	        return new ShellIntegrationState(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.scope = source["scope"];
	        this.lastCheck = source["lastCheck"];
	    }
	}
	export class IntegrationSettings {
	    zsh: ShellIntegrationState;
	    bash: ShellIntegrationState;
	    powerShell: ShellIntegrationState;

	    static createFrom(source: any = {}) {
	        return new IntegrationSettings(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.zsh = this.convertValues(source["zsh"], ShellIntegrationState);
	        this.bash = this.convertValues(source["bash"], ShellIntegrationState);
	        this.powerShell = this.convertValues(source["powerShell"], ShellIntegrationState);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HotkeySettings {
	    stopAudio: string;
	    toggleListening: string;
	    toggleMute: string;
	    openApp: string;

	    static createFrom(source: any = {}) {
	        return new HotkeySettings(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stopAudio = source["stopAudio"];
	        this.toggleListening = source["toggleListening"];
	        this.toggleMute = source["toggleMute"];
	        this.openApp = source["openApp"];
	    }
	}
	export class RuleRecord {
	    id: string;
	    name: string;
	    enabled: boolean;
	    eventType: string;
	    soundId: string;
	    matchMode?: string;
	    commandPattern?: string;
	    exitCode?: number;
	    createdAt: string;
	    updatedAt: string;
	    playlistId?: string;
	    cooldownMs?: number;
	    probability?: number;

	    static createFrom(source: any = {}) {
	        return new RuleRecord(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	        this.eventType = source["eventType"];
	        this.soundId = source["soundId"];
	        this.matchMode = source["matchMode"];
	        this.commandPattern = source["commandPattern"];
	        this.exitCode = source["exitCode"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.playlistId = source["playlistId"];
	        this.cooldownMs = source["cooldownMs"];
	        this.probability = source["probability"];
	    }
	}
	export class PlaylistRecord {
	    id: string;
	    name: string;
	    soundIds: string[];

	    static createFrom(source: any = {}) {
	        return new PlaylistRecord(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.soundIds = source["soundIds"];
	    }
	}
	export class SoundRecord {
	    id: string;
	    name: string;
	    originalPath: string;
	    processedPath?: string;
	    durationMs?: number;
	    format: string;
	    createdAt: string;
	    status: string;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new SoundRecord(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.originalPath = source["originalPath"];
	        this.processedPath = source["processedPath"];
	        this.durationMs = source["durationMs"];
	        this.format = source["format"];
	        this.createdAt = source["createdAt"];
	        this.status = source["status"];
	        this.error = source["error"];
	    }
	}
	export class AppConfig {
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
	    hotkeys: HotkeySettings;
	    integrations: IntegrationSettings;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.listening = source["listening"];
	        this.muted = source["muted"];
	        this.eventEngineEnabled = source["eventEngineEnabled"];
	        this.playbackEnabled = source["playbackEnabled"];
	        this.stopPreviousSoundOnNewEvent = source["stopPreviousSoundOnNewEvent"];
	        this.localEventPort = source["localEventPort"];
	        this.launchHelperAtStartup = source["launchHelperAtStartup"];
	        this.sounds = this.convertValues(source["sounds"], SoundRecord);
	        this.playlists = this.convertValues(source["playlists"], PlaylistRecord);
	        this.rules = this.convertValues(source["rules"], RuleRecord);
	        this.hotkeys = this.convertValues(source["hotkeys"], HotkeySettings);
	        this.integrations = this.convertValues(source["integrations"], IntegrationSettings);
	        this.updatedAt = source["updatedAt"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppDataPaths {
	    configDir: string;
	    dataDir: string;
	    configFile: string;
	    matcherCacheFile: string;
	    originalSoundsDir: string;
	    processedSoundsDir: string;
	    logsDir: string;
	    binDir: string;
	    integrationsDir: string;
	    helperBinary: string;
	    zshScript: string;
	    bashScript: string;
	    powerShellScript: string;

	    static createFrom(source: any = {}) {
	        return new AppDataPaths(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.configDir = source["configDir"];
	        this.dataDir = source["dataDir"];
	        this.configFile = source["configFile"];
	        this.matcherCacheFile = source["matcherCacheFile"];
	        this.originalSoundsDir = source["originalSoundsDir"];
	        this.processedSoundsDir = source["processedSoundsDir"];
	        this.logsDir = source["logsDir"];
	        this.binDir = source["binDir"];
	        this.integrationsDir = source["integrationsDir"];
	        this.helperBinary = source["helperBinary"];
	        this.zshScript = source["zshScript"];
	        this.bashScript = source["bashScript"];
	        this.powerShellScript = source["powerShellScript"];
	    }
	}
	export class ConfigSnapshot {
	    config: AppConfig;
	    paths: AppDataPaths;

	    static createFrom(source: any = {}) {
	        return new ConfigSnapshot(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = this.convertValues(source["config"], AppConfig);
	        this.paths = this.convertValues(source["paths"], AppDataPaths);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DependencySuggestion {
	    platform: string;
	    label: string;
	    command?: string;
	    detail: string;

	    static createFrom(source: any = {}) {
	        return new DependencySuggestion(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.label = source["label"];
	        this.command = source["command"];
	        this.detail = source["detail"];
	    }
	}


	export class PlaybackStatus {
	    supported: boolean;
	    platform: string;
	    method: string;
	    playing: boolean;
	    message: string;
	    alternatives: string[];
	    suggestions: DependencySuggestion[];

	    static createFrom(source: any = {}) {
	        return new PlaybackStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supported = source["supported"];
	        this.platform = source["platform"];
	        this.method = source["method"];
	        this.playing = source["playing"];
	        this.message = source["message"];
	        this.alternatives = source["alternatives"];
	        this.suggestions = this.convertValues(source["suggestions"], DependencySuggestion);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class TerminalEvent {
	    eventType: string;
	    command?: string;
	    exitCode?: number;
	    cwd?: string;
	    timestamp: string;
	    durationMs?: number;

	    static createFrom(source: any = {}) {
	        return new TerminalEvent(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.eventType = source["eventType"];
	        this.command = source["command"];
	        this.exitCode = source["exitCode"];
	        this.cwd = source["cwd"];
	        this.timestamp = source["timestamp"];
	        this.durationMs = source["durationMs"];
	    }
	}
	export class RecentEventRecord {
	    id: string;
	    event: TerminalEvent;
	    matched: boolean;
	    ruleId?: string;
	    ruleName?: string;
	    soundId?: string;
	    soundName?: string;
	    missingSound: boolean;
	    playbackAttempted: boolean;
	    playbackStarted: boolean;
	    playbackSkipped: boolean;
	    playbackSkipReason?: string;
	    playbackError?: string;
	    message: string;
	    timestamp: string;

	    static createFrom(source: any = {}) {
	        return new RecentEventRecord(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.event = this.convertValues(source["event"], TerminalEvent);
	        this.matched = source["matched"];
	        this.ruleId = source["ruleId"];
	        this.ruleName = source["ruleName"];
	        this.soundId = source["soundId"];
	        this.soundName = source["soundName"];
	        this.missingSound = source["missingSound"];
	        this.playbackAttempted = source["playbackAttempted"];
	        this.playbackStarted = source["playbackStarted"];
	        this.playbackSkipped = source["playbackSkipped"];
	        this.playbackSkipReason = source["playbackSkipReason"];
	        this.playbackError = source["playbackError"];
	        this.message = source["message"];
	        this.timestamp = source["timestamp"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RuleMatchResult {
	    matched: boolean;
	    rule?: RuleRecord;
	    sound?: SoundRecord;
	    soundPath?: string;
	    missingSound: boolean;
	    playbackAttempted: boolean;
	    playbackStarted: boolean;
	    playbackError?: string;
	    eventEngineEnabled: boolean;
	    playbackEnabled: boolean;
	    message: string;
	    event: TerminalEvent;

	    static createFrom(source: any = {}) {
	        return new RuleMatchResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.matched = source["matched"];
	        this.rule = this.convertValues(source["rule"], RuleRecord);
	        this.sound = this.convertValues(source["sound"], SoundRecord);
	        this.soundPath = source["soundPath"];
	        this.missingSound = source["missingSound"];
	        this.playbackAttempted = source["playbackAttempted"];
	        this.playbackStarted = source["playbackStarted"];
	        this.playbackError = source["playbackError"];
	        this.eventEngineEnabled = source["eventEngineEnabled"];
	        this.playbackEnabled = source["playbackEnabled"];
	        this.message = source["message"];
	        this.event = this.convertValues(source["event"], TerminalEvent);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}




}

export namespace integrations {

	export class CheckResult {
	    ok: boolean;
	    label: string;
	    detail: string;
	    path?: string;

	    static createFrom(source: any = {}) {
	        return new CheckResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.label = source["label"];
	        this.detail = source["detail"];
	        this.path = source["path"];
	    }
	}
	export class IntegrationStatus {
	    shell: string;
	    displayName: string;
	    supported: boolean;
	    platformRelevant: boolean;
	    shellExecutableFound: boolean;
	    shellExecutable?: string;
	    helperInstalled: boolean;
	    helperExecutable: boolean;
	    scriptInstalled: boolean;
	    profileConfigured: boolean;
	    profileState: string;
	    profilePath?: string;
	    helperPath: string;
	    scriptPath: string;
	    matcherCachePath: string;
	    currentSessionCommand: string;
	    disableSessionCommand: string;
	    debugEnableCommand: string;
	    debugDisableCommand: string;
	    debugLogCommand: string;
	    problems: string[];
	    warnings: string[];
	    checks: CheckResult[];

	    static createFrom(source: any = {}) {
	        return new IntegrationStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shell = source["shell"];
	        this.displayName = source["displayName"];
	        this.supported = source["supported"];
	        this.platformRelevant = source["platformRelevant"];
	        this.shellExecutableFound = source["shellExecutableFound"];
	        this.shellExecutable = source["shellExecutable"];
	        this.helperInstalled = source["helperInstalled"];
	        this.helperExecutable = source["helperExecutable"];
	        this.scriptInstalled = source["scriptInstalled"];
	        this.profileConfigured = source["profileConfigured"];
	        this.profileState = source["profileState"];
	        this.profilePath = source["profilePath"];
	        this.helperPath = source["helperPath"];
	        this.scriptPath = source["scriptPath"];
	        this.matcherCachePath = source["matcherCachePath"];
	        this.currentSessionCommand = source["currentSessionCommand"];
	        this.disableSessionCommand = source["disableSessionCommand"];
	        this.debugEnableCommand = source["debugEnableCommand"];
	        this.debugDisableCommand = source["debugDisableCommand"];
	        this.debugLogCommand = source["debugLogCommand"];
	        this.problems = source["problems"];
	        this.warnings = source["warnings"];
	        this.checks = this.convertValues(source["checks"], CheckResult);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DoctorResult {
	    ok: boolean;
	    platform: string;
	    configPath: string;
	    matcherCacheValid: boolean;
	    ruleCount: number;
	    listening: boolean;
	    eventEngineEnabled: boolean;
	    muted: boolean;
	    playbackEnabled: boolean;
	    playback: core.PlaybackStatus;
	    integrations: IntegrationStatus[];

	    static createFrom(source: any = {}) {
	        return new DoctorResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.platform = source["platform"];
	        this.configPath = source["configPath"];
	        this.matcherCacheValid = source["matcherCacheValid"];
	        this.ruleCount = source["ruleCount"];
	        this.listening = source["listening"];
	        this.eventEngineEnabled = source["eventEngineEnabled"];
	        this.muted = source["muted"];
	        this.playbackEnabled = source["playbackEnabled"];
	        this.playback = this.convertValues(source["playback"], core.PlaybackStatus);
	        this.integrations = this.convertValues(source["integrations"], IntegrationStatus);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {

	export class AudioToolsStatus {
	    ffmpegAvailable: boolean;
	    ffprobeAvailable: boolean;
	    ffmpegPath: string;
	    ffprobePath: string;
	    message: string;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new AudioToolsStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ffmpegAvailable = source["ffmpegAvailable"];
	        this.ffprobeAvailable = source["ffprobeAvailable"];
	        this.ffmpegPath = source["ffmpegPath"];
	        this.ffprobePath = source["ffprobePath"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class RenameSoundRequest {
	    id: string;
	    name: string;

	    static createFrom(source: any = {}) {
	        return new RenameSoundRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class RuleRequest {
	    id: string;
	    name: string;
	    enabled: boolean;
	    eventType: string;
	    soundId: string;
	    matchMode: string;
	    commandPattern: string;
	    exitCode?: number;

	    static createFrom(source: any = {}) {
	        return new RuleRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	        this.eventType = source["eventType"];
	        this.soundId = source["soundId"];
	        this.matchMode = source["matchMode"];
	        this.commandPattern = source["commandPattern"];
	        this.exitCode = source["exitCode"];
	    }
	}

}

