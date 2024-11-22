package flow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"mycmd/internal/flow/models"
	"mycmd/pkg/config"
	"mycmd/pkg/logger"
)

type todoArchiveOptions struct {
	todoType string
	date     string
}

func NewTodoArchiveCmd() *cobra.Command {
	opts := &todoArchiveOptions{}

	cmd := &cobra.Command{
		Use:   "todo-archive",
		Short: "归档指定日期范围内的 todo 项目",
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run()
		},
	}

	cmd.Flags().StringVar(&opts.todoType, "type", "", "todo 类型 (work)")
	cmd.Flags().StringVar(&opts.date, "date", "", "归档日期范围，格式：MM/DD,MM/DD")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("date")

	return cmd
}

func (o *todoArchiveOptions) run() error {
	dates := strings.Split(o.date, ",")
	if len(dates) != 2 {
		return fmt.Errorf("日期格式错误，应为: MM/DD,MM/DD")
	}
	startDate, endDate := dates[0], dates[1]

	archiveStartDate := strings.ReplaceAll(startDate, "/", "-")
	archiveEndDate := strings.ReplaceAll(endDate, "/", "-")

	todoDir := config.Get().Flow.TodoDir
	todoFile := filepath.Join(todoDir, o.todoType, fmt.Sprintf("%s.todo", o.todoType))
	archiveFile := filepath.Join(todoDir, o.todoType, fmt.Sprintf("%s(%s~%s).archive",
		o.todoType, archiveStartDate, archiveEndDate))

	if err := os.MkdirAll(filepath.Dir(archiveFile), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 处理 todo 文件
	tasks, err := o.processTodoFile(todoFile, startDate, endDate)
	if err != nil {
		return err
	}

	// 生成归档内容
	content := o.generateArchiveContent(tasks)

	// 写入归档文件
	if err := os.WriteFile(archiveFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入归档文件失败: %w", err)
	}

	logger.Success("已成功创建归档文件: %s", archiveFile)
	return nil
}

func (o *todoArchiveOptions) processTodoFile(todoFile string, startDate, endDate string) ([]models.TaskInfo, error) {
	// logger.Debug("开始处理 todo 文件: %s", todoFile)
	// logger.Debug("归档日期范围: %s ~ %s", startDate, endDate)

	// file, err := os.Open(todoFile)
	// if err != nil {
	// 	return nil, fmt.Errorf("打开 todo 文件失败: %w", err)
	// }
	// defer file.Close()

	// var tasks []taskInfo
	// scanner := bufio.NewScanner(file)
	// inArchive := false
	// lineNum := 0

	// for scanner.Scan() {
	// 	lineNum++
	// 	line := scanner.Text()
	// 	line = strings.TrimSpace(line)

	// 	logger.Debug("\n--- 处理第 %d 行 ---", lineNum)
	// 	logger.Debug("原始内容: %s", line)

	// 	if line == "Archive:" {
	// 		inArchive = true
	// 		logger.Debug("进入归档区域")
	// 		continue
	// 	} else if line != "" && !strings.HasPrefix(line, " ") && inArchive {
	// 		inArchive = false
	// 		logger.Debug("离开归档区域")
	// 	}

	// 	if (inArchive || strings.Contains(line, "@started")) &&
	// 		line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "//") {
	// 		logger.Debug("发现任务行")
	// 		if task := o.parseTaskLine(line); task != nil {
	// 			logger.Debug("解析结果: 状态=%s, 开始时间=%s, 结束时间=%s, 分类=%s, 项目=%s",
	// 				task.status, task.startDate, task.endDate, task.category, task.project)

	// 			if o.isDateInRange(task.startDate, startDate, endDate) {
	// 				logger.Debug("✓ 任务在日期范围内，添加到归档列表")
	// 				tasks = append(tasks, *task)
	// 			} else {
	// 				logger.Debug("✗ 任务不在日期范围内，跳过")
	// 			}
	// 		} else {
	// 			logger.Debug("任务解析失败，跳过")
	// 		}
	// 	} else {
	// 		logger.Debug("跳过非任务行")
	// 	}
	// }

	// logger.Debug("\n总结: 共处理 %d 行，找到 %d 个符合条件的任务", lineNum, len(tasks))
	// return tasks, scanner.Err()

	return nil, nil
}

func (o *todoArchiveOptions) parseTaskLine(line string) *models.TaskInfo {
	task := &models.TaskInfo{}

	line = strings.TrimSpace(line)
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}

	// 1. get status
	symbol := fields[0]
	var startedBySymbol bool
	task.Status, startedBySymbol = models.SymbolSet[symbol]
	if !startedBySymbol {
		return nil
	}

	// range fields[1:]
	for _, field := range fields[1:] {
		if !strings.HasPrefix(field, "@") {
			continue
		}

		// 截取 @ 到 ( 前一个字符的内容 （包含 @，如果没有 (，则为整个字符串）
		// 例如: @project(BUGFIX.BCS) -> @project
		// 例如: @created -> @created
		tagName := field
		if idx := strings.Index(field, "("); idx != -1 {
			tagName = field[:idx]
		}

		parseFn := models.TagParserFns[models.TagSet[tagName]]
		if parseFn == nil {
			logger.Warning("未找到 tag %s 的解析函数", tagName)
			continue
		}

		if err := parseFn(field, task); err != nil {
			logger.Warning("解析 tag %s 失败: %v", tagName, err)
			continue
		}

	}

	return task
}

