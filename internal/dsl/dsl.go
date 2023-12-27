package dsl

import "github.com/denizumutdereli/stream-admin/internal/types"

type DSLFields struct {
	DSLSearch         *string                 `form:"dsl_search"`
	DSLSearchOperator *[]types.QueryCondition `form:"dsl_search_operator"`
}

type DSLSearcher interface {
	GetDSLSearch() *string
	GetDSLSearchOperator() *[]types.QueryCondition
}

func (d *DSLFields) GetDSLSearch() *string {
	return d.DSLSearch
}

func (d *DSLFields) GetDSLSearchOperator() *[]types.QueryCondition {
	return d.DSLSearchOperator
}
