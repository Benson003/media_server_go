package media

import (
	"fmt"
	"os"
)

const cache_folder string = "thumbnail_cache"

func PreGenrateThumbnails() error {
	if err := preCheckForFolder(); err != nil {
		return fmt.Errorf("check failed due to: %v", err)
	}
	return nil
}

func preCheckForFolder() error {
	err := os.Mkdir(cache_folder, 0750)
	if err != nil && os.IsExist(err) {
		return fmt.Errorf("failed to create cache folder: %e", err)
	}
	return nil
}
