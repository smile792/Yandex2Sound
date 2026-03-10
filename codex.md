# Codex Prompt: Yandex2Sound (Go + React)

> Full-stack приложение для переноса музыки из Яндекс Музыки в SoundCloud.

---

## Tech Stack

- **Backend:** Go + Fiber v2
- **Frontend:** React + TypeScript + Tailwind CSS
- **Auth:** OAuth2 для SoundCloud, token-based для Яндекс Музыки
- **Session storage:** in-memory `sync.Map` (без Redis)

---

## Backend (Go + Fiber v2)

### Структура проекта

```
backend/
├── cmd/
│   └── main.go
├── internal/
│   ├── handlers/
│   │   ├── yandex.go
│   │   ├── soundcloud.go
│   │   └── transfer.go
│   ├── services/
│   │   ├── yandex.go
│   │   ├── soundcloud.go
│   │   └── transfer.go
│   ├── models/
│   │   └── models.go
│   └── session/
│       └── store.go
├── go.mod
└── .env
```

### Зависимости (`go.mod`)

```
github.com/gofiber/fiber/v2
github.com/gofiber/fiber/v2/middleware/cors
github.com/gofiber/fiber/v2/middleware/session
github.com/joho/godotenv
github.com/google/uuid
```

---

### Models (`internal/models/models.go`)

```go
type Track struct {
    ID         string `json:"id"`
    Title      string `json:"title"`
    Artists    string `json:"artists"`
    Album      string `json:"album"`
    DurationMs int    `json:"duration_ms"`
    CoverURL   string `json:"cover_url"`
}

type Playlist struct {
    ID         string  `json:"id"`
    Title      string  `json:"title"`
    TrackCount int     `json:"track_count"`
    CoverURL   string  `json:"cover_url"`
    Tracks     []Track `json:"tracks,omitempty"`
}

type TransferJob struct {
    ID          string        `json:"id"`
    Status      string        `json:"status"` // pending|running|done|error
    Total       int           `json:"total"`
    Current     int           `json:"current"`
    Transferred int           `json:"transferred"`
    NotFound    int           `json:"not_found"`
    Errors      int           `json:"errors"`
    Log         []TransferLog `json:"log"`
    ResultURL   string        `json:"result_url"`
}

type TransferLog struct {
    TrackTitle string `json:"track_title"`
    Status     string `json:"status"` // found|not_found|error
}
```

---

### Session Store (`internal/session/store.go`)

Использовать `sync.Map` для хранения данных сессии:

```go
type SessionData struct {
    YandexToken      string
    SoundCloudToken  string
    SoundCloudUserID string
}

var store sync.Map // key: session_id (UUID), value: *SessionData
```

Генерировать `session_id` как UUID через `github.com/google/uuid`, возвращать как cookie `"sid"` при первом запросе.

---

### Yandex Music Service (`internal/services/yandex.go`)

Яндекс Музыка не имеет официального публичного API. Использовать прямые HTTP-запросы к внутреннему API.

**Base URL:** `https://api.music.yandex.net`

**Обязательные заголовки:**

```
Authorization: OAuth {token}
X-Yandex-Music-Client: WindowsPhone/3.20
```

**Методы:**

1. `ValidateToken(token string) (userID string, err error)`
   — `GET /account/status`

2. `GetPlaylists(token string) ([]Playlist, error)`
   — `GET /users/{userID}/playlists/list`

3. `GetLikedTracks(token string) ([]Track, error)`
   — `GET /users/{userID}/likes/tracks`
   — затем `GET /tracks?track-ids={ids joined by ,}`

4. `GetPlaylistTracks(token, userID, playlistKind string) ([]Track, error)`
   — `GET /users/{userID}/playlists/{kind}`

> Яндекс возвращает вложенный JSON с полем `"result"`. Парсить вручную.
> Поле Artists: объединить `artist.name` через `", "`.

---

