package main

import (
	"fmt"
	"os"

	"mycmd/cmd"
	"mycmd/pkg/initialize"
)

func main() {
	// 初始化
	if err := initialize.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 执行命令
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
} 