package media

import (
	"fmt"
	"media_server/internal/logger"
	"os"
	"path/filepath"
)

const cacheFolder string = "thumbnail_cache"

func PreGenerateThumbnail(m *MediaFile) error {
	if err := preCheckForFolder(); err != nil {
		return fmt.Errorf("check failed: %v", err)
	}
	logger.Log().Info("Thumbnail cache folder ready")

	image,err := ExtractFrameAt(m.Path)
	if(err != nil){
		return fmt.Errorf("error creating image: %v",err)
	}
	filename := filepath.Join(cacheFolder, m.ID + ".jpg")
	if err := saveImageToFile(filename,image); err != nil{
		return fmt.Errorf("failed to create thumbnail: %v",err)
	}
	return nil
}
func FetchPreGeneratedThumbnails(){
	
}

func saveImageToFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0644) // 0644 = -rw-r--r--
	if err != nil {
		return fmt.Errorf("failed to save image to file: %w", err)
	}
	return nil
}



func preCheckForFolder() error {
	if _, err := os.Stat(cacheFolder); os.IsNotExist(err) {
		// Doesn't exist — try to create
		if err := os.Mkdir(cacheFolder, 0750); err != nil {
			return fmt.Errorf("failed to create folder: %v", err)
		}
	}
	return nil // Either already exists or created successfully
}
