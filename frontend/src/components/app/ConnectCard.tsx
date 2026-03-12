import { useEffect, useState } from "react";

const YANDEX_TOKEN_URL =
  "https://oauth.yandex.ru/authorize?response_type=token&client_id=23cabbbdc6cd418abb4b39c32c41195d";

export type ConnectCardProps = {
  yandexConnected: boolean;
  scConnected: boolean;
  scUsername: string;
  scClientId: string;
  scHasClientSecret: boolean;
  onYandexConnect: (token: string) => Promise<void>;
  onYandexClear: () => Promise<void>;
  onSCClear: () => Promise<void>;
  onSCConnect: (clientId: string, clientSecret: string) => Promise<void>;
  error: string;
};

export function ConnectCard({
  yandexConnected,
  scConnected,
  scUsername,
  scClientId,
  scHasClientSecret,
  onYandexConnect,
  onYandexClear,
  onSCClear,
  onSCConnect,
  error,
}: ConnectCardProps) {
  const [token, setToken] = useState("");
  const [scClientIdInput, setScClientIdInput] = useState(scClientId ?? "");
  const [scClientSecretInput, setScClientSecretInput] = useState("");
  const [showClientSecret, setShowClientSecret] = useState(false);
  const [showTokenHelp, setShowTokenHelp] = useState(false);
  const [showSoundCloudHelp, setShowSoundCloudHelp] = useState(false);

  useEffect(() => {
    setScClientIdInput(scClientId ?? "");
  }, [scClientId]);

  const canConnectSoundCloud =
    !!scClientIdInput && (!!scClientSecretInput || scHasClientSecret);
  const soundCloudStatusText = scConnected
    ? scUsername
      ? `${scUsername} подключено`
      : "подключено"
    : "не подключено";

  function handleYandexValidate() {
    onYandexConnect(token).catch(() => undefined);
  }

  function handleYandexTokenClear() {
    setToken("");
    onYandexClear().catch((e: Error) => {
      alert(e.message);
    });
  }

  function handleSoundCloudConnect() {
    onSCConnect(scClientIdInput, scClientSecretInput).catch(() => undefined);
  }

  function handleSoundCloudClear() {
    setScClientIdInput("");
    setScClientSecretInput("");
    onSCClear().catch((e: Error) => {
      alert(e.message);
    });
  }

  return (
    <section className="surface-panel ui-fade-up mx-auto mt-10 max-w-5xl rounded-3xl p-6 md:p-8">
      <h1 className="display-title text-4xl font-bold md:text-5xl">
        Yandex2Sound
      </h1>
      <p className="mt-2 text-zinc-400">
        Перенос плейлистов из Яндекс Музыки в SoundCloud
      </p>

      <div className="mt-6 grid gap-4 md:grid-cols-2">
        <div className="surface-card rounded-2xl p-5">
          <h2 className="font-bold text-yandex">Яндекс Музыка</h2>
          <input
            id="yandex-oauth-token"
            name="yandex_oauth_token"
            className="input-field mt-3"
            placeholder="OAuth token"
            value={token}
            onChange={(e) => setToken(e.target.value)}
          />
          <div className="mt-1 flex flex-wrap gap-2">
            <button className="btn btn-yandex" onClick={handleYandexValidate}>
              Validate token
            </button>
            <button
              className="btn btn-ghost text-sm"
              onClick={handleYandexTokenClear}
              type="button"
            >
              Clear token
            </button>
            <button
              className="btn btn-ghost text-sm"
              onClick={() => setShowTokenHelp(true)}
              type="button"
            >
              Как получить токен?
            </button>
          </div>
          <p className="mt-2 text-sm">
            Статус: {yandexConnected ? "подключено" : "не подключено"}
          </p>
        </div>

        <div className="surface-card rounded-2xl p-5">
          <h2 className="font-bold text-soundcloud">SoundCloud</h2>
          <input
            id="soundcloud-client-id"
            name="soundcloud_client_id"
            className="input-field mt-3"
            placeholder="Client ID"
            value={scClientIdInput}
            onChange={(e) => setScClientIdInput(e.target.value)}
          />
          <div className="mt-2 flex gap-2">
            <input
              id="soundcloud-client-secret"
              name="soundcloud_client_secret"
              className="input-field"
              placeholder="Client Secret"
              type={showClientSecret ? "text" : "password"}
              value={scClientSecretInput}
              onChange={(e) => setScClientSecretInput(e.target.value)}
            />
            <button
              className="btn btn-ghost whitespace-nowrap px-3 text-sm"
              onClick={() => setShowClientSecret((v) => !v)}
              type="button"
            >
              {showClientSecret ? "👀" : "👁️"}
            </button>
          </div>
          {scHasClientSecret && !scClientSecretInput && (
            <p className="mt-1 text-xs text-zinc-400">
              Client Secret сохранен на сервере и скрыт.
            </p>
          )}
          <div className="mt-1 flex flex-wrap gap-2">
            <button
              className="btn btn-soundcloud  disabled:cursor-not-allowed disabled:opacity-50"
              onClick={handleSoundCloudConnect}
              disabled={!canConnectSoundCloud}
            >
              Connect via OAuth
            </button>
            <button
              className="btn btn-ghost text-sm"
              onClick={handleSoundCloudClear}
              type="button"
            >
              Clear
            </button>
            <button
              className="btn btn-ghost text-sm"
              onClick={() => setShowSoundCloudHelp(true)}
              type="button"
            >
              Как получить Client ID / Secret?
            </button>
          </div>
          <p className="mt-2 text-sm">Статус: {soundCloudStatusText}</p>
        </div>
      </div>

      {error && <p className="mt-4 text-sm text-red-400">{error}</p>}

      {showTokenHelp && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div className="surface-panel w-full max-w-2xl rounded-2xl p-5">
            <div className="flex items-start justify-between">
              <h3 className="text-xl font-bold">
                Как получить access_token Яндекс
              </h3>
              <button
                className="btn btn-ghost px-2 py-1 text-sm"
                onClick={() => setShowTokenHelp(false)}
                type="button"
              >
                Закрыть
              </button>
            </div>
            <ol className="mt-4 list-decimal space-y-2 pl-5 text-sm text-zinc-200">
              <li>
                (Опционально) В DevTools на вкладке Network включи троттлинг.
              </li>
              <li>
                Перейди по ссылке{" "}
                <a
                  className="text-yandex underline"
                  href={YANDEX_TOKEN_URL}
                  target="_blank"
                  rel="noreferrer"
                >
                  Получить токен Яндекс
                </a>
                .
              </li>
              <li>Авторизуйся и предоставь доступ.</li>
              <li>
                Скопируй URL вида{" "}
                <code>
                  https://music.yandex.ru/#access_token=...&amp;token_type=bearer...
                </code>
                до быстрого редиректа.
              </li>
              <li>
                Возьми значение после <code>access_token=</code> и вставь в поле
                токена.
              </li>
            </ol>
          </div>
        </div>
      )}

      {showSoundCloudHelp && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div className="surface-panel w-full max-w-2xl rounded-2xl p-5">
            <div className="flex items-start justify-between">
              <h3 className="text-xl font-bold">
                Где взять Client ID и Secret
              </h3>
              <button
                className="btn btn-ghost px-2 py-1 text-sm"
                onClick={() => setShowSoundCloudHelp(false)}
                type="button"
              >
                Закрыть
              </button>
            </div>
            <ol className="mt-4 list-decimal space-y-2 pl-5 text-sm text-zinc-200">
              <li>
                Открой{" "}
                <a
                  className="text-soundcloud underline"
                  href="https://soundcloud.com/you/apps"
                  target="_blank"
                  rel="noreferrer"
                >
                  soundcloud.com/you/apps
                </a>
                .
              </li>
              <li>Создай приложение (или открой существующее).</li>
              <li>
                В Redirect URI укажи:{" "}
                <code>http://localhost:8080/api/soundcloud/auth/callback</code>
              </li>
              <li>Скопируй Client ID и Client Secret в поля выше.</li>
              <li>Нажми Connect via OAuth.</li>
            </ol>
          </div>
        </div>
      )}
    </section>
  );
}
