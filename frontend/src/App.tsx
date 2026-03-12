import { useEffect } from "react";
import { useUnit } from "effector-react";

import { api } from "./api";
import { ConnectView } from "./components/app/ConnectView";
import { LibraryView } from "./components/app/LibraryView";
import { ProgressView } from "./components/app/ProgressView";
import { useSSE } from "./hooks/useSSE";
import {
  $jobId,
  $playlistName,
  $preserveOriginalNames,
  $selected,
  $showConnectScreen,
  clearSelectedPlaylists,
  selectPlaylists,
  setJobId,
  setPlaylistName,
  setPreserveOriginalNames,
  setShowConnectScreen,
  togglePlaylist,
} from "./model/appUi";
import {
  $soundCloudClientId,
  $soundCloudConnected,
  $soundCloudHasClientSecret,
  $soundCloudUsername,
  clearSoundCloudCredentialsFx,
  getSoundCloudAuthURLFx,
  refreshSoundCloudStatusFx,
  saveSoundCloudCredentialsFx,
} from "./model/soundcloud";
import {
  $yandexConnected,
  $yandexError,
  $yandexPlaylists,
  clearYandexConnectionFx,
  loadYandexPlaylistsFx,
  refreshYandexStatusFx,
  validateYandexTokenFx,
} from "./model/yandex";

