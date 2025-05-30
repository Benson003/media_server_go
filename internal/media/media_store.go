package media

import (
	"context"
	"database/sql"

	"github.com/Benson003/media_server/internal/db"
)

type MediaStore struct {
	db *db.Queries
}
type InsertMediaParams struct {
	ID   string
	Name string
	Path string
	Ext  string
}

func NewMediaStore(dbConn *sql.DB) *MediaStore {
	return &MediaStore{db: db.New(dbConn)}
}
func (m * MediaStore)GetAllMedia(ctx context.Context)([]MediaFile,error){
	dbMedia,err := m.db.GetAllMedia(ctx)
	if err != nil {
		return nil,err
	}
	mediaFiles := make([]MediaFile,len(dbMedia))
	for i,dm := range dbMedia{
		mediaFiles[i] = fromDB(dm)
	}
	return mediaFiles,nil
}

func fromDB(m db.Medium) MediaFile{
	ext := ""
	if m.Ext.Valid{
		ext = m.Ext.String
	}
	return MediaFile{
		ID: m.ID,
		Name: m.Name,
		Path: m.Path,
		Ext: ext,
	}
}
func (m *MediaStore) InsertMediaIfNotExists(ctx context.Context, media MediaFile) error {
	// Check if media already exists by ID
	_, err := m.db.GetMediaByID(ctx, media.ID)
	if err == nil {
		// Found it, skip insert
		return nil
	}
	if err != sql.ErrNoRows {
		// Unexpected error
		return err
	}

	// Not found, so insert it
	dbArg := db.InsertMediaParams{
		ID:   media.ID,
		Name: media.Name,
		Path: media.Path,
		Ext:  sql.NullString{String: media.Ext, Valid: media.Ext != ""},
	}

	return m.db.InsertMedia(ctx, dbArg)
}

func (m *MediaStore) GetMediaByID(ctx context.Context, id string) (*MediaFile, error) {
	dbMedia, err := m.db.GetMediaByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	result := fromDB(dbMedia)
	return &result, nil
}
