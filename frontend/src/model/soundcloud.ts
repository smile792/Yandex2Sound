import { createEffect, createStore } from "effector";

import { api } from "../api";

export const getSoundCloudAuthURLFx = createEffect(async () => {
  const data = await api<{ url: string }>("/api/soundcloud/auth/url");
  return data.url;
});

export const saveSoundCloudCredentialsFx = createEffect(
  async (payload: { clientId: string; clientSecret: string }) => {
    await api<{ ok: boolean }>("/api/soundcloud/config", {
      method: "POST",
      body: JSON.stringify({
        client_id: payload.clientId,
        client_secret: payload.clientSecret,
      }),
    });
    return payload.clientId;
  },
);

export const clearSoundCloudCredentialsFx = createEffect(async () => {
  await api<{ ok: boolean }>("/api/soundcloud/clear", { method: "POST" });
});

export const refreshSoundCloudStatusFx = createEffect(async () => {
  return api<{
    connected: boolean;
    username: string;
    client_id: string;
    has_client_secret: boolean;
  }>("/api/soundcloud/status");
});

export const $soundCloudConnected = createStore(false)
  .on(refreshSoundCloudStatusFx.doneData, (_, data) => Boolean(data.connected))
  .on(refreshSoundCloudStatusFx.fail, () => false)
  .on(clearSoundCloudCredentialsFx, () => false);

export const $soundCloudUsername = createStore("")
  .on(refreshSoundCloudStatusFx.doneData, (_, data) => data.username ?? "")
  .on(refreshSoundCloudStatusFx.fail, () => "")
  .on(clearSoundCloudCredentialsFx, () => "");

export const $soundCloudClientId = createStore("")
  .on(refreshSoundCloudStatusFx.doneData, (_, data) => data.client_id ?? "")
  .on(saveSoundCloudCredentialsFx.doneData, (_, clientId) => clientId)
  .on(refreshSoundCloudStatusFx.fail, () => "")
  .on(clearSoundCloudCredentialsFx, () => "");

export const $soundCloudHasClientSecret = createStore(false)
  .on(
    refreshSoundCloudStatusFx.doneData,
    (_, data) => Boolean(data.has_client_secret),
  )
  .on(saveSoundCloudCredentialsFx.done, () => true)
  .on(refreshSoundCloudStatusFx.fail, () => false)
  .on(clearSoundCloudCredentialsFx, () => false);
