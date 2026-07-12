import {useEffect, useMemo, useState} from 'react';
import {Gauge, Keyboard, Music2, Plug, Settings, SlidersHorizontal} from 'lucide-react';
import {LoadConfig, SaveConfig} from '../wailsjs/go/main/App';
import {core as wailsModels} from '../wailsjs/go/models';
import {PageHeader} from './components/PageHeader';
import {Sidebar} from './components/Sidebar';
import {LoadingState} from './components/LoadingState';
import {DashboardPage} from './pages/DashboardPage';
import {HotkeysPage} from './pages/HotkeysPage';
import {IntegrationsPage} from './pages/IntegrationsPage';
import {RulesPage} from './pages/RulesPage';
import {SettingsPage} from './pages/SettingsPage';
import {SoundsPage} from './pages/SoundsPage';
import {AppConfig, ConfigSnapshot, LoadState, PageKey, pages} from './types/app';

function App() {
  const [activePage, setActivePage] = useState<PageKey>('dashboard');
  const [snapshot, setSnapshot] = useState<ConfigSnapshot | null>(null);
  const [draftConfig, setDraftConfig] = useState<AppConfig | null>(null);
  const [loadState, setLoadState] = useState<LoadState>('loading');
  const [message, setMessage] = useState('Loading ProdTag...');

  useEffect(() => {
    LoadConfig()
      .then((loadedSnapshot: ConfigSnapshot) => {
        setSnapshot(loadedSnapshot);
        setDraftConfig(loadedSnapshot.config);
        setLoadState('ready');
        setMessage('Config loaded');
      })
      .catch((error: unknown) => {
        setLoadState('error');
        setMessage(error instanceof Error ? error.message : 'Unable to load config');
      });
  }, []);

  const hasUnsavedChanges = useMemo(() => {
    if (!snapshot || !draftConfig) {
      return false;
    }

    return JSON.stringify(snapshot.config) !== JSON.stringify(draftConfig);
  }, [draftConfig, snapshot]);

  function updateDraft(patch: Partial<AppConfig>) {
    setDraftConfig((current) => current ? {...current, ...patch} : current);
  }

  function handleConfigUpdated(nextSnapshot: ConfigSnapshot) {
    setSnapshot(nextSnapshot);
    setDraftConfig(nextSnapshot.config);
    setLoadState('ready');
    setMessage('Sound library updated');
  }

  function saveConfig() {
    if (!draftConfig) {
      return;
    }

    setLoadState('saving');
    setMessage('Saving config...');
    SaveConfig(draftConfig as unknown as wailsModels.AppConfig)
      .then((savedSnapshot: ConfigSnapshot) => {
        setSnapshot(savedSnapshot);
        setDraftConfig(savedSnapshot.config);
        setLoadState('ready');
        setMessage('Config saved');
      })
      .catch((error: unknown) => {
        setLoadState('error');
        setMessage(error instanceof Error ? error.message : 'Unable to save config');
      });
  }

  const config = draftConfig ?? snapshot?.config ?? null;
  const paths = snapshot?.paths ?? null;

  return (
    <div className="app-shell">
      <Sidebar activePage={activePage} pages={pages} statusMessage={message} onNavigate={setActivePage} />

      <main className="min-w-0 flex-1 overflow-y-auto">
        <div className="mx-auto flex min-h-full w-full max-w-6xl flex-col px-8 py-6">
          <PageHeader
            title={pageTitle(activePage)}
            icon={pageHeaderIcon(activePage)}
            hasUnsavedChanges={hasUnsavedChanges}
            isSaving={loadState === 'saving'}
            canSave={Boolean(draftConfig) && hasUnsavedChanges}
            onSave={saveConfig}
          />

          {loadState === 'loading' && <LoadingState />}
          {loadState === 'error' && <ErrorPanel message={message} />}
          {config && activePage === 'dashboard' && <DashboardPage config={config} paths={paths} />}
          {config && activePage === 'sounds' && <SoundsPage config={config} onConfigUpdated={handleConfigUpdated} />}
          {config && activePage === 'rules' && <RulesPage config={config} onConfigUpdated={handleConfigUpdated} />}
          {config && activePage === 'hotkeys' && <HotkeysPage config={config} />}
          {config && activePage === 'integrations' && <IntegrationsPage config={config} />}
          {config && activePage === 'settings' && (
            <SettingsPage config={config} paths={paths} updateDraft={updateDraft} />
          )}
        </div>
      </main>
    </div>
  );
}

function ErrorPanel({message}: {message: string}) {
  return (
    <div className="rounded-lg border border-rose-200 bg-rose-50 p-5 text-rose-950">
      <div className="font-semibold">Could not open ProdTag config</div>
      <p className="mt-2 text-sm">{message}</p>
    </div>
  );
}

function pageTitle(page: PageKey) {
  return pages.find((item) => item.key === page)?.label ?? 'ProdTag';
}

function pageHeaderIcon(page: PageKey) {
  const iconProps = {size: 21, strokeWidth: 2};
  switch (page) {
    case 'dashboard':
      return <Gauge {...iconProps} />;
    case 'sounds':
      return <Music2 {...iconProps} />;
    case 'rules':
      return <SlidersHorizontal {...iconProps} />;
    case 'hotkeys':
      return <Keyboard {...iconProps} />;
    case 'integrations':
      return <Plug {...iconProps} />;
    case 'settings':
      return <Settings {...iconProps} />;
  }
}

export default App;
