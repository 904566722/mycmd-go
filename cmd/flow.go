package cmd

import (
	"github.com/spf13/cobra"
	"mycmd/internal/flow"
)

var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "管理工作流和学习流",
	Long: `flow 模块用于管理工作流相关的功能，主要包括 todo 文件的管理。
例如：初始化或刷新 todo 文件，归档指定日期范围内的 todo 项目等。`,
}

func init() {
	// 注册 flow 子命令
	flowCmd.AddCommand(
		flow.NewTodoFlushCmd(),
		flow.NewTodoArchiveCmd(),
	)
} 