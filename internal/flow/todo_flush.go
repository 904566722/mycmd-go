package flow

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mycmd/pkg/config"
	"mycmd/pkg/logger"
)

type todoFlushOptions struct {
	todoType string
	projects string
}

func NewTodoFlushCmd() *cobra.Command {
	opts := &todoFlushOptions{}

	cmd := &cobra.Command{
		Use:   "todo-flush",
		Short: "初始化或刷新 todo 文件",
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run()
		},
	}

	cmd.Flags().StringVar(&opts.todoType, "type", "", "todo 类型 (work)")
	cmd.Flags().StringVar(&opts.projects, "project", "", "项目名称列表，用逗号分隔")
	cmd.MarkFlagRequired("type")

	return cmd
}

func (o *todoFlushOptions) run() error {
	if o.todoType == "work" && o.projects == "" {
		return fmt.Errorf("work 类型必须指定 --project 参数")
	}

	todoDir := config.Get().Flow.TodoDir
	templateFile := filepath.Join(todoDir, o.todoType, fmt.Sprintf("%s-template.todo", o.todoType))
	targetFile := filepath.Join(todoDir, o.todoType, fmt.Sprintf("%s.todo", o.todoType))

	// 检查模板文件是否存在
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		return fmt.Errorf("模板文件不存在: %s", templateFile)
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(targetFile); err == nil {
		fmt.Print("文件已存在，是否覆盖（覆盖前请确保已经归档）？(y/n) ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			logger.Info("操作已取消")
			return nil
		}
	}

	// 读取模板文件
	content, err := o.processTemplateFile(templateFile)
	if err != nil {
		return err
	}

	// 写入目标文件
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	logger.Success("已成功创建 todo 文件: %s", targetFile)
	if o.todoType == "work" {
		logger.Success("已添加项目: %s", o.projects)
	}

	return nil
}

func (o *todoFlushOptions) processTemplateFile(templateFile string) (string, error) {
	file, err := os.Open(templateFile)
	if err != nil {
		return "", fmt.Errorf("打开模板文件失败: %w", err)
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)
	inCategory := false

	for scanner.Scan() {
		line := scanner.Text()

		// 检查是否是根分类行
		if line != "" && !strings.HasPrefix(line, " ") &&
			!strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "//") {
			inCategory = true
			// 确保根分类有冒号结尾
			if !strings.HasSuffix(line, ":") {
				line += ":"
			}
			result.WriteString(line + "\n")

			// 如果是 work 类型，添加项目
			if o.todoType == "work" {
				projects := strings.Split(o.projects, ",")
				for _, project := range projects {
					project = strings.TrimSpace(project)
					if !strings.HasSuffix(project, ":") {
						project += ":"
					}
					result.WriteString("    " + project + "\n")
				}
			}
		} else if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") || line == "" {
			// 注释行或空行
			result.WriteString(line + "\n")
			inCategory = false
		} else if !inCategory {
			// 非分类内容
			result.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取模板文件失败: %w", err)
	}

	return result.String(), nil
}
