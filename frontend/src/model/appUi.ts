import { createEvent, createStore } from "effector";

export const togglePlaylist = createEvent<string>();
export const selectPlaylists = createEvent<string[]>();
export const clearSelectedPlaylists = createEvent();

export const setPlaylistName = createEvent<string>();
export const setPreserveOriginalNames = createEvent<boolean>();
export const setJobId = createEvent<string | null>();
export const setShowConnectScreen = createEvent<boolean>();

export const $selected = createStore<Set<string>>(new Set())
  .on(togglePlaylist, (state, id) => {
    const next = new Set(state);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    return next;
  })
  .on(selectPlaylists, (_, ids) => new Set(ids))
  .on(clearSelectedPlaylists, () => new Set());

export const $playlistName = createStore("From Yandex").on(
  setPlaylistName,
  (_, next) => next,
);

export const $preserveOriginalNames = createStore(true).on(
  setPreserveOriginalNames,
  (_, next) => next,
);

export const $jobId = createStore<string | null>(null).on(
  setJobId,
  (_, next) => next,
);

export const $showConnectScreen = createStore(false).on(
  setShowConnectScreen,
  (_, next) => next,
);
