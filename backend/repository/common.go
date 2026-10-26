package repository

import (
	"gorm.io/gorm"
)

// PageQuery 通用分页参数
type PageQuery struct {
	Page     int // 页码
	PageSize int // 每页条数
}

// Normalize 规范化分页参数，设置默认值
func (q *PageQuery) Normalize() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
}

// Offset 计算偏移量
func (q *PageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

// PageResult 分页元数据（不含数据列表）
type PageResult struct {
	TotalPage  int // 总页数
	TotalCount int // 总记录数
	Page       int // 当前页码
	PageSize   int // 每页条数
}

// Paginate 通用分页查询
// 接收已附带 Where 条件的 GORM db 链，自动执行 Count + Offset + Limit + Find。
// result 需传入切片指针（如 &[]po.User{}）。
func Paginate[T any](db *gorm.DB, query *PageQuery, result *[]T) (*PageResult, error) {
	query.Normalize()

	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if err := db.Offset(query.Offset()).Limit(query.PageSize).Find(result).Error; err != nil {
		return nil, err
	}

	totalPage := int(totalCount) / query.PageSize
	if int(totalCount)%query.PageSize > 0 {
		totalPage++
	}

	return &PageResult{
		TotalPage:  totalPage,
		TotalCount: int(totalCount),
		Page:       query.Page,
		PageSize:   query.PageSize,
	}, nil
}