### SoundCloud Service (`internal/services/soundcloud.go`)

**Base URL:** `https://api-v2.soundcloud.com`

**Методы:**

1. `GetAuthURL() string`
   — Строить OAuth2 URL:

   ```
   https://soundcloud.com/connect?client_id=...&redirect_uri=...&response_type=code&scope=non-expiring
   ```

2. `ExchangeCode(code string) (accessToken string, err error)`
   — `POST https://api.soundcloud.com/oauth2/token`
   — Body: `grant_type=authorization_code&code=...&client_id=...&client_secret=...&redirect_uri=...`

3. `GetMe(token string) (id string, username string, err error)`
   — `GET /me` с `Authorization: OAuth {token}`

4. `SearchTrack(token, query string) (trackID, permalinkURL string, found bool, err error)`
   — `GET /search/tracks?q={url-encoded query}&limit=1&client_id=...`
   — Заголовок: `Authorization: OAuth {token}`
   — Вернуть первый результат, если `collection` непустой.

5. `CreatePlaylist(token, name string) (playlistID, url string, err error)`
   — `POST /playlists`
   — Body JSON: `{"playlist": {"title": name, "sharing": "public", "tracks": []}}`
   — Заголовки: `Authorization: OAuth {token}`, `Content-Type: application/json`

6. `AddTrackToPlaylist(token, playlistID, trackID string) error`
   — `PUT /playlists/{playlistID}`
   — Сначала `GET /playlists/{playlistID}` для получения текущего списка треков
   — Затем `PUT` с добавленным `track_id`
   — Body: `{"playlist": {"tracks": [{"id": trackID}, ...]}}`

---

### Transfer Service (`internal/services/transfer.go`)

```go
var Jobs sync.Map // key: jobID, value: *TransferJob

func RunTransfer(jobID string, tracks []Track, scToken string, playlistName string) {
    // 1. Создать плейлист в SC
    // 2. Для каждого трека:
    //    a. query = track.Artists + " - " + track.Title
    //    b. SearchTrack(query)
    //    c. если найден: AddTrackToPlaylist
    //    d. обновить прогресс в Jobs sync.Map
    //    e. time.Sleep(300ms) — rate limit
    // 3. Установить job.Status = "done"
}
```

Запускать трансфер в горутине. Обновлять `Jobs` в реальном времени — SSE хендлер читает из него.

---

### API Endpoints

| Метод  | Путь                              | Описание                                     |
| ------ | --------------------------------- | -------------------------------------------- |
| `POST` | `/api/yandex/auth/validate`       | Принять токен, сохранить в сессию            |
| `GET`  | `/api/yandex/playlists`           | Вернуть плейлисты + "Liked tracks"           |
| `GET`  | `/api/yandex/playlist/:id/tracks` | Треки конкретного плейлиста                  |
| `GET`  | `/api/soundcloud/auth/url`        | Вернуть `{ "url": "..." }`                   |
| `GET`  | `/api/soundcloud/auth/callback`   | OAuth callback, редирект на `/?sc=connected` |
| `POST` | `/api/transfer`                   | Запустить трансфер, вернуть `job_id`         |
| `GET`  | `/api/transfer/progress/:job_id`  | **SSE** — стримить прогресс                  |

#### POST `/api/transfer` — тело запроса:

```json
{
  "playlist_ids": ["3", "liked"],
  "playlist_name": "From Yandex"
}
```

Собрать треки из всех выбранных плейлистов (дедупликация по `id`), запустить горутину `RunTransfer`, вернуть `{ "job_id": "..." }`.

#### GET `/api/transfer/progress/:job_id` — SSE endpoint:

```go
c.Set("Content-Type", "text/event-stream")
c.Set("Cache-Control", "no-cache")
c.Set("Connection", "keep-alive")

// Опрашивать Jobs каждые 500ms, стримить:
// data: {"current":5,"total":100,"status":"running","last_track":"...","log":[...]}
// При status == "done": отправить финальный event и закрыть.
```

