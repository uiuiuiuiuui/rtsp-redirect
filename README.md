# rtsp-redirect

Простой HTTP-сервис: принимает ссылку на RTSP-поток и отдаёт редирект **301 Moved Permanently** с `Location`, указывающим на этот RTSP URL.

## Запуск

```bash
cd rtsp-redirect
go run .
```

Переменные окружения:

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `PORT`     | `8080`       | Порт HTTP |
| `BASE_URL` | —            | Базовый URL для `/link` (например `https://redirect.example.com`) |

## API

### `GET /redirect?url=<rtsp_url>`

Ответ **301**, заголовок `Location: <rtsp_url>`.

Пример:

```bash
curl -i "http://localhost:8080/redirect?url=rtsp%3A%2F%2Fuser%3Apass%4010.0.0.1%3A554%2Fstream"
```

### `GET /link?url=<rtsp_url>`

JSON со ссылкой на HTTP-редирект (удобно отдать клиенту «промежуточную» ссылку):

```json
{
  "redirect_url": "http://localhost:8080/redirect?url=rtsp%3A%2F%2F...",
  "location": "rtsp://..."
}
```

### `POST /redirect` или `POST /link`

Тело формы: `url=rtsp://...` — то же поведение, что у GET.

### `GET /health`

Проверка живости (200 OK).

## Ограничения

- Разрешены только схемы `rtsp://` и `rtsps://`.
- Параметр `url` передавайте в URL-encoded виде.
