import { combine, createEffect, createStore, sample } from "effector";

import { api } from "../api";
import type { Playlist } from "../types";

function toErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export const validateYandexTokenFx = createEffect(async (token: string) => {
  return api<{ ok: boolean; user_id: string }>("/api/yandex/auth/validate", {
    method: "POST",
    body: JSON.stringify({ token }),
  });
});

export const loadYandexPlaylistsFx = createEffect(async () => {
  const data = await api<{ items: Playlist[] }>("/api/yandex/playlists");
  return data.items;
});

export const refreshYandexStatusFx = createEffect(async () => {
  // Prefer explicit status endpoint; fallback keeps compatibility with older backend.
  try {
    const status = await api<{ connected: boolean; user_id: string }>(
      "/api/yandex/status",
    );
    if (!status.connected) {
      return { connected: false, playlists: [] as Playlist[] };
    }
    const data = await api<{ items: Playlist[] }>("/api/yandex/playlists");
    return { connected: true, playlists: data.items };
  } catch {
    try {
      const data = await api<{ items: Playlist[] }>("/api/yandex/playlists");
      return { connected: true, playlists: data.items };
    } catch {
      return { connected: false, playlists: [] as Playlist[] };
    }
  }
});

export const clearYandexConnectionFx = createEffect(async () => {
  await api<{ ok: boolean }>("/api/yandex/clear", { method: "POST" });
});

export const getYandexAuthURLFx = createEffect(async () => {
  const data = await api<{ url: string }>("/api/yandex/auth/url");
  return data.url;
});

export const $yandexConnected = createStore(false)
  .on(loadYandexPlaylistsFx.done, () => true)
  .on(loadYandexPlaylistsFx.fail, () => false)
  .on(validateYandexTokenFx.fail, () => false)
  .on(clearYandexConnectionFx.done, () => false)
  .on(refreshYandexStatusFx.doneData, (_, payload) => payload.connected);

export const $yandexPlaylists = createStore<Playlist[]>([])
  .on(loadYandexPlaylistsFx.doneData, (_, items) => items)
  .on(loadYandexPlaylistsFx.fail, () => [])
  .on(clearYandexConnectionFx.done, () => [])
  .on(refreshYandexStatusFx.doneData, (_, payload) => payload.playlists);

export const $yandexError = createStore("")
  .on(validateYandexTokenFx, () => "")
  .on(loadYandexPlaylistsFx, () => "")
  .on(clearYandexConnectionFx.done, () => "")
  .on(validateYandexTokenFx.failData, (_, error) =>
    toErrorMessage(error, "Validation failed"),
  )
  .on(loadYandexPlaylistsFx.failData, (state, error) => {
    const message = toErrorMessage(error, "Failed to load playlists");
    if (message.includes("yandex is not connected")) {
      return state;
    }
    return message;
  });

export const $yandexLoading = combine(
  validateYandexTokenFx.pending,
  loadYandexPlaylistsFx.pending,
  refreshYandexStatusFx.pending,
  (validatePending, loadPending, refreshPending) =>
    validatePending || loadPending || refreshPending,
);

sample({
  clock: validateYandexTokenFx.done,
  target: loadYandexPlaylistsFx,
});
