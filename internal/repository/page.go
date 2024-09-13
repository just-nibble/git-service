package repository

import "github.com/just-nibble/git-service/internal/http/dtos"

const (
	DEFAULTPAGE                  = 1
	DEFAULTLIMIT                 = 10
	PageDefaultSortBy            = "created_at"
	PageDefaultSortDirectionDesc = "desc"
)

func getPaginationInfo(query dtos.APIPagingDto) (dtos.APIPagingDto, int) {
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

func getPagingInfo(query dtos.APIPagingDto, count int) dtos.PagingInfo {
	var hasNextPage bool

	next := int64((query.Page * query.Limit) - count)
	if next < 0 {
		hasNextPage = true
	}

	pagingInfo := dtos.PagingInfo{
		TotalCount:  int64(count),
		HasNextPage: hasNextPage,
		Page:        int(query.Page),
	}

	return pagingInfo
}
