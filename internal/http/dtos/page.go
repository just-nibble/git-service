package dtos

type (
	APIPagingDto struct {
		Limit     int    `json:"limit,omitempty"`
		Page      int    `json:"page,omitempty"`
		Sort      string `json:"sort,omitempty"`
		Direction string `json:"direction,omitempty"`
	}

	PagingInfo struct {
		TotalCount  int64 `json:"totalCount"`
		Page        int   `json:"page"`
		HasNextPage bool  `json:"hasNextPage"`
		Count       int   `json:"count"`
	}
)
