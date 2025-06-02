package database

import (
	"errors"
	"fmt"
	"media_server/internal/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MediaItem struct {
	ID   string `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Ext  string `gorm:"default:''" json:"ext"`
}

func InitDataBase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		logger.Log().Sugar().Errorf("Failed to init db: %v \n", err)
		return nil, err
	}
	logger.Log().Info("Database connection launched")

	err = db.AutoMigrate(&MediaItem{})
	if err != nil {
		logger.Log().Sugar().Errorf("Failed to auto-migrate MediaItem: %v \n", err)
		return nil, err
	}
	return db, nil
}

func AddMediaItem(db *gorm.DB, item *MediaItem) error {
	var exsting MediaItem
	err := db.Where("id = ?", item.ID).First(&exsting).Error
	if err == nil {
		logger.Log().Sugar().Errorf("media already exists %v \n", err)
		return fmt.Errorf("media already exists: %v \n", err)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	tx := db.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(item).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func GetPaginated(db *gorm.DB, page int, count int) (itemList []MediaItem, numberOfElements int, pages int, err error) {
	offset := (page - 1) * count
	var total_number_of_rows int64
	var items []MediaItem

	if page < 1 {
		return nil, 0, 0, fmt.Errorf("page number can't be less than one\n")
	}
	if count < 1 {
		return nil, 0, 0, fmt.Errorf("count number can't be less than one\n")
	}

	if err := db.Model(&MediaItem{}).Count(&total_number_of_rows).Error; err != nil {
		return nil, 0, 0, err
	}
	page_no := (total_number_of_rows + int64(count) - 1) / int64(count)

	if err := db.Limit(count).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, 0, err
	}
	return items, len(items), int(page_no), nil
}

func GetAll(db *gorm.DB) ([]MediaItem, error) {
	var items []MediaItem
	err := db.Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func GetByID(db *gorm.DB, id string) (MediaItem, error) {
	var item MediaItem

	if err := db.Where("id = ?", id).First(&item).Error; err != nil {
		return item, err
	}
	return item, nil
}

func DeleteAll(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("1 = 1").Delete(&MediaItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