export default function App() {
  const {
    yandexConnected,
    playlists,
    yandexError,
    scConnected,
    scUsername,
    scClientId,
    scHasClientSecret,
  } = useUnit({
    yandexConnected: $yandexConnected,
    playlists: $yandexPlaylists,
    yandexError: $yandexError,
    scConnected: $soundCloudConnected,
    scUsername: $soundCloudUsername,
    scClientId: $soundCloudClientId,
    scHasClientSecret: $soundCloudHasClientSecret,
  });

  const {
    selected,
    playlistName,
    preserveOriginalNames,
    jobId,
    showConnectScreen,
  } = useUnit({
    selected: $selected,
    playlistName: $playlistName,
    preserveOriginalNames: $preserveOriginalNames,
    jobId: $jobId,
    showConnectScreen: $showConnectScreen,
  });
  const {
    togglePlaylistFx,
    selectPlaylistsFx,
    clearSelectedPlaylistsFx,
    setPlaylistNameFx,
    setPreserveOriginalNamesFx,
    setJobIdFx,
    setShowConnectScreenFx,
  } = useUnit({
    togglePlaylistFx: togglePlaylist,
    selectPlaylistsFx: selectPlaylists,
    clearSelectedPlaylistsFx: clearSelectedPlaylists,
    setPlaylistNameFx: setPlaylistName,
    setPreserveOriginalNamesFx: setPreserveOriginalNames,
    setJobIdFx: setJobId,
    setShowConnectScreenFx: setShowConnectScreen,
  });

  const {
    validateYandexTokenFxBound,
    clearYandexConnectionFxBound,
    loadYandexPlaylistsFxBound,
    refreshYandexStatusFxBound,
    clearSoundCloudCredentialsFxBound,
    refreshSoundCloudStatusFxBound,
    saveSoundCloudCredentialsFxBound,
    getSoundCloudAuthURLFxBound,
  } = useUnit({
    validateYandexTokenFxBound: validateYandexTokenFx,
    clearYandexConnectionFxBound: clearYandexConnectionFx,
    loadYandexPlaylistsFxBound: loadYandexPlaylistsFx,
    refreshYandexStatusFxBound: refreshYandexStatusFx,
    clearSoundCloudCredentialsFxBound: clearSoundCloudCredentialsFx,
    refreshSoundCloudStatusFxBound: refreshSoundCloudStatusFx,
    saveSoundCloudCredentialsFxBound: saveSoundCloudCredentialsFx,
    getSoundCloudAuthURLFxBound: getSoundCloudAuthURLFx,
  });

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const yandexConnectedFromOAuth = params.get("ym") === "connected";
    const soundCloudConnectedFromOAuth = params.get("sc") === "connected";

    if (yandexConnectedFromOAuth) {
      loadYandexPlaylistsFxBound().catch(() => undefined);
      params.delete("ym");
    } else {
      refreshYandexStatusFxBound().catch(() => undefined);
    }

    if (soundCloudConnectedFromOAuth) {
      refreshSoundCloudStatusFxBound().catch(() => undefined);
      params.delete("sc");
    } else {
      refreshSoundCloudStatusFxBound().catch(() => undefined);
    }

    if (yandexConnectedFromOAuth || soundCloudConnectedFromOAuth) {
      const next = params.toString();
      const cleanURL = `${window.location.pathname}${next ? `?${next}` : ""}`;
      window.history.replaceState({}, "", cleanURL);
    }
  }, [
    loadYandexPlaylistsFxBound,
    refreshYandexStatusFxBound,
    refreshSoundCloudStatusFxBound,
  ]);

  const sse = useSSE(jobId);

  const selectedCount = selected.size;
  const canGoLibrary = yandexConnected && scConnected;
  const view = jobId
    ? "progress"
    : showConnectScreen || !canGoLibrary
      ? "connect"
      : "library";

  async function handleSCClear() {
    await clearSoundCloudCredentialsFxBound();
    await refreshSoundCloudStatusFxBound();
    setShowConnectScreenFx(true);
  }

  async function handleYandexConnect(token: string) {
    await validateYandexTokenFxBound(token);
  }

  async function handleSCConnect(clientId: string, clientSecret: string) {
    const nextClientID = clientId.trim();
    const nextClientSecret = clientSecret.trim();
    const canUseStoredSecret =
      scHasClientSecret &&
      nextClientID === scClientId &&
      nextClientSecret === "";

    if (!canUseStoredSecret) {
      if (!nextClientID || !nextClientSecret) {
        throw new Error("Client ID and Client Secret are required");
      }
      await saveSoundCloudCredentialsFxBound({
        clientId: nextClientID,
        clientSecret: nextClientSecret,
      });
    }

    const url = await getSoundCloudAuthURLFxBound();
    window.location.href = url;
  }

  function toggle(id: string) {
    togglePlaylistFx(id);
  }

  function handleSelectAll() {
    selectPlaylistsFx(playlists.map(({ id }) => id));
  }

  function handleDeselectAll() {
    clearSelectedPlaylistsFx();
  }

  async function startTransfer() {
    const data = await api<{ job_id: string }>("/api/transfer", {
      method: "POST",
      body: JSON.stringify({
        playlist_ids: [...selected],
        playlist_name: playlistName,
        preserve_original_names: preserveOriginalNames,
      }),
    });
    setJobIdFx(data.job_id);
  }

  if (view === "connect") {
    return (
      <ConnectView
        yandexConnected={yandexConnected}
        scConnected={scConnected}
        scUsername={scUsername}
        scClientId={scClientId}
        scHasClientSecret={scHasClientSecret}
        onYandexConnect={handleYandexConnect}
        onYandexClear={clearYandexConnectionFxBound}
        onSCClear={handleSCClear}
        onSCConnect={handleSCConnect}
        error={yandexError}
        canGoLibrary={canGoLibrary}
        showConnectScreen={showConnectScreen}
        onGoToLibrary={() => setShowConnectScreenFx(false)}
      />
    );
  }

  if (view === "library") {
    return (
      <LibraryView
        playlists={playlists}
        selected={selected}
        onToggle={toggle}
        onSelectAll={handleSelectAll}
        onDeselectAll={handleDeselectAll}
        preserveOriginalNames={preserveOriginalNames}
        onPreserveOriginalNamesChange={setPreserveOriginalNamesFx}
        playlistName={playlistName}
        onPlaylistNameChange={setPlaylistNameFx}
        selectedCount={selectedCount}
        canTransfer={scConnected}
        onStartTransfer={() => startTransfer().catch((e) => alert(e.message))}
        onBack={() => setShowConnectScreenFx(true)}
      />
    );
  }

  return (
    <ProgressView
      progress={sse.progress}
      log={sse.log}
      done={sse.done}
      onTransferMore={() => setJobIdFx(null)}
    />
  );
}

