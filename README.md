# Media Server API
## I made a WebUI in svelte so it's at [Media Viewer](https://github.com/Benson003/media_server_ui)

A simple media streaming server written in Go — stream your videos, serve thumbnails, and enjoy a slick REST API with pagination and Swagger docs.

---

## Features

- Stream media files with HTTP range support
- Paginated media listing API
- Thumbnail extraction at 4 seconds using FFmpeg
- Swagger/OpenAPI documentation
- SQLite backend via GORM ORM
- Configurable media directories scanning

---

## Prerequisites

- [Go](https://go.dev/dl/) 1.20+
- [FFmpeg](https://ffmpeg.org/download.html) installed and in your PATH (required for thumbnail generation)
- `swag` CLI tool for docs generation (optional, only if modifying docs):
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

---

## Getting Started

1. **Clone the repo**

   ```bash
   git clone https://github.com/Benson003/media-server.git
   cd media-server
   ```

2. **Generate Swagger docs (only if you change or add API annotations)**

   ```bash
   swag init
   ```

3. **Create a config file `config.json`**
   Example:

   ```json
   {
     "media_dirs": ["./media"]
   }
   ```

4. **Run the server**

   If your `main.go` is in the root directory:
   ```bash
   go run main.go
   ```
   If your entrypoint is in `cmd/media_server/main.go` (recommended Go structure):
   ```bash
   cd cmd/media_server
   go run .
   ```

   The server will start on `http://localhost:8000`

5. **Access API docs**
   Open your browser to:

   ```
   http://localhost:8000/docs/index.html
   ```

---

## API Endpoints

| Method | Path                    | Description              |
| ------ | ----------------------- | ------------------------ |
| GET    | `/media/paginated?page=1&count=10`                | Get paginated media list |
| GET    | `/media/all`            | Get all media items      |
| GET    | `/media/{id}`           | Get media item by ID     |
| GET    | `/media/{id}/stream`    | Stream media file        |
| GET    | `/media/{id}/thumbnail` | Get thumbnail image      |

---

## Configuration

The `config.json` controls which directories are scanned for media files:

```json
{
  "media_dirs": ["./media", "/path/to/other/media"]
}
```

---

## Troubleshooting

* **Swagger docs not loading?**  
  Make sure you ran `swag init` and that the generated docs exist in the `/docs` directory.

* **Thumbnail generation fails?**
  Ensure `ffmpeg` is installed and available in your system PATH.

* **CORS issues?**
  The server includes CORS middleware allowing common dev origins. Adjust in `main.go` if needed.

---

## License

MIT License — do whatever you want, but don’t blame me if your media collection explodes.

---

## Contact

Developed by Benson
Email: [nwankwobenson29@gmail.com](mailto:nwankwobenson29@gmail.com)

---

Enjoy streaming! 🚀
