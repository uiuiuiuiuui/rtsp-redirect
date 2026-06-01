# rtsp-redirect

Промежуточный сервис: RTSP знает только бэкенд, пользователю отдаётся короткая ссылка с редиректом **301**.

## Схема

```
Бэкенд                             Пользователь / плеер
  | POST /api/streams {url: rtsp}      |
  |<-- {redirect_url: .../r/abc123}    |
  |         redirect_url               |
  |----------------------------------->|
  |                                    | GET /r/abc123 → 301 → rtsp://...
```

## API

### `POST /api/streams`

```json
{ "url": "rtsp://server:554/main/stream" }
```

Ответ:

```json
{
  "redirect_url": "https://rtsp-redirect.onrender.com/r/a1b2c3d4...",
  "token": "a1b2c3d4...",
  "expires_at": "2026-05-29T09:00:00Z"
}
```

### `GET /r/{token}`

**301** + `Location: rtsp://...`

### `GET /health`

200 OK.

## Примеры

```bash
curl -X POST "https://rtsp-redirect.onrender.com/api/streams" \
  -H "Content-Type: application/json" \
  -d '{"url":"rtsp://b2o-vcore29.video.goodline.info:554/main/oooprovzor_59584"}'

curl -i "https://rtsp-redirect.onrender.com/r/TOKEN_ИЗ_ОТВЕТА"
```

## Запуск

```bash
go run .
```

На Render ничего настраивать не нужно — порт подставит платформа сама.

Ссылка живёт **1 час**, потом 404.

## Ограничения

- Токены в памяти — после перезапуска ссылки пропадают.
- Браузер не откроет `rtsp://` — нужен VLC, ffmpeg или RTSP-клиент.
