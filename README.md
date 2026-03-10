# Yandex2Sound

Yandex2Sound переносит треки из плейлистов Яндекс Музыки в SoundCloud.

## Что умеет
- Подключение Яндекс Музыки по `access_token`.
- Подключение SoundCloud через OAuth.
- Перенос выбранных плейлистов.
- Режим "Сохранять исходные названия плейлистов" (создаёт отдельный плейлист в SoundCloud для каждого выбранного).
- Прогресс переноса в реальном времени (SSE).

## Стек
- Backend: Go + Fiber v2
- Frontend: React + TypeScript + Tailwind CSS
- Сессии: in-memory `sync.Map`

## Требования
- Go 1.23+
- Node.js 18+ (лучше 20+)
- npm

## Структура проекта
- `backend/` — API и логика переноса
- `frontend/` — интерфейс
- `docker-compose.yml` — контейнерный запуск

## Быстрый запуск

### 1) Подготовка `.env`
Скопируй шаблон:

```bash
copy backend\.env.example backend\.env
```

Заполни `backend/.env`:

```env
SOUNDCLOUD_CLIENT_ID=your_client_id
SOUNDCLOUD_CLIENT_SECRET=your_client_secret
SOUNDCLOUD_REDIRECT_URI=http://localhost:8080/api/soundcloud/auth/callback
FRONTEND_URL=http://localhost:5173
```

### 2) Установи зависимости фронта

```bash
cd frontend
npm install
```

### 3) Запусти оба сервера одной командой

```bash
npm run dev
```

Открой фронт по адресу, который покажет Vite (обычно `http://localhost:5173`).

## Настройка SoundCloud OAuth
1. Открой `https://soundcloud.com/you/apps`.
2. Создай приложение.
3. Добавь Redirect URI: `http://localhost:8080/api/soundcloud/auth/callback`.
4. Скопируй `Client ID` и `Client Secret` в `backend/.env`.

## Как получить токен Яндекс Музыки
В приложении на экране подключения есть кнопка **"Как получить токен?"** с инструкцией.

Кратко:
1. Открой ссылку:
   `https://oauth.yandex.ru/authorize?response_type=token&client_id=23cabbbdc6cd418abb4b39c32c41195d`
2. Авторизуйся и выдай доступ.
3. Скопируй `access_token` из URL вида:
   `https://music.yandex.ru/#access_token=...&token_type=bearer...`
4. Вставь токен в поле Яндекса и нажми `Validate token`.

## Сценарий использования
1. Подключи Яндекс токеном.
2. Подключи SoundCloud через OAuth.
3. Выбери плейлисты.
4. При необходимости включи/выключи режим сохранения исходных названий.
5. Нажми `Transfer` и дождись завершения.

## Ограничения
- Используется неофициальный API Яндекс Музыки.
- Поиск треков в SoundCloud может быть неточным.
- Некоторые SoundCloud-приложения могут иметь ограничения API (403 на отдельных методах).
- Сессии хранятся в памяти процесса и теряются после перезапуска.

## Docker (опционально)

```bash
docker compose up --build
```

## Полезно при проблемах
- Если не работает SoundCloud OAuth: проверь совпадение Redirect URI в `backend/.env` и в настройках приложения SoundCloud.
- Если Яндекс токен не принимается: сгенерируй новый и вставь только значение `access_token`.
- После изменения `backend/.env` всегда перезапускай `npm run dev`.
