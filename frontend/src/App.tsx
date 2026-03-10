import { useMemo, useState } from "react";
import { motion } from "framer-motion";

import { api } from "./api";
import { useSSE } from "./hooks/useSSE";
import { useSoundCloud } from "./hooks/useSoundCloud";
import { useYandex } from "./hooks/useYandex";
import { PlaylistGrid } from "./components/PlaylistGrid";

function ConnectCard(props: {
  yandexConnected: boolean;
  scConnected: boolean;
  onYandexConnect: (token: string) => Promise<void>;
  onSCConnect: () => Promise<void>;
  error: string;
}) {
  const [token, setToken] = useState("");
  const [showTokenHelp, setShowTokenHelp] = useState(false);
  const tokenUrl = "https://oauth.yandex.ru/authorize?response_type=token&client_id=23cabbbdc6cd418abb4b39c32c41195d";

  return (
    <section className="mx-auto mt-12 max-w-4xl rounded-2xl border border-zinc-700 bg-card p-6">
      <h1 className="text-3xl font-bold">Yandex2Sound</h1>
      <p className="mt-2 text-zinc-400">Перенос плейлистов из Яндекс Музыки в SoundCloud</p>

      <div className="mt-6 grid gap-4 md:grid-cols-2">
        <div className="rounded-xl border border-zinc-700 p-4">
          <h2 className="font-bold text-yandex">Яндекс Музыка</h2>
          <input
            className="mt-3 w-full rounded-lg border border-zinc-600 bg-zinc-900 p-2"
            placeholder="OAuth token"
            value={token}
            onChange={(e) => setToken(e.target.value)}
          />
          <button
            className="mt-3 rounded-lg bg-yandex px-3 py-2 font-bold text-black"
            onClick={() => {
              props.onYandexConnect(token).catch(() => undefined);
            }}
          >
            Validate token
          </button>
          <button
            className="mt-2 rounded-lg border border-zinc-600 px-3 py-2 text-sm text-zinc-200"
            onClick={() => {
              setShowTokenHelp(true);
            }}
          >
            Как получить токен?
          </button>
          <p className="mt-2 text-sm">Статус: {props.yandexConnected ? "? подключено" : "не подключено"}</p>
          <p className="mt-2 text-xs text-zinc-400">Вставь access_token из URL после авторизации.</p>
        </div>

        <div className="rounded-xl border border-zinc-700 p-4">
          <h2 className="font-bold text-soundcloud">SoundCloud</h2>
          <button
            className="mt-3 rounded-lg bg-soundcloud px-3 py-2 font-bold"
            onClick={() => props.onSCConnect()}
          >
            Connect via OAuth
          </button>
          <p className="mt-2 text-sm">Статус: {props.scConnected ? "? подключено" : "не подключено"}</p>
        </div>
      </div>

      {props.error && <p className="mt-4 text-sm text-red-400">{props.error}</p>}

      {showTokenHelp && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div className="w-full max-w-2xl rounded-xl border border-zinc-700 bg-zinc-900 p-5">
            <div className="flex items-start justify-between">
              <h3 className="text-xl font-bold">Как получить access_token Яндекс</h3>
              <button className="rounded border border-zinc-600 px-2 py-1 text-sm" onClick={() => setShowTokenHelp(false)}>
                Закрыть
              </button>
            </div>
            <ol className="mt-4 list-decimal space-y-2 pl-5 text-sm text-zinc-200">
              <li>(Опционально) В DevTools на вкладке Network включи троттлинг.</li>
              <li>
                Перейди по ссылке{" "}
                <a className="text-yandex underline" href={tokenUrl} target="_blank" rel="noreferrer">
                  Получить токен Яндекс
                </a>
                .
              </li>
              <li>Авторизуйся и предоставь доступ.</li>
              <li>
                Скопируй URL вида <code>https://music.yandex.ru/#access_token=...&amp;token_type=bearer...</code>
                до быстрого редиректа.
              </li>
              <li>Возьми значение после <code>access_token=</code> и вставь в поле токена.</li>
            </ol>
          </div>
        </div>
      )}
    </section>
  );
}

