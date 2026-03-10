import { useEffect, useState } from "react";
import { api } from "../api";

export function useSoundCloud() {
  const [connected, setConnected] = useState(false);
  const [username, setUsername] = useState("");

  async function getAuthURL() {
    const data = await api<{ url: string }>("/api/soundcloud/auth/url");
    return data.url;
  }

  async function refreshStatus() {
    try {
      const data = await api<{ connected: boolean; username: string }>("/api/soundcloud/status");
      setConnected(data.connected);
      setUsername(data.username);
    } catch {
      setConnected(false);
      setUsername("");
    }
  }

  useEffect(() => {
    refreshStatus().catch(() => undefined);
  }, []);

  return { connected, username, getAuthURL, refreshStatus };
}

