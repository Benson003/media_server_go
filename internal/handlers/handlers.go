package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Benson003/media_server/internal/media"
	"github.com/go-chi/chi/v5"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Handler struct {
	mediaStore *media.MediaStore
}

func NewHandler(mediaStore *media.MediaStore) *Handler {
	return &Handler{mediaStore: mediaStore}
}

func (h *Handler) HandleGetAllMedia(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Parse query params
    pageStr := r.URL.Query().Get("page")
    limitStr := r.URL.Query().Get("limit")

    page := 1
    limit := 20 // default limit

    if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
        page = p
    }
    if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
        limit = l
    }

    // Fetch all media (or better: fetch only the needed slice)
    mediaList, err := h.mediaStore.GetAllMedia(ctx)
    if err != nil {
        http.Error(w, "Failed to get media list", http.StatusInternalServerError)
        return
    }

    total := len(mediaList)
    start := (page - 1) * limit
    if start > total {
        start = total
    }
    end := start + limit
    if end > total {
        end = total
    }

    pagedMedia := mediaList[start:end]

    // Return paginated result with metadata
    resp := map[string]interface{}{
        "page":  page,
        "limit": limit,
        "total": total,
        "data":  pagedMedia,
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}


func (h *Handler) HandleStreamMedia(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	mediaFile, err := h.mediaStore.GetMediaByID(ctx, id)
	if err != nil {
		http.Error(w, "Failed to fetch media", http.StatusInternalServerError)
		return
	}

	if mediaFile == nil {
		http.NotFound(w, r)
		return
	}

	f, err := os.Open(filepath.Clean(mediaFile.Path))
	if err != nil {
		http.Error(w, "Could Not Opne File", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-type", "video/mp4")
	w.Header().Set("Content-Disposition", "inline; filename=\""+mediaFile.Name+"\"")

	fileInfo, err := f.Stat()
	if err != nil {
		http.Error(w, "Could not get file info", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, mediaFile.Name, fileInfo.ModTime(), f)
}
func (h *Handler) HandleGetMediaInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	mediaFile, err := h.mediaStore.GetMediaByID(ctx, id)
	if err != nil {
		http.Error(w, "Failed to fetch media info", http.StatusInternalServerError)
		return
	}
	if mediaFile == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mediaFile); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
func (h *Handler) HandleThumbnail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	mediaFile, err := h.mediaStore.GetMediaByID(ctx, id)
	if err != nil {
		http.Error(w, "Failed to fetch media", http.StatusInternalServerError)
		return
	}
	if mediaFile == nil {
		http.NotFound(w, r)
		return
	}

	videoPath := filepath.Clean(mediaFile.Path)
	buf := bytes.NewBuffer(nil)

	err = ffmpeg.Input(videoPath).
		Filter("select", ffmpeg.Args{"gte(n\\,150)"}). // wrap in ffmpeg.Args
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "mjpeg"}).
		WithOutput(buf, nil).
		Run()

	if err != nil {
		log.Printf("ffmpeg error generating thumbnail: %v", err)
		http.Error(w, "Failed to generate thumbnail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(buf.Bytes())
}

func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}