export default function App() {
  const yandex = useYandex();
  const sc = useSoundCloud();

  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [playlistName, setPlaylistName] = useState("From Yandex");
  const [preserveOriginalNames, setPreserveOriginalNames] = useState(true);
  const [jobId, setJobID] = useState<string | null>(null);
  const sse = useSSE(jobId);

  const selectedCount = selected.size;
  const canGoLibrary = yandex.connected && sc.connected;

  const view = useMemo(() => {
    if (jobId) return "progress";
    if (canGoLibrary) return "library";
    return "connect";
  }, [jobId, canGoLibrary]);

  async function handleSCConnect() {
	const url = await sc.getAuthURL();
	window.location.href = url;
  }

  function toggle(id: string) {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelected(next);
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
    setJobID(data.job_id);
  }

  if (view === "connect") {
    return (
      <main className="min-h-screen px-4 py-8 text-zinc-100">
        <ConnectCard
          yandexConnected={yandex.connected}
          scConnected={sc.connected}
          onYandexConnect={yandex.validateToken}
          onSCConnect={handleSCConnect}
          error={yandex.error}
        />
      </main>
    );
  }

  if (view === "library") {
    return (
      <main className="mx-auto min-h-screen max-w-6xl px-4 py-8 text-zinc-100">
        <h1 className="text-3xl font-bold">Выберите плейлисты</h1>
        <div className="mt-3 flex gap-2">
          <button className="rounded border border-zinc-600 px-3 py-2" onClick={() => setSelected(new Set(yandex.playlists.map((p) => p.id)))}>
            Select All
          </button>
          <button className="rounded border border-zinc-600 px-3 py-2" onClick={() => setSelected(new Set())}>
            Deselect All
          </button>
        </div>

        <div className="mt-6">
          <PlaylistGrid playlists={yandex.playlists} selected={selected} onToggle={toggle} />
        </div>

        <div className="mt-6 rounded-xl border border-zinc-700 bg-card p-4">
          <label className="flex items-center gap-2 text-sm text-zinc-300">
            <input
              type="checkbox"
              checked={preserveOriginalNames}
              onChange={(e) => setPreserveOriginalNames(e.target.checked)}
            />
            Сохранять исходные названия плейлистов
          </label>
          <label className="mt-4 block text-sm text-zinc-400">Новое имя плейлиста в SoundCloud</label>
          <input
            className="mt-2 w-full rounded-lg border border-zinc-600 bg-zinc-900 p-2"
            value={playlistName}
            onChange={(e) => setPlaylistName(e.target.value)}
            disabled={preserveOriginalNames}
          />
          <button
            className="mt-4 w-full rounded-lg bg-gradient-to-r from-yandex to-soundcloud px-3 py-3 font-bold text-black disabled:opacity-50"
            disabled={!sc.connected || selectedCount === 0}
            onClick={() => startTransfer().catch((e) => alert(e.message))}
          >
            Transfer {selectedCount} playlists {"->"}
          </button>
        </div>
      </main>
    );
  }

  return (
    <main className="mx-auto min-h-screen max-w-5xl px-4 py-8 text-zinc-100">
      <h1 className="text-3xl font-bold">Transfer Progress</h1>
      <div className="mt-4 h-3 overflow-hidden rounded-full bg-zinc-700">
        <motion.div
          className="h-full bg-gradient-to-r from-yandex to-soundcloud"
          animate={{ width: `${((sse.progress?.current ?? 0) / Math.max(sse.progress?.total ?? 1, 1)) * 100}%` }}
        />
      </div>
      <p className="mt-2 text-sm text-zinc-400">
        {sse.progress?.current ?? 0}/{sse.progress?.total ?? 0} · transferred: {sse.progress?.transferred ?? 0} · not found: {sse.progress?.not_found ?? 0} · errors: {sse.progress?.errors ?? 0}
      </p>

      <div className="mt-6 max-h-96 overflow-auto rounded-xl border border-zinc-700 bg-card p-4">
        {sse.log.map((item, i) => (
          <div className="border-b border-zinc-800 py-2 text-sm" key={`${item.track_title}-${i}`}>
            <span className="mr-2">{item.status === "found" ? "?" : "?"}</span>
            {item.track_title}
          </div>
        ))}
      </div>

      {sse.done && (
        <div className="mt-6 flex gap-3">
          {sse.progress?.result_url && (
            <a className="rounded-lg bg-soundcloud px-4 py-2 font-bold" href={sse.progress.result_url} target="_blank" rel="noreferrer">
              Open playlist in SoundCloud ?
            </a>
          )}
          <button className="rounded-lg border border-zinc-600 px-4 py-2" onClick={() => setJobID(null)}>
            Transfer more
          </button>
        </div>
      )}
    </main>
  );
}


