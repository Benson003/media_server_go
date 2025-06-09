package database

import (
	"fmt"
	"media_server/internal/logger"
	"media_server/internal/media"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MediaItem struct {
	ID   string `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Ext  string `gorm:"default:''" json:"ext"`
}

type DBObject struct {
	DB  *gorm.DB
	Err error
}

func InitDataBase(dbPath string) DBObject {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		logger.Log().Sugar().Errorf("Failed to init db: %v \n", err)
		return DBObject{DB: nil, Err: err}
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Log().Sugar().Errorf("Failed to get generic DB: %v \n", err)
		return DBObject{DB: nil, Err: err}
	}

	// Enable WAL mode to improve concurrency
	_, err = sqlDB.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		logger.Log().Sugar().Warnf("Failed to enable WAL mode: %v\n", err)
	} else {
		logger.Log().Info("SQLite WAL mode enabled")
	}

	logger.Log().Info("Database connection launched")

	err = db.AutoMigrate(&MediaItem{})
	if err != nil {
		logger.Log().Sugar().Errorf("Failed to auto-migrate MediaItem: %v \n", err)
		return DBObject{DB: nil, Err: err}
	}
	return DBObject{DB: db, Err: nil}
}

func (object DBObject) AddMediaItem(item *MediaItem) error {
	result := object.DB.FirstOrCreate(item, MediaItem{ID: item.ID})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		logger.Log().Sugar().Infof("Media %s already exists", item.ID)
	} else {
		logger.Log().Sugar().Infof("Media %s added to DB", item.ID)
	}
	return nil
}

func (object DBObject) GetPaginated(page int, count int) (itemList []MediaItem, numberOfElements int, pages int, err error) {
	offset := (page - 1) * count
	var total_number_of_rows int64
	var items []MediaItem

	if page < 1 {
		return nil, 0, 0, fmt.Errorf("page number can't be less than one\n")
	}
	if count < 1 {
		return nil, 0, 0, fmt.Errorf("count number can't be less than one\n")
	}

	if err := object.DB.Model(&MediaItem{}).Count(&total_number_of_rows).Error; err != nil {
		return nil, 0, 0, err
	}
	page_no := (total_number_of_rows + int64(count) - 1) / int64(count)

	if err := object.DB.Limit(count).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, 0, err
	}
	return items, len(items), int(page_no), nil
}

func (object DBObject) GetAll() ([]MediaItem, error) {
	var items []MediaItem
	err := object.DB.Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (object DBObject) GetByID(id string) (MediaItem, error) {
	var item MediaItem

	if err := object.DB.Where("id = ?", id).First(&item).Error; err != nil {
		return item, err
	}
	return item, nil
}

func (object DBObject) DeleteAll() error {
	tx := object.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("1 = 1").Delete(&MediaItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}


func (object DBObject) SyncDatabase(mediaFiles *[]media.MediaFile) error {
	const workerCount = 4
	jobs := make(chan media.MediaFile, len(*mediaFiles))
	errChan := make(chan error, len(*mediaFiles))
	var wg sync.WaitGroup

	logger.Log().Sugar().Infof("Starting SyncDatabase with %d media files", len(*mediaFiles))

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			logger.Log().Sugar().Infof("Worker %d started", workerID)
			for media := range jobs {
				logger.Log().Sugar().Infof("Worker %d syncing media: %s", workerID, media.ID)
				err := object.AddMediaItem(&MediaItem{
					ID:   media.ID,
					Name: media.Name,
					Path: media.Path,
					Ext:  media.Ext,
				})
				if err != nil {
					logger.Log().Sugar().Errorf("Worker %d failed to add media %s: %v", workerID, media.ID, err)
					errChan <- fmt.Errorf("failed to add %s: %w", media.ID, err)
				} else {
					logger.Log().Sugar().Infof("Worker %d successfully added media %s", workerID, media.ID)
				}
			}
			logger.Log().Sugar().Infof("Worker %d finished", workerID)
		}(i + 1)
	}

	// Send all media files into the job channel
	for _, media := range *mediaFiles {
		jobs <- media
	}
	close(jobs)
	logger.Log().Sugar().Info("All jobs sent, waiting for workers to finish")

	// Wait for all workers to finish
	wg.Wait()
	close(errChan)
	logger.Log().Sugar().Info("All workers finished")

	// Check for any errors and return the first one found
	for err := range errChan {
		if err != nil {
			logger.Log().Sugar().Errorf("SyncDatabase error encountered: %v", err)
			return err
		}
	}

	logger.Log().Sugar().Info("SyncDatabase completed successfully without errors")
	return nil
}
