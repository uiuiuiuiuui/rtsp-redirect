# rtsp-redirect

Стабильные HTTP-ссылки на RTSP-потоки камер. ID камеры берётся из имени потока (`ooprovzor_123` → `123`).

При переезде камеры на другой сервер бэкенд обновляет RTSP через `POST /api/streams`, а ссылка для пользователя остаётся той же:

```
rtsp://vcore22.../ooprovzor_123  →  https://host/123
rtsp://vcore34.../ooprovzor_123  →  https://host/123   (та же ссылка)
```

## API

### `POST /api/streams`

Регистрация или обновление потока.

```json
{
  "url": "rtsp://vcore22.video.goodline.info:554/main/ooprovzor_123"
}
```

Опционально явный id:

```json
{
  "id": "123",
  "url": "rtsp://vcore22.video.goodline.info:554/main/ooprovzor_123"
}
```

Ответ:

```json
{
  "id": "123",
  "redirect_url": "https://rtsp-redirect.onrender.com/123",
  "url": "rtsp://vcore22.video.goodline.info:554/main/ooprovzor_123"
}
```

### `GET /{id}`

**301 Moved Permanently** + `Location: rtsp://...`

### `GET /health`

200 OK.

## Деплой на Render

1. Закоммить и запушить в `master` репозитория `uiuiuiuiuui/rtsp-redirect`.
2. Render подхватит сборку автоматически (`go build -o app`).
3. Переменные окружения не нужны — Render сам задаёт `PORT`.

## Полный цикл тестирования

### 1. Health

```powershell
curl.exe -s -w "`nHTTP:%{http_code}`n" "https://rtsp-redirect.onrender.com/health"
```

### 2. Регистрация на vcore22

```powershell
$body = '{"url":"rtsp://vcore22.video.goodline.info:554/main/ooprovzor_123"}'
Invoke-RestMethod -Uri "https://rtsp-redirect.onrender.com/api/streams" -Method POST -ContentType "application/json" -Body $body
```

Ожидаешь `redirect_url: .../123`.

### 3. Редирект

```powershell
curl.exe -s -i "https://rtsp-redirect.onrender.com/123"
```

Ожидаешь:

```
HTTP/1.1 301 Moved Permanently
Location: rtsp://vcore22.video.goodline.info:554/main/ooprovzor_123
```

### 4. Камера переехала на vcore34 — обновляем RTSP

```powershell
$body = '{"url":"rtsp://vcore34.video.goodline.info:554/main/ooprovzor_123"}'
Invoke-RestMethod -Uri "https://rtsp-redirect.onrender.com/api/streams" -Method POST -ContentType "application/json" -Body $body
```

`redirect_url` снова `.../123` — **не меняется**.

### 5. Тот же redirect_url, новый RTSP

```powershell
curl.exe -s -i "https://rtsp-redirect.onrender.com/123"
```

Ожидаешь `Location: rtsp://vcore34...`.

## Ограничения

- ID извлекается из последнего сегмента пути: `ooprovzor_123` → `123`.
- Данные в памяти: после рестарта Render нужно заново вызвать `POST /api/streams` (обычно делает бэкенд при старте / переезде камеры).
- Браузер не откроет `rtsp://` — нужен VLC, ffmpeg или RTSP-клиент.

## Локально

```bash
go test ./...
go run .
```
