package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"mycmd/pkg/logger"
)

type TaskInfo struct {
	Status    TaskStatus // 已完成、进行中、已取消
	StartDate *TaskTime
	EndDate   *TaskTime
	Category  string // 分类 （todo 文件的根分类）
	Project   string // 项目 （分类下的子分类）
	Name      string // 名称
	Percent   int    // 百分比 0-100
}

type TaskStatus string

const (
	TaskStatusInProgress TaskStatus = "进行中"
	TaskStatusDone       TaskStatus = "已完成"
	TaskStatusCancel     TaskStatus = "已取消"
)

func (t *TaskInfo) String() string {
	res := strings.Builder{}
	res.WriteString(string(t.Status))

	if t.Status == TaskStatusInProgress {
		res.WriteString(fmt.Sprintf("(%d%%)", t.Percent))
	}

	dateRange := ""
	if t.StartDate != nil && t.EndDate != nil {
		if t.StartDate.Year == t.EndDate.Year {
			dateRange = t.StartDate.MMDD()
		} else {
			dateRange = fmt.Sprintf("%s~%s", t.StartDate.MMDD(), t.EndDate.MMDD())
		}
	} else if t.StartDate != nil {
		dateRange = fmt.Sprintf("%s~至今(%s)", t.StartDate.MMDD(), time.Now().Format("01/02"))
	} else if t.EndDate != nil {
		dateRange = t.EndDate.MMDD()
	}

	if dateRange != "" {
		if res.Len() > 0 {
			res.WriteString("-")
		}
		res.WriteString(dateRange)
	}

	if t.Category != "" {
		if res.Len() > 0 {
			res.WriteString("-")
		}
		res.WriteString(t.Category)
	}

	if t.Project != "" {
		if res.Len() > 0 {
			res.WriteString("-")
		}
		res.WriteString(t.Project)
	}

	if t.Name != "" {
		if res.Len() > 0 {
			res.WriteString("-")
		}
		res.WriteString(t.Name)
	} else {
		res.WriteString("-unknown task name")
	}

	return res.String()
}

func (t *TaskInfo) IgnoreCategory() *TaskInfo {
	t.Category = ""
	return t
}

func (t *TaskInfo) IgnoreStatus() *TaskInfo {
	t.Status = ""
	return t
}

type TaskTime struct {
	Year  int
	Month int
	Day   int
	Hour  int
	Min   int
}

func (t *TaskTime) String() string {
	return fmt.Sprintf("%02d-%02d-%02d %02d:%02d", t.Year, t.Month, t.Day, t.Hour, t.Min)
}

func (t *TaskTime) MMDD() string {
	return fmt.Sprintf("%02d/%02d", t.Month, t.Day)
}

func NewTaskTime(year, month, day, hour, min int) *TaskTime {
	return &TaskTime{
		Year:  year,
		Month: month,
		Day:   day,
		Hour:  hour,
		Min:   min,
	}
}

// 进行中: - ❍ ❑ ■ ⬜ □ ☐ ▪ ▫ – — ≡ → › [] [ ]
// 已完成: ✔ ✓ ☑ + [x] [X] [+]
// 已取消: ✘ x X [-]
var SymbolSet = map[string]TaskStatus{
	"-":   TaskStatusInProgress,
	"❍":   TaskStatusInProgress,
	"❑":   TaskStatusInProgress,
	"■":   TaskStatusInProgress,
	"⬜":   TaskStatusInProgress,
	"□":   TaskStatusInProgress,
	"☐":   TaskStatusInProgress,
	"▪":   TaskStatusInProgress,
	"▫":   TaskStatusInProgress,
	"–":   TaskStatusInProgress,
	"—":   TaskStatusInProgress,
	"≡":   TaskStatusInProgress,
	"→":   TaskStatusInProgress,
	"›":   TaskStatusInProgress,
	"[]":  TaskStatusInProgress,
	"[ ]": TaskStatusInProgress,

	"✔":   TaskStatusDone,
	"✓":   TaskStatusDone,
	"☑":   TaskStatusDone,
	"+":   TaskStatusDone,
	"[x]": TaskStatusDone,
	"[X]": TaskStatusDone,
	"[+]": TaskStatusDone,

	"✘":   TaskStatusCancel,
	"x":   TaskStatusCancel,
	"X":   TaskStatusCancel,
	"[-]": TaskStatusCancel,
}

