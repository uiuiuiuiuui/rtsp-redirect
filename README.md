# rtsp-redirect

Стабильная **RTSP**-ссылка на поток камеры. При переезде камеры на другой vcore меняется только целевой RTSP в `Location`, стабильный URL не меняется.

```
Стабильно:  rtsp://redirect-host:8554/camera/key/59584
            rtsp://redirect-host:8554/59584          (короткий путь)

Было:       rtsp://vcore22.../oooprovzor_59584
Стало:      rtsp://vcore34.../oooprovzor_59584
```

Клиент делает `DESCRIBE` на стабильный URL → сервер отвечает **302 Found** + `Location: rtsp://vcore.../...` → клиент подключается к реальному потоку (VLC, ffmpeg, go2rtc и др. с поддержкой RTSP redirect).

## Порты и переменные

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `PORT` | `8080` | HTTP API (`POST /api/streams`, `GET /health`) |
| `RTSP_PORT` | `8554` | Порт RTSP-редиректа |
| `RTSP_LISTEN` | `:8554` | Адрес bind RTSP (перекрывает `RTSP_PORT`) |
| `RTSP_PUBLIC_HOST` | `127.0.0.1` | Хост в `redirect_url` (публичный IP/DNS) |
| `RTSP_PUBLIC_PORT` | как `RTSP_PORT` | Порт в `redirect_url` |

**Render** подходит только для HTTP API. RTSP нужен VPS/VM с открытым TCP **8554** (или своим портом).

## API (HTTP, только регистрация)

### `POST /api/streams`

```json
{
  "url": "rtsp://vcore22.video.goodline.info:554/main/ooprovzor_59584"
}
```

Ответ:

```json
{
  "id": "59584",
  "redirect_url": "rtsp://redirect-host:8554/camera/key/59584",
  "url": "rtsp://vcore22.video.goodline.info:554/main/ooprovzor_59584"
}
```

### `GET /health`

200 OK.

## RTSP (основной протокол)

### Стабильные URL

- `rtsp://{host}:{port}/camera/key/{id}` — рекомендуемый (как в ТЗ)
- `rtsp://{host}:{port}/{id}` — короткий вариант

### Поведение

На `DESCRIBE` (и `SETUP` без сессии):

```
RTSP/1.0 302 Found
Location: rtsp://vcore22.video.goodline.info:554/main/ooprovzor_59584
```

## Локальный запуск

```powershell
$env:RTSP_PUBLIC_HOST = "127.0.0.1"
go run .
```

```powershell
# Регистрация
$body = '{"url":"rtsp://vcore22.video.goodline.info:554/main/ooprovzor_59584"}'
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/streams" -Method POST -ContentType "application/json" -Body $body

# Проверка redirect (ffmpeg)
ffprobe -rtsp_transport tcp "rtsp://127.0.0.1:8554/camera/key/59584"
```

## Ограничения

- ID из имени потока: `ooprovzor_59584` → `59584`, или явный `id` в POST.
- Данные в памяти — после рестарта снова `POST /api/streams`.
- `RTSP_PUBLIC_HOST` обязателен на проде (не `127.0.0.1`).

## Тесты

```bash
go test ./...
```
