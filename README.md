# Yandex2Sound

Перенос плейлистов из Яндекс Музыки в SoundCloud.

## Возможности

- Подключение Яндекс Музыки через OAuth `access_token` (вставка вручную).
- Подключение SoundCloud через OAuth (`client_id` + `client_secret` + callback).
- Выбор одного или нескольких плейлистов Яндекса, включая `Liked tracks`.
- Два режима переноса:
- Объединить выбранные плейлисты в один плейлист SoundCloud.
- Сохранить исходные названия и создать отдельные плейлисты в SoundCloud.
- Прогресс переноса в реальном времени через SSE.

## Технологии

- Backend: Go + Fiber
- Frontend: React + TypeScript + Vite + Tailwind
- State management: Effector
- Сессии: зашифрованная HTTP-only cookie (AES-GCM)

## Требования

- Go 1.23+
- Node.js 18+ (рекомендуется Node.js 20+)
- npm

## Структура проекта

- `backend/` - API, OAuth-обработчики, логика переноса
- `frontend/` - интерфейс и клиентская логика
- `docker-compose.yml` - опциональный контейнерный запуск

## Переменные окружения

Создай `backend/.env` на основе `backend/.env.example`.

Обязательные:

```env
SOUNDCLOUD_CLIENT_ID=your_client_id
SOUNDCLOUD_CLIENT_SECRET=your_client_secret
SOUNDCLOUD_REDIRECT_URI=http://localhost:8080/api/soundcloud/auth/callback
FRONTEND_URL=http://localhost:5173
```

Опциональные (только если нужен OAuth code flow Яндекса через backend):

```env
YANDEX_CLIENT_ID=your_yandex_client_id
YANDEX_CLIENT_SECRET=your_yandex_client_secret
YANDEX_REDIRECT_URI=http://localhost:8080/api/yandex/auth/callback
```

Примечания:

- `SESSION_SECRET` теперь не обязателен. Если не задан, backend генерирует ключ при старте процесса.
- `COOKIE_SECURE` больше не нужен. Флаг `Secure` для cookie определяется автоматически.

## Локальный запуск

1. Установи зависимости фронта:

```bash
cd frontend
npm install
```

2. Запусти backend + frontend одной командой:

```bash
npm run dev
```

Что делает команда:

- запускает backend из `../backend` на порту `8080` (если порт занят, запуск завершится с ошибкой)
- записывает выбранный порт backend в `frontend/.backend-port`
- запускает Vite и подставляет `VITE_API_URL` из этого порта

Открой адрес, который покажет Vite (обычно `http://localhost:5173`).

## Настройка OAuth

### SoundCloud

1. Открой `https://soundcloud.com/you/apps`.
2. Создай приложение или открой существующее.
3. Укажи Redirect URI:
   `http://localhost:8080/api/soundcloud/auth/callback`
4. Вставь `Client ID` и `Client Secret` в интерфейс приложения.
5. Нажми `Connect via OAuth`.

Если `Client ID` и `Client Secret` уже заданы в `backend/.env`, интерфейс может использовать их.

### Яндекс Музыка

Сейчас основной путь в интерфейсе: ручная проверка токена, а не полный OAuth через кнопку.

Используй подсказку в приложении `Как получить токен?` или шаги ниже:

1. Открой:
   `https://oauth.yandex.ru/authorize?response_type=token&client_id=23cabbbdc6cd418abb4b39c32c41195d`
2. Авторизуйся и выдай доступ.
3. Скопируй `access_token` из URL-фрагмента:
   `https://music.yandex.ru/#access_token=...&token_type=bearer...`
4. Вставь только значение токена и нажми `Validate token`.

## Сценарий использования

1. Подключи Яндекс.
2. Подключи SoundCloud.
3. Экран выбора плейлистов становится доступен только после активного подключения обоих сервисов.
4. Выбери плейлисты и режим переноса.
5. Запусти перенос и наблюдай прогресс.

## Режимы переноса

- `Preserve original names = off`:
- треки из выбранных плейлистов дедуплицируются и переносятся в один плейлист SoundCloud.
- имя берется из поля `New playlist name in SoundCloud`.
- `Preserve original names = on`:
- для каждого выбранного плейлиста Яндекса создается отдельный плейлист в SoundCloud.
- исходные названия сохраняются.

## Актуальные API endpoints

- `POST /api/yandex/auth/validate`
- `POST /api/yandex/clear`
- `GET /api/yandex/status`
- `GET /api/yandex/playlists`
- `GET /api/yandex/playlist/:id/tracks`
- `POST /api/soundcloud/config`
- `POST /api/soundcloud/clear`
- `GET /api/soundcloud/auth/url`
- `GET /api/soundcloud/auth`
- `GET /api/soundcloud/auth/callback`
- `GET /api/soundcloud/status`
- `POST /api/transfer`
- `GET /api/transfer/progress/:job_id`

## Docker (опционально)

```bash
docker compose up --build
```

Backend читает `backend/.env` через `env_file`.

## Частые проблемы

- `failed to listen ... :8080`:
- порт `8080` уже занят, останови конфликтующий процесс.
- Ошибки SoundCloud OAuth callback (`403`, `502`):
- проверь точное совпадение Redirect URI в приложении SoundCloud:
- `http://localhost:8080/api/soundcloud/auth/callback`
- `401 Unauthorized` от Яндекса:
- токен недействителен или не подходит для Music API, сгенерируй новый.
- После OAuth открывается не тот фронтенд URL:
- проверь `FRONTEND_URL` в `backend/.env`.
- Сессия сбрасывается после рестарта backend:
- это ожидаемо, если не задан `SESSION_SECRET` (временный ключ процесса).
