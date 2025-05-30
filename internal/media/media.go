package media

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	MediaDirs           []string `json:"media_dirs"`
	SupportedExtensions []string `json:"supported_extensions"`
}

type MediaFile struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Ext string `json:"ext"`
}


func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil{
		return nil,err
	}
	var config Config

	if erro := json.Unmarshal(data, &config); erro != nil {
		return nil, erro
	}
	return &config,nil
}


func ScanMediaDirs(config Config)([]MediaFile,error)  {
	var files []MediaFile
	for _ , dir := range config.MediaDirs{
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if contains(config.SupportedExtensions,ext){
				id := hashFilePath(path)
				files = append(files, MediaFile{
					ID: id,
					Name: info.Name(),
					Path: path,
					Ext: ext,
				})

			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files,nil
}

func contains(list []string,target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func hashFilePath(path string) string {
	hash := sha1.Sum([]byte(path))
	return fmt.Sprintf("%x", hash)
}
