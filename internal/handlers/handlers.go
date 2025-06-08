package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	database "media_server/internal/db"
	"media_server/internal/logger"
	"media_server/internal/media"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	DB     *database.DBObject
	Logger *zap.Logger
}

type PaginatedResponse struct {
	Items            []database.MediaItem `json:"items"`
	NumberOfElements int                  `json:"number_of_elements"`
	Pages            int                  `json:"pages"`
	Page             int                  `json:"page"`
	Count            int                  `json:"count"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"internal server error"`
}
type WSMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type ConfigFolderPayload struct {
	Folder []string `json:"folder"`
}

// GetAll godoc
// @Summary      Get all media items
// @Description  Retrieves all media items from the database.
// @Tags         media
// @Produce      json
// @Success      200  {array}   database.MediaItem
// @Failure      500  {object}  handlers.ErrorResponse
// @Router       /media/all [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.DB.GetAll()
	if err != nil {
		h.Logger.Error("failed to fetch media", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		h.Logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// GetPaginatedHandler godoc
// @Summary      Get paginated media items
// @Description  Retrieves media items with pagination.
// @Tags         media
// @Produce      json
// @Param        page   query     int  false  "Page number, default 1"
// @Param        count  query     int  false  "Number of items per page, default 10"
// @Success      200    {object}  PaginatedResponse
// @Failure      400    {object}  handlers.ErrorResponse
// @Failure      500    {object}  handlers.ErrorResponse
// @Router       /media/paginated [get]
func (h *Handler) GetPaginatedHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	count := 10

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		} else {
			http.Error(w, "Invalid page parameter", http.StatusBadRequest)
			return
		}
	}

	if c := r.URL.Query().Get("count"); c != "" {
		if parsed, err := strconv.Atoi(c); err == nil && parsed > 0 {
			count = parsed
		} else {
			http.Error(w, "Invalid count parameter", http.StatusBadRequest)
			return
		}
	}

	items, numberOfElements, pages, err := h.DB.GetPaginated(page, count)
	if err != nil {
		h.Logger.Error("Failed to get paginated media", zap.Error(err))
		http.Error(w, "Failed to get media", http.StatusInternalServerError)
		return
	}

	resp := PaginatedResponse{
		Items:            items,
		NumberOfElements: numberOfElements,
		Pages:            pages,
		Page:             page,
		Count:            count,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetByID godoc
// @Summary      Get media item by ID
// @Description  Returns a single media item by its unique ID.
// @Tags         media
// @Produce      json
// @Param        id   path      string  true  "Media Item ID"
// @Success      200  {object}  database.MediaItem
// @Failure      400  {object}  handlers.ErrorResponse
// @Failure      404  {object}  handlers.ErrorResponse
// @Failure      500  {object}  handlers.ErrorResponse
// @Router       /media/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	mediaItem, err := h.DB.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "media item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(mediaItem); err != nil {
		logger.Log().Error("Failed to encode media item response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// StreamMedia godoc
// @Summary      Stream media file by ID
// @Description  Streams the media file to the client supporting range requests.
// @Tags         media
// @Produce      video/mp4
// @Param        id   path      string  true  "Media Item ID"
// @Success      200  {file}    binary
// @Failure      400  {object}  handlers.ErrorResponse
// @Failure      404  {object}  handlers.ErrorResponse
// @Failure      500  {object}  handlers.ErrorResponse
// @Router       /media/{id}/stream [get]
func (h *Handler) StreamMedia(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	mediaItem, err := h.DB.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "media item not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	file, err := os.Open(mediaItem.Path)
	if err != nil {
		http.Error(w, "failed to open media file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		http.Error(w, "failed to get file info", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, mediaItem.Name, fi.ModTime(), file)
}

// ThumbnailHandler godoc
// @Summary      Get thumbnail image for media
// @Description  Extracts and returns a JPEG thumbnail from the media file at 4 seconds.
// @Tags         media
// @Produce      image/jpeg
// @Param        id   path      string  true  "Media Item ID"
// @Success      200  {file}    binary
// @Failure      404  {object}  handlers.ErrorResponse
// @Failure      500  {object}  handlers.ErrorResponse
// @Router       /media/{id}/thumbnail [get]
func (h *Handler) ThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	mediaItem, err := h.DB.GetByID(id)
	if err != nil {
		http.Error(w, "media not found", http.StatusNotFound)
		return
	}

	if _, err := os.Stat(mediaItem.Path); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	imgBytes, err := extractFrameAt4s(mediaItem.Path)
	if err != nil {
		logger.Log().Sugar().Errorf("failed to extract thumbnail: %v \n", err)
		http.Error(w, "failed to generate thumbnail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(imgBytes)
}

func extractFrameAt4s(videoPath string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	err := ffmpeg.Input(videoPath, ffmpeg.KwArgs{"ss": "4"}).
		Output("pipe:", ffmpeg.KwArgs{
			"vframes": "1",
			"format":  "mjpeg",
		}).
		WithOutput(buf, os.Stderr).
		Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg-go error: %w", err)
	}

	return buf.Bytes(), nil
}

var configMutex sync.Mutex

func (h *Handler) MediaConfigWS(conn *websocket.Conn) {
	defer conn.Close()

	send := func(event string, data any) {
		msg := WSMessage{Event: event}
		var err error
		msg.Data, err = json.Marshal(data)
		if err != nil {
			h.Logger.Warn("Failed to marshal WS response", zap.Error(err))
			return
		}
		if err := conn.WriteJSON(msg); err != nil {
			h.Logger.Warn("Failed to send WS response", zap.Error(err))
		}
	}

	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			h.Logger.Warn("WebSocket read error", zap.Error(err))
			break
		}

		configMutex.Lock()
		cfg, err := media.LoadConfig()
		if err != nil {
			configMutex.Unlock()
			send("error", "failed to load config")
			continue
		}

		switch msg.Event {
		case "add_folder":
			var payload struct {
				Folder string `json:"folder"`
			}
			if err := json.Unmarshal(msg.Data, &payload); err != nil {
				send("error", "invalid payload")
				configMutex.Unlock()
				continue
			}
			if contains(cfg.MediaDirs, payload.Folder) {
				send("error", "folder already exists")
				configMutex.Unlock()
				continue
			}
			cfg.MediaDirs = append(cfg.MediaDirs, payload.Folder)

		case "remove_folder":
			var payload struct {
				Folder string `json:"folder"`
			}
			if err := json.Unmarshal(msg.Data, &payload); err != nil {
				send("error", "invalid payload")
				configMutex.Unlock()
				continue
			}
			found := false
			newDirs := []string{}
			for _, f := range cfg.MediaDirs {
				if f == payload.Folder {
					found = true
				} else {
					newDirs = append(newDirs, f)
				}
			}
			if !found {
				send("error", "folder not found")
				configMutex.Unlock()
				continue
			}
			cfg.MediaDirs = newDirs

		case "toggle_stream":
			cfg.StreamOnDemand = !cfg.StreamOnDemand

		case "fetch_config":
			// Just send the current config without changes
			send("config_data", cfg)
			configMutex.Unlock()
			continue

		default:
			send("error", "unknown event")
			configMutex.Unlock()
			continue
		}

		// Save updated config
		if err := saveConfig(cfg); err != nil {
			send("error", "failed to save config")
			configMutex.Unlock()
			continue
		}
		configMutex.Unlock()

		send("config_updated", cfg)
	}
}

func saveConfig(cfg *media.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(media.GetConfigPath(), data, 0644)
}

func contains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
