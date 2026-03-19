package repository_test

import (
	"testing"

	"github.com/example/go-api-starter/internal/repository"
)

func TestPagination_Normalize(t *testing.T) {
	tests := []struct {
		name         string
		input        repository.Pagination
		wantPage     int
		wantPageSize int
	}{
		{
			name:         "零值应规范为第1页、每页10条",
			input:        repository.Pagination{},
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name:         "负数页码应规范为1",
			input:        repository.Pagination{Page: -5, PageSize: 20},
			wantPage:     1,
			wantPageSize: 20,
		},
		{
			name:         "page=0应规范为1",
			input:        repository.Pagination{Page: 0, PageSize: 10},
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name:         "负数PageSize应规范为10",
			input:        repository.Pagination{Page: 1, PageSize: -1},
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name:         "超大PageSize应截断为100",
			input:        repository.Pagination{Page: 1, PageSize: 9999},
			wantPage:     1,
			wantPageSize: 100,
		},
		{
			name:         "正常值不应被修改",
			input:        repository.Pagination{Page: 3, PageSize: 25},
			wantPage:     3,
			wantPageSize: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.input
			p.Normalize()
			if p.Page != tt.wantPage {
				t.Errorf("Page: 期望 %d，实际 %d", tt.wantPage, p.Page)
			}
			if p.PageSize != tt.wantPageSize {
				t.Errorf("PageSize: 期望 %d，实际 %d", tt.wantPageSize, p.PageSize)
			}
		})
	}
}

func TestPagination_Offset(t *testing.T) {
	tests := []struct {
		name       string
		input      repository.Pagination
		wantOffset int
	}{
		{name: "page=1 offset=0", input: repository.Pagination{Page: 1, PageSize: 10}, wantOffset: 0},
		{name: "page=2 offset=10", input: repository.Pagination{Page: 2, PageSize: 10}, wantOffset: 10},
		{name: "page=3 pageSize=20 offset=40", input: repository.Pagination{Page: 3, PageSize: 20}, wantOffset: 40},
		// 安全边界：未调用 Normalize() 时 Page<=1 不能返回负数
		{name: "page=0 offset不能为负", input: repository.Pagination{Page: 0, PageSize: 10}, wantOffset: 0},
		{name: "page=-1 offset不能为负", input: repository.Pagination{Page: -1, PageSize: 10}, wantOffset: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Offset()
			if got != tt.wantOffset {
				t.Errorf("Offset(): 期望 %d，实际 %d", tt.wantOffset, got)
			}
			if got < 0 {
				t.Errorf("Offset() 不能返回负数，实际 %d", got)
			}
		})
	}
}

func TestPagination_NormalizeAndOffset_Consistent(t *testing.T) {
	// Normalize 后 Offset 必须与直接计算结果一致
	p := repository.Pagination{Page: 5, PageSize: 15}
	p.Normalize()
	got := p.Offset()
	want := (5 - 1) * 15 // = 60
	if got != want {
		t.Errorf("Normalize 后 Offset(): 期望 %d, 实际 %d", want, got)
	}
}
