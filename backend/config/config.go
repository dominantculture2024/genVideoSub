package config

import (
	"encoding/json"
	"os"
)

// Config 應用程式配置結構
type Config struct {
	Server struct {
		Port string `json:"port"`
		Host string `json:"host"`
	} `json:"server"`
	FalAI struct {
		APIKey  string `json:"api_key"`
		BaseURL string `json:"base_url"`
		Timeout int    `json:"timeout"`
	} `json:"fal_ai"`
	Storage struct {
		TempDir         string `json:"temp_dir"`
		DataDir         string `json:"data_dir"`
		MaxFileSize     int64  `json:"max_file_size"`
		CleanupInterval int    `json:"cleanup_interval"`
	} `json:"storage"`
	Worker struct {
		MaxWorkers int `json:"max_workers"`
		QueueSize  int `json:"queue_size"`
	} `json:"worker"`
}

// LoadConfig 載入配置文件
func LoadConfig(path string) (*Config, error) {
	config := &Config{}
	
	// 設置默認值
	config.Server.Port = "8080"
	config.Server.Host = "localhost"
	config.FalAI.BaseURL = "https://fal.run/fal-ai"
	config.FalAI.Timeout = 300
	config.Storage.TempDir = "temp"
	config.Storage.DataDir = "data"
	config.Storage.MaxFileSize = 104857600 // 100MB
	config.Storage.CleanupInterval = 3600
	config.Worker.MaxWorkers = 5
	config.Worker.QueueSize = 100

	// 如果配置文件存在，則載入
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}