func (o *todoArchiveOptions) isDateInRange(targetDate, startDate, endDate string) bool {
	logger.Debug("检查日期范围: 目标=%s, 开始=%s, 结束=%s", targetDate, startDate, endDate)

	if targetDate == "" || targetDate == "unknown" {
		logger.Debug("目标日期无效")
		return false
	}

	parseDate := func(date string) time.Time {
		currentYear := time.Now().Year()
		t, _ := time.Parse("2006/01/02", fmt.Sprintf("%d/%s", currentYear, date))
		return t
	}

	target := parseDate(targetDate)
	start := parseDate(startDate)
	end := parseDate(endDate)

	logger.Debug("解析后的日期: 目标=%v, 开始=%v, 结束=%v", target, start, end)

	// 处理跨年的情况
	if end.Before(start) {
		if target.Before(start) {
			target = target.AddDate(1, 0, 0)
			logger.Debug("跨年处理: 调整目标日期 +1 年")
		}
		end = end.AddDate(1, 0, 0)
		logger.Debug("跨年处理: 调整结束日期 +1 年")
	}

	result := !target.Before(start) && !target.After(end)
	logger.Debug("日期范围检查结果: %v", result)
	return result
}

func (o *todoArchiveOptions) generateArchiveContent(tasks []models.TaskInfo) string {
	// var content strings.Builder

	// // 格式1
	// content.WriteString("---------------------------------------------\n")
	// content.WriteString("format1. 状态-开始时间-结束时间-分类-项目-名称\n")
	// content.WriteString("---------------------------------------------\n\n")

	// for _, task := range tasks {
	// 	content.WriteString(fmt.Sprintf("%s-%s-%s-%s-%s-%s\n",
	// 		task.status, task.startDate, task.endDate,
	// 		task.category, task.project, task.name))
	// }

	// // 格式2
	// content.WriteString("\n\n---------------------------------------------\n")
	// content.WriteString("format2. (把已完成和进行中的任务按照分类罗列)\n")
	// content.WriteString("---------------------------------------------\n")

	// // 按分类组织任务
	// categoryTasks := make(map[string][]string)
	// for _, task := range tasks {
	// 	if task.status != "已取消" {
	// 		categoryTasks[task.category] = append(
	// 			categoryTasks[task.category],
	// 			fmt.Sprintf("%s-%s", task.project, task.name),
	// 		)
	// 	}
	// }

	// // 按分类名称排序
	// categories := make([]string, 0, len(categoryTasks))
	// for category := range categoryTasks {
	// 	categories = append(categories, category)
	// }
	// sort.Strings(categories)

	// for _, category := range categories {
	// 	content.WriteString(fmt.Sprintf("\n%s:\n", category))
	// 	for i, task := range categoryTasks[category] {
	// 		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, task))
	// 	}
	// }

	// return content.String()

	return ""
}
