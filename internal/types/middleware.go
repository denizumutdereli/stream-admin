package types

type QueryParams struct {
	Pagination PaginationParams
	DSLQuery   []QueryCondition
}

type PaginationParams struct {
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
}