// @project
// @created
// @started
// @done
// @cancelled
// @lasted
// @progress
type TagType string

const (
	tagTypeProject   TagType = "@project"
	tagTypeCreated   TagType = "@created"
	tagTypeStarted   TagType = "@started"
	tagTypeDone      TagType = "@done"
	tagTypeCancelled TagType = "@cancelled"
	tagTypeLasted    TagType = "@lasted"
	tagTypePercent   TagType = "@progress"
)

var TagSet = map[string]TagType{
	"@project":   tagTypeProject,
	"@created":   tagTypeCreated,
	"@started":   tagTypeStarted,
	"@done":      tagTypeDone,
	"@cancelled": tagTypeCancelled,
	"@lasted":    tagTypeLasted,
	"@progress":  tagTypePercent,
}

var TagParserFns = map[TagType]tagParser{
	tagTypeProject:   parseProject,
	tagTypeStarted:   parseStarted,
	tagTypeDone:      parseDone,
	tagTypeCancelled: parseCancelled,
	tagTypePercent:   parseProgress,
}

type TagParseResult struct {
	TagType TagType
	Content string
}

type tagParser func(tagContent string, task *TaskInfo) error

// @project 的内容可能如：
// @project(BUGFIX.BCS)
var parseProject = tagParser(func(tagContent string, task *TaskInfo) error {
	tagContent = strings.TrimPrefix(tagContent, "@project")

	// 校验 tagContent 格式
	if len(tagContent) == 0 {
		return fmt.Errorf("project tag content cannot be empty")
	}

	// 检查是否以括号包裹
	if !strings.HasPrefix(tagContent, "(") || !strings.HasSuffix(tagContent, ")") {
		return fmt.Errorf("project tag content must be wrapped in parentheses")
	}

	// 提取括号中的内容
	projectContent := strings.TrimSpace(tagContent[1 : len(tagContent)-1])
	if len(projectContent) == 0 {
		return fmt.Errorf("project content cannot be empty")
	}

	// split by "."
	projectParts := strings.Split(projectContent, ".")
	if len(projectParts) == 0 {
		return fmt.Errorf("project content must contain at least one dot")
	}

	task.Category = projectParts[0]
	if len(projectParts) > 1 {
		task.Project = strings.Join(projectParts[1:], ".")
	}

	logger.Debug("解析 project tag %s 成功: 分类=%s, 项目=%s", tagContent, task.Category, task.Project)
	return nil
})

// e.g. @started(24-11-22 14:58)
var parseStarted = tagParser(func(tagContent string, task *TaskInfo) error {
	tagContent = strings.TrimPrefix(tagContent, "@started")

	// 校验 tagContent 格式
	if len(tagContent) == 0 {
		return fmt.Errorf("started tag content cannot be empty")
	}

	// 检查是否以括号包裹
	if !strings.HasPrefix(tagContent, "(") || !strings.HasSuffix(tagContent, ")") {
		return fmt.Errorf("started tag content must be wrapped in parentheses")
	}

	// 提取括号中的内容
	timeContent := strings.TrimSpace(tagContent[1 : len(tagContent)-1])
	if len(timeContent) == 0 {
		return fmt.Errorf("started time content cannot be empty")
	}

	// 解析时间
	startTime, err := parseDatetime(timeContent)
	if err != nil {
		return fmt.Errorf("parse started time failed: %w", err)
	}

	task.StartDate = startTime
	logger.Debug("解析 started tag %s 成功: %v", tagContent, startTime)
	return nil
})

