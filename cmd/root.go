package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mycmd",
	Short: "mycmd - 管理自定义命令的工具",
	Long: `mycmd 是一个命令行工具，用于创建、编辑和管理自己的命令。
每个子命令对应调用不同模块目录的脚本。`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 注册子命令
	rootCmd.AddCommand(flowCmd)
} 