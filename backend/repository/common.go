package repository

import (
	"gorm.io/gorm"
)

type PageQuery struct {
	Page     int
	PageSize int
}

func (q *PageQuery) Normalize() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
}

func (q *PageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

type PageResult struct {
	TotalPage  int
	TotalCount int
	Page       int
	PageSize   int
}

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
