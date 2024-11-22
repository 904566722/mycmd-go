package initialize

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"mycmd/pkg/config"
)

var (
	configFile string
)

// Init 初始化所有模块
func Init() error {
	// 解析命令行参数
	parseFlags()

	// 初始化配置
	if err := initConfig(); err != nil {
		return fmt.Errorf("初始化配置失败: %w", err)
	}

	return nil
}

// parseFlags 解析命令行参数
func parseFlags() {
	flag.StringVar(&configFile, "config", "config.yaml", "配置文件路径")
	flag.Parse()
}

// initConfig 初始化配置
func initConfig() error {
	// 获取配置文件的绝对路径
	absPath, err := filepath.Abs(configFile)
	if err != nil {
		return fmt.Errorf("获取配置文件绝对路径失败: %w", err)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", absPath)
	}

	// 加载配置文件
	if err := config.LoadConfig(absPath); err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}

	return nil
} 