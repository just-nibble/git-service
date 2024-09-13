package domain

type (
	APIPaging struct {
		Limit     int
		Page      int
		Sort      string
		Direction string
	}

	PagingInfo struct {
		TotalCount  int64
		Page        int
		HasNextPage bool
		Count       int
	}
)
