package flow

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	logger.Debug("开始处理 todo 文件: %s", todoFile)
	logger.Info("归档日期范围: %s ~ %s", startDate, endDate)

	file, err := os.Open(todoFile)
	if err != nil {
		return nil, fmt.Errorf("打开 todo 文件失败: %w", err)
	}
	defer file.Close()

	var tasks []models.TaskInfo
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		line = strings.TrimSpace(line)

		task := o.parseTaskLine(line)
		if task == nil {
			continue
		}

		if o.isDateInRange(startDate, endDate, task.StartDate, task.EndDate) {
			logger.Success("找到符合条件的任务: %s", task.Name)
			tasks = append(tasks, *task)
		} else {
			logger.Warning("任务 %s 不在日期范围内", task.Name)
		}
	}

	logger.Debug("\n总结: 共处理 %d 行，找到 %d 个符合条件的任务", lineNum, len(tasks))
	return tasks, scanner.Err()
}

// ✔ mock-duale @done(24-11-21 15:41) @project(REFACTOR.DUALENGINE)
// 当第一个 @ 出现之后，后面所有内容都是 tag，应该拆分每个 @ 内容，而不是按照空格拆分
func (o *todoArchiveOptions) parseTaskLine(line string) *models.TaskInfo {
	task := &models.TaskInfo{}

	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// 检查每个已知的符号
	var matchedSymbol string
	var status models.TaskStatus
	var startedBySymbol bool

	// 先尝试匹配最长的符号
	for symbol := range models.SymbolSet {
		if strings.HasPrefix(line, symbol) {
			if len(symbol) > len(matchedSymbol) {
				matchedSymbol = symbol
				status, startedBySymbol = models.SymbolSet[symbol]
			}
		}
	}

	if !startedBySymbol {
		return nil
	}

	logger.Info("开始解析任务行: %s", line)

	task.Status = status

	// 去掉状态符号
	line = strings.TrimPrefix(line, matchedSymbol)
	line = strings.TrimSpace(line)

	// 找到第一个 @ 的位置
	firstAtIndex := strings.Index(line, "@")
	if firstAtIndex == -1 {
		// 没有标签，整行都是任务名
		task.Name = line
		return task
	}

	// 提取任务名称和标签部分
	task.Name = strings.TrimSpace(line[:firstAtIndex])
	tagsPart := line[firstAtIndex:]

	// 按 @ 分割标签
	tags := strings.Split(tagsPart, "@")
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		tag = "@" + tag

		// 截取标签名
		tagName := tag
		if idx := strings.Index(tag, "("); idx != -1 {
			tagName = tag[:idx]
		}

		parseFn := models.TagParserFns[models.TagSet[tagName]]
		if parseFn == nil {
			continue
		}

		tag = strings.TrimSpace(tag)
		if err := parseFn(tag, task); err != nil {
			logger.Warning("解析 tag %s 失败: %v", tagName, err)
			continue
		}
	}

	return task
}

// isDateInRange 检查任务时间范围与日期范围是否存在交集
func (o *todoArchiveOptions) isDateInRange(rangeStart, rangeEnd string, taskStartTime, taskEndTime *models.TaskTime) bool {
	// 如果任务没有开始时间和结束时间，直接返回 false
	if taskStartTime == nil && taskEndTime == nil {
		return false
	}

	// 将 TaskTime 转换为 time.Time
	parseTaskTime := func(t *models.TaskTime) time.Time {
		return time.Date(2000+t.Year, time.Month(t.Month), t.Day, 0, 0, 0, 0, time.Local)
	}

	// 将日期字符串转换为 time.Time
	parseRangeDate := func(date string) time.Time {
		parts := strings.Split(date, "/")
		if len(parts) != 2 {
			logger.Warning("无效的日期格式: %s", date)
			return time.Time{}
		}
		month, _ := strconv.Atoi(parts[0])
		day, _ := strconv.Atoi(parts[1])
		currentYear := time.Now().Year() - 2000 // 转换为两位数年份
		return time.Date(2000+currentYear, time.Month(month), day, 0, 0, 0, 0, time.Local)
	}

	rangeStartTime := parseRangeDate(rangeStart)
	rangeEndTime := parseRangeDate(rangeEnd)

	// 处理跨年的情况
	if rangeEndTime.Before(rangeStartTime) {
		rangeEndTime = rangeEndTime.AddDate(1, 0, 0)
		logger.Debug("跨年处理: 调整范围结束时间 +1 年")
	}

	// 如果只有结束时间，且结束时间在范围内，则符合条件
	if taskStartTime == nil && taskEndTime != nil {
		taskEnd := parseTaskTime(taskEndTime)
		result := !taskEnd.Before(rangeStartTime) && !taskEnd.After(rangeEndTime)
		return result
	}

	// 正常情况：有开始时间
	taskStart := parseTaskTime(taskStartTime)
	var taskEnd time.Time
	if taskEndTime != nil {
		taskEnd = parseTaskTime(taskEndTime)
	} else {
		taskEnd = time.Now() // 如果没有结束时间，表示至今
	}

	// 判断时间范围是否有重叠
	// 两个时间范围有重叠的条件是:
	// !(任务结束 < 范围开始 || 任务开始 > 范围结束)
	result := !(taskEnd.Before(rangeStartTime) || taskStart.After(rangeEndTime))
	logger.Debug("日期范围检查结果: %v (范围: %s ~ %s, 任务: %s ~ %v)", result, rangeStart, rangeEnd, taskStart.Format("01/02"), taskEnd.Format("01/02"))
	return result
}

func (o *todoArchiveOptions) generateArchiveContent(tasks []models.TaskInfo) string {
	var content strings.Builder

	// 格式1
	content.WriteString("---------------------------------------------\n")
	content.WriteString("format1. 状态-开始时间-结束时间-分类-项目-名称\n")
	content.WriteString("---------------------------------------------\n\n")

	for _, task := range tasks {
		content.WriteString(task.String() + "\n")
	}

	// 格式2
	content.WriteString("\n\n---------------------------------------------\n")
	content.WriteString("format2. (把已完成和进行中的任务按照分类罗列)\n")
	content.WriteString("---------------------------------------------\n")

	// 按分类组织任务
	categoryTasks := make(map[string][]string)
	for _, task := range tasks {
		task := &task
		category := task.Category
		if category == "" {
			category = "OTHER"
		}
		if task.Status == models.TaskStatusDone {
			task = task.IgnoreStatus()
		}
		s := task.IgnoreCategory().String()

		if task.Status != models.TaskStatusCancel {
			categoryTasks[category] = append(
				categoryTasks[category],
				s,
			)
		}
	}

	// 按分类名称排序,把 OTHER 放到最后
	categories := make([]string, 0, len(categoryTasks))
	hasOther := false
	for category := range categoryTasks {
		if category == "OTHER" {
			hasOther = true
			continue
		}
		categories = append(categories, category)
	}
	sort.Strings(categories)
	if hasOther {
		categories = append(categories, "OTHER")
	}

	// 同时写入文件内容和打印日志
	for _, category := range categories {
		content.WriteString(fmt.Sprintf("\n%s:\n", category))
		logger.Info("\n%s:", category)
		
		for i, task := range categoryTasks[category] {
			content.WriteString(fmt.Sprintf("%d. %s\n", i+1, task))
			logger.Info("%d. %s", i+1, task)
		}
	}

	return content.String()
}