var parseDone = tagParser(func(tagContent string, task *TaskInfo) error {
	tagContent = strings.TrimPrefix(tagContent, "@done")

	// 校验 tagContent 格式
	if len(tagContent) == 0 {
		return fmt.Errorf("done tag content cannot be empty")
	}

	// 检查是否以括号包裹
	if !strings.HasPrefix(tagContent, "(") || !strings.HasSuffix(tagContent, ")") {
		return fmt.Errorf("done tag content must be wrapped in parentheses")
	}

	// 提取括号中的内容
	timeContent := strings.TrimSpace(tagContent[1 : len(tagContent)-1])
	if len(timeContent) == 0 {
		return fmt.Errorf("done time content cannot be empty")
	}

	// 解析时间
	endTime, err := parseDatetime(timeContent)
	if err != nil {
		return fmt.Errorf("parse done time failed: %w", err)
	}

	task.EndDate = endTime
	logger.Debug("解析 done tag %s 成功: %v", tagContent, endTime)

	return nil
})

// content e.g. 24-11-22 14:58
func parseDatetime(content string) (*TaskTime, error) {
	parts := strings.Split(content, " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid datetime format, expect: YY-MM-DD HH:mm")
	}

	dateParts := strings.Split(parts[0], "-")
	if len(dateParts) != 3 {
		return nil, fmt.Errorf("invalid date format, expect: YY-MM-DD")
	}

	timeParts := strings.Split(parts[1], ":")
	if len(timeParts) != 2 {
		return nil, fmt.Errorf("invalid time format, expect: HH:mm")
	}

	year, err := strconv.Atoi(dateParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid year: %w", err)
	}

	month, err := strconv.Atoi(dateParts[1])
	if err != nil || month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month: %w", err)
	}

	day, err := strconv.Atoi(dateParts[2])
	if err != nil || day < 1 || day > 31 {
		return nil, fmt.Errorf("invalid day: %w", err)
	}

	hour, err := strconv.Atoi(timeParts[0])
	if err != nil || hour < 0 || hour > 23 {
		return nil, fmt.Errorf("invalid hour: %w", err)
	}

	min, err := strconv.Atoi(timeParts[1])
	if err != nil || min < 0 || min > 59 {
		return nil, fmt.Errorf("invalid minute: %w", err)
	}

	return &TaskTime{
		Year:  year,
		Month: month,
		Day:   day,
		Hour:  hour,
		Min:   min,
	}, nil
}

func parseCancelled(tagContent string, task *TaskInfo) error {
	task.Status = TaskStatusCancel
	return nil
}

// @progress 的内容可能如：
// @progress(100)
var parseProgress = tagParser(func(tagContent string, task *TaskInfo) error {
	tagContent = strings.TrimPrefix(tagContent, "@progress")

	// 校验 tagContent 格式
	if len(tagContent) == 0 {
		return fmt.Errorf("progress tag content cannot be empty")
	}

	// 检查是否以括号包裹
	if !strings.HasPrefix(tagContent, "(") || !strings.HasSuffix(tagContent, ")") {
		return fmt.Errorf("progress tag content must be wrapped in parentheses")
	}

	// 提取括号中的内容
	timeContent := strings.TrimSpace(tagContent[1 : len(tagContent)-1])
	if len(timeContent) == 0 {
		return fmt.Errorf("progress content cannot be empty")
	}

	// 只保留数字部分
	var buffer strings.Builder
	for _, c := range timeContent {
		if c >= '0' && c <= '9' {
			buffer.WriteRune(c)
		}
	}
	timeContent = buffer.String()

	// 解析百分比
	percent, err := strconv.Atoi(timeContent)
	if err != nil || percent < 0 || percent > 100 {
		return fmt.Errorf("invalid progress: %w", err)
	}

	logger.Debug("解析 progress tag %s 成功: %d", tagContent, percent)
	task.Percent = percent
	return nil
})
