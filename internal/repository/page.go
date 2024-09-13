package repository

import (
	"github.com/just-nibble/git-service/internal/domain"
)

const (
	DEFAULTPAGE                  = 1
	DEFAULTLIMIT                 = 10
	PageDefaultSortBy            = "created_at"
	PageDefaultSortDirectionDesc = "desc"
)

func getPaginationInfo(query domain.APIPaging) (domain.APIPaging, int) {
	var offset int
	// load defaults
	if query.Page == 0 {
		query.Page = DEFAULTPAGE
	}
	if query.Limit == 0 {
		query.Limit = DEFAULTLIMIT
	}

	if query.Sort == "" {
		query.Sort = PageDefaultSortBy
	}

	if query.Direction == "" {
		query.Direction = PageDefaultSortDirectionDesc
	}

	if query.Page > 1 {
		offset = query.Limit * (query.Page - 1)
	}
	return query, offset
}

func getPagingInfo(query domain.APIPaging, count int) domain.PagingInfo {
	var hasNextPage bool

	next := int64((query.Page * query.Limit) - count)
	if next < 0 {
		hasNextPage = true
	}

	pagingInfo := domain.PagingInfo{
		TotalCount:  int64(count),
		HasNextPage: hasNextPage,
		Page:        int(query.Page),
	}

	return pagingInfo
}
