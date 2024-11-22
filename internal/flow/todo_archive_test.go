package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"mycmd/internal/flow/models"
)

func TestTodoArchiveOptions_parseTaskLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected *models.TaskInfo
	}{
		{
			name: "完整的已完成任务",
			line: "✓ 完成功能开发 @project(FEATURE.BCS) @created(23-11-20 10:00) @started(23-11-20 10:00) @done(23-11-21 18:00) @lasted(1d8h) @est(1d)",
			expected: &models.TaskInfo{
				Status: "已完成",
				StartDate: &models.TaskTime{
					Year:  23,
					Month: 11,
					Day:   20,
					Hour:  10,
					Min:   0,
				},
				EndDate: &models.TaskTime{
					Year:  23,
					Month: 11,
					Day:   21,
					Hour:  18,
					Min:   0,
				},
				Category: "FEATURE",
				Project:  "BCS",
				Name:     "完成功能开发",
			},
		},
		{
			name: "已取消的任务",
			line: "✘ 取消的任务 @project(BUG.BCS) @created(23-11-20 10:00) @cancelled(23-11-20 11:00)",
			expected: &models.TaskInfo{
				Status: "已取消",
				StartDate: &models.TaskTime{
					Year:  23,
					Month: 11,
					Day:   20,
					Hour:  10,
					Min:   0,
				},
				EndDate: &models.TaskTime{
					Year:  23,
					Month: 11,
					Day:   20,
					Hour:  11,
					Min:   0,
				},
				Category: "BUG",
				Project:  "BCS",
				Name:     "取消的任务",
			},
		},
		{
			name: "进行中的任务",
			line: "☐ 正在进行的任务 @project(FEATURE.BCS) @created(23-11-20 10:00) @started(23-11-20 10:00)",
			expected: &models.TaskInfo{
				Status: "进行中",
				StartDate: &models.TaskTime{
					Year:  23,
					Month: 11,
					Day:   20,
					Hour:  10,
					Min:   0,
				},
				EndDate:  &models.TaskTime{},
				Category: "FEATURE",
				Project:  "BCS",
				Name:     "正在进行的任务",
			},
		},
		{
			name: "带进度的任务",
			line: "☐ 带进度的任务 @project(FEATURE.BCS) @created(23-11-20 10:00) @started(23-11-20 10:00) @progress(50%)",
			expected: &models.TaskInfo{
				Status: "进行中",
				StartDate: &models.TaskTime{
					Year:  23,
					Month: 11,
					Day:   20,
					Hour:  10,
					Min:   0,
				},
				EndDate:  &models.TaskTime{},
				Category: "FEATURE",
				Project:  "BCS",
				Name:     "带进度的任务",
				Percent:  50,
			},
		},
		{
			name:     "无效的任务行",
			line:     "",
			expected: nil,
		},
		{
			name:     "注释行",
			line:     "// 这是一个注释",
			expected: nil,
		},
		{
			name: "缺少必要标签的任务",
			line: "☐ 缺少项目标签的任务 @created(23-11-20 10:00)",
			expected: &models.TaskInfo{
				Status:    "进行中",
				StartDate: &models.TaskTime{},
				EndDate:   &models.TaskTime{},
				Name:      "缺少项目标签的任务",
			},
		},
	}

	opts := &todoArchiveOptions{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opts.parseTaskLine(tt.line)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.Category, result.Category)
			assert.Equal(t, tt.expected.Project, result.Project)
			// assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Percent, result.Percent)

			// 比较时间
			if result.StartDate != nil {
				assert.Equal(t, tt.expected.StartDate.Year, result.StartDate.Year)
				assert.Equal(t, tt.expected.StartDate.Month, result.StartDate.Month)
				assert.Equal(t, tt.expected.StartDate.Day, result.StartDate.Day)
				assert.Equal(t, tt.expected.StartDate.Hour, result.StartDate.Hour)
				assert.Equal(t, tt.expected.StartDate.Min, result.StartDate.Min)
			}

			if result.EndDate != nil {
				assert.Equal(t, tt.expected.EndDate.Year, result.EndDate.Year)
				assert.Equal(t, tt.expected.EndDate.Month, result.EndDate.Month)
				assert.Equal(t, tt.expected.EndDate.Day, result.EndDate.Day)
				assert.Equal(t, tt.expected.EndDate.Hour, result.EndDate.Hour)
				assert.Equal(t, tt.expected.EndDate.Min, result.EndDate.Min)
			}
		})
	}
}
