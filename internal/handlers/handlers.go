package handlers

import (
	"encoding/json"
	"errors"
	database "media_server/internal/db"
	"media_server/internal/logger"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
	Logger *zap.Logger
}

type PaginatedResponse struct {
	Items            []database.MediaItem`json:"items"`
	NumberOfElements int         `json:"number_of_elements"`
	Pages            int         `json:"pages"`
	Page             int         `json:"page"`
	Count            int         `json:"count"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	items, err := database.GetAll(h.DB)
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

func (h *Handler) GetPaginatedHandler(w http.ResponseWriter, r *http.Request) {
	// Default values
	page := 1
	count := 10

	// Parse page query param
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		} else {
			http.Error(w, "Invalid page parameter", http.StatusBadRequest)
			return
		}
	}

	// Parse count query param
	if c := r.URL.Query().Get("count"); c != "" {
		if parsed, err := strconv.Atoi(c); err == nil && parsed > 0 {
			count = parsed
		} else {
			http.Error(w, "Invalid count parameter", http.StatusBadRequest)
			return
		}
	}

	items, numberOfElements, pages, err := database.GetPaginated(h.DB, page, count)
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

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        http.Error(w, "missing id parameter", http.StatusBadRequest)
        return
    }

    mediaItem, err := database.GetByID(h.DB, id)
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
        // Log encoding error and send 500
        logger.Log().Error("Failed to encode media item response", zap.Error(err))
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }
}

func (h *Handler) StreamMedia(w http.ResponseWriter, r *http.Request) {
    // Grab the ID from the route
    id := chi.URLParam(r, "id")
    if id == "" {
        http.Error(w, "missing id parameter", http.StatusBadRequest)
        return
    }

    // Fetch media info from DB using your existing function
    mediaItem, err := database.GetByID(h.DB, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            http.Error(w, "media item not found", http.StatusNotFound)
        } else {
            http.Error(w, "internal server error", http.StatusInternalServerError)
        }
        return
    }

    // Open the actual file using the path stored in the database
    file, err := os.Open(mediaItem.Path)
    if err != nil {
        http.Error(w, "failed to open media file", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    // Get file info for content headers and range support
    fi, err := file.Stat()
    if err != nil {
        http.Error(w, "failed to get file info", http.StatusInternalServerError)
        return
    }

    // This handles Content-Length, Range requests, and streaming
    http.ServeContent(w, r, mediaItem.Name, fi.ModTime(), file)
}

