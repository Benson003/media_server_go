package media

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type Config struct {
	MediaDirs           []string `json:"media_dirs"`
	SupportedExtensions []string `json:"supported_extensions"`
	StreamOnDemand      bool     `json:"on_demand"`
	AllowedOrigins      []string `json:"allowed_origins"`
}

type MediaFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Ext  string `json:"ext"`
}

var configLock sync.Mutex
var configPath string

func SetConfigPath(path string) {
	configPath = path
}
func GetConfigPath() (path string) {
	return configPath
}

func LoadConfig() (*Config, error) {
	if configPath == "" {
		return nil, errors.New("config path not set")
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config Config

	if erro := json.Unmarshal(data, &config); erro != nil {
		return nil, erro
	}
	return &config, nil
}

func (config *Config) GetAlowedOrigns() (alllowed_origins []string, err error) {
	if config.AllowedOrigins != nil {
		return config.AllowedOrigins, nil
	}
	return nil, fmt.Errorf("Config not found")
}

func (config *Config) ScanMediaDirs() ([]MediaFile, error) {
	var files []MediaFile
	for _, dir := range config.MediaDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if contains(config.SupportedExtensions, ext) {
				id := hashFilePath(path)
				files = append(files, MediaFile{
					ID:   id,
					Name: info.Name(),
					Path: path,
					Ext:  ext,
				})

			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func contains(list []string, target string) bool {
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

func (config *Config) SaveConfig() error {
	configLock.Lock()
	defer configLock.Unlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
func (config *Config) AddMediaDir(newDir string) error {
	configLock.Lock()
	defer configLock.Unlock()

	// Normalize path (optional but good)
	newDir = filepath.Clean(newDir)

	// Check if already exists
	for _, d := range config.MediaDirs {
		if d == newDir {
			return errors.New("folder already exists")
		}
	}

	config.MediaDirs = append(config.MediaDirs, newDir)
	return config.SaveConfig()
}
func (config *Config) RemoveMediaDir(dirToRemove string) error {
	configLock.Lock()
	defer configLock.Unlock()

	dirToRemove = filepath.Clean(dirToRemove)

	newDirs := make([]string, 0, len(config.MediaDirs))
	found := false
	for _, d := range config.MediaDirs {
		if d == dirToRemove {
			found = true
			continue
		}
		newDirs = append(newDirs, d)
	}
	if !found {
		return errors.New("folder not found")
	}

	config.MediaDirs = newDirs
	return config.SaveConfig()
}
func (config *Config) ToggleStreamOnDemand() error {
	configLock.Lock()
	defer configLock.Unlock()

	config.StreamOnDemand = !config.StreamOnDemand
	return config.SaveConfig()
}

// FetchConfig reloads the config from file, useful for syncing changes
func FetchConfig() (*Config, error) {
	return LoadConfig()
}

func ExtractFrameAt(videoPath string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	err := ffmpeg_go.Input(videoPath, ffmpeg_go.KwArgs{"ss": "7"}).
		Output("pipe:", ffmpeg_go.KwArgs{
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