### `main.go`

```go
app := fiber.New()
app.Use(cors.New(cors.Config{
    AllowOrigins:     "http://localhost:5173",
    AllowCredentials: true,
}))

// Регистрация роутов...

app.Listen(":8080")
```

Загружать `.env` через `godotenv.Load()` при старте.

---

## Frontend (React + TypeScript)

### Страницы

**1. Connect page**

- Яндекс: поле ввода OAuth-токена + инструкция как его получить
  _(DevTools → Application → Cookies → `Session_id` на music.yandex.ru)_
- SoundCloud: кнопка "Connect via OAuth" → открывает `/api/soundcloud/auth/url`
- Индикаторы статуса подключения: ✓ / не подключён

**2. Library page** _(после подключения Яндекса)_

- Сетка карточек плейлистов (обложка, название, кол-во треков)
- Карточка "Liked tracks ❤️" вверху
- Мультиселект с чекбоксами
- Кнопки "Select All" / "Deselect All"
- Поле ввода имени нового SoundCloud-плейлиста
- Нижняя панель: кнопка "Transfer N playlists →"

**3. Transfer progress page**

- Real-time прогресс-бар (через `EventSource`)
- Живой лог: название трека + иконка ✓/✗ на каждую строку (виртуализированный список)
- Статистика: transferred / not found / errors
- По завершении: кнопка "Open playlist in SoundCloud ↗" + "Transfer more"

### Хуки

| Хук               | Описание                                                        |
| ----------------- | --------------------------------------------------------------- |
| `useSSE(jobId)`   | Обёртка над `EventSource`, возвращает `{ progress, log, done }` |
| `useYandex()`     | Валидировать токен, загружать плейлисты                         |
| `useSoundCloud()` | Получать auth URL, проверять статус подключения                 |

### State Management

React Context или Zustand.

---

## UI / Дизайн

| Параметр          | Значение                          |
| ----------------- | --------------------------------- |
| Тема              | Тёмная                            |
| Background        | `#0F0F0F`                         |
| Cards             | `#1A1A1A`                         |
| Яндекс акцент     | `#FFCC00`                         |
| SoundCloud акцент | `#FF5500`                         |
| Шрифт             | `Space Mono` или `JetBrains Mono` |
| Анимации          | `framer-motion`                   |
| Layout            | Mobile-responsive                 |

---

## Обработка ошибок

- **Невалидный Яндекс-токен** → понятное сообщение + ссылка как получить токен
- **OAuth SoundCloud упал** → кнопка "Повторить"
- **Трек не найден на SC** → пропустить и залогировать, не прерывать трансфер
- **Rate limiting** → авто-retry с exponential backoff (макс. 3 попытки)
- **Сетевая ошибка во время трансфера** → возможность возобновить (сохранять прогресс в `localStorage`)

---

## `.env`

```env
SOUNDCLOUD_CLIENT_ID=your_client_id
SOUNDCLOUD_CLIENT_SECRET=your_client_secret
SOUNDCLOUD_REDIRECT_URI=http://localhost:8080/api/soundcloud/auth/callback
FRONTEND_URL=http://localhost:5173
```

---

## `docker-compose.yml`

```yaml
services:
  backend:
    build: ./backend
    ports: ["8080:8080"]
    env_file: ./backend/.env

  frontend:
    build: ./frontend
    ports: ["5173:5173"]
    environment:
      - VITE_API_URL=http://localhost:8080
```

---

## README должен включать

1. Как получить токен Яндекс Музыки (пошагово)
2. Как зарегистрировать SoundCloud app на `soundcloud.com/you/apps`
3. Инструкция по локальному запуску: `go run ./cmd/main.go` + `npm run dev`
4. Известные ограничения: неофициальный Яндекс API, нечёткий поиск на SC
