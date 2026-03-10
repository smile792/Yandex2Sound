import { useEffect, useState } from "react";
import { api } from "../api";
import type { Playlist } from "../types";

export function useYandex() {
  const [connected, setConnected] = useState(false);
  const [playlists, setPlaylists] = useState<Playlist[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");

  async function validateToken(token: string) {
    setLoading(true);
    setError("");
    try {
      await api<{ ok: boolean }>("/api/yandex/auth/validate", {
        method: "POST",
        body: JSON.stringify({ token }),
      });
      setConnected(true);
      await loadPlaylists();
    } catch (e) {
      setConnected(false);
      setError(e instanceof Error ? e.message : "Validation failed");
    } finally {
      setLoading(false);
    }
  }

  async function loadPlaylists() {
    setLoading(true);
    setError("");
    try {
      const data = await api<{ items: Playlist[] }>("/api/yandex/playlists");
      setPlaylists(data.items);
      setConnected(true);
    } catch (e) {
      const message = e instanceof Error ? e.message : "Failed to load playlists";
      if (!message.includes("yandex is not connected")) {
        setError(message);
      }
      setConnected(false);
    } finally {
      setLoading(false);
    }
  }

  async function getAuthURL() {
    const data = await api<{ url: string }>("/api/yandex/auth/url");
    return data.url;
  }

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get("ym") === "connected") {
      loadPlaylists().catch(() => undefined);
      params.delete("ym");
      const next = params.toString();
      const cleanURL = `${window.location.pathname}${next ? `?${next}` : ""}`;
      window.history.replaceState({}, "", cleanURL);
      return;
    }
    loadPlaylists().catch(() => undefined);
  }, []);

  return { connected, playlists, loading, error, validateToken, loadPlaylists, getAuthURL };
}
