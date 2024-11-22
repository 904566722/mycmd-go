package config

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"

	"mycmd/pkg/logger"
)

// Config 存储所有配置信息
type Config struct {
	Base struct {
		ConfigPath string `yaml:"config_path" json:"config_path"`
	} `yaml:"base" json:"base"`
	Flow struct {
		TodoDir string `yaml:"todo_dir" json:"todo_dir"`
	} `yaml:"flow" json:"flow"`
}

var GlobalConfig Config

// LoadConfig 从 YAML 文件加载配置
func LoadConfig(configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &GlobalConfig)
	if err != nil {
		return err
	}

	// 打印配置信息
	printConfig()

	return nil
}

// GetConfigPath 返回配置文件路径
func GetConfigPath() string {
	if GlobalConfig.Base.ConfigPath != "" {
		return GlobalConfig.Base.ConfigPath
	}
	return "./configs" // 默认配置路径
}

// Get 返回全局配置
func Get() Config {
	return GlobalConfig
}

// printConfig 以格式化的方式打印配置信息
func printConfig() {
	jsonData, err := json.MarshalIndent(GlobalConfig, "", "  ")
	if err != nil {
		logger.Debug("打印配置信息失败: %v", err)
		return
	}

	logger.Debug("加载配置信息:\n%s", string(jsonData))
}
