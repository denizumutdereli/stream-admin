package builders

import (
	"errors"
	"strings"

	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

type HandleBinding interface {
	BindAndValidate() *handleBinding
	BindQuery() *handleBinding
	BindJson() *handleBinding
	BindDSL() *handleBinding
	BindPagination(paginationParams *types.PaginationParams) *handleBinding
	Validate() *handleBinding
	GetError() error
	GetErrorMessages() map[string]string
}

type handleBinding struct {
	Context          *gin.Context
	Error            error
	ErrorMessages    map[string]string
	Modal            interface{}
	DslQ             *[]types.QueryCondition
	PaginationParams types.PaginationParams
	mid_             types.QueryParams
}

func (b *builderService) NewHandleBinding(c *gin.Context, paramater interface{}, dDslQ *[]types.QueryCondition) HandleBinding {
	return &handleBinding{Context: c, Modal: paramater, DslQ: dDslQ}
}

func (b *handleBinding) BindAndValidate() *handleBinding {
	if b.Error != nil {
		return b
	}
	switch b.Context.Request.Header.Get("Content-Type") {
	case "application/json":
		b.Error = b.Context.ShouldBindJSON(b.Modal)
	case "application/x-www-form-urlencoded":
		b.Error = b.Context.ShouldBind(b.Modal)
	}

	if b.Error != nil {
		return b
	}

	b.Validate()

	return b
}

func (b *handleBinding) BindQuery() *handleBinding {
	if b.Error == nil {
		if err := b.Context.ShouldBindQuery(b.Modal); err != nil {
			b.Error = err
		}
	}
	return b
}

func (b *handleBinding) BindJson() *handleBinding {
	if b.Error == nil {
		if err := b.Context.ShouldBindJSON(b.Modal); err != nil {
			b.Error = err
		}
	}
	return b
}

func (b *handleBinding) BindDSL() *handleBinding {

	if b.Error == nil {
		dsl_search := b.Context.Query("dsl_search")
		if dsl_search != "" {
			dslQuery, parseErr := ParseDSLSearch(dsl_search, b.Modal)
			if parseErr != nil {
				b.Error = errors.New(parseErr.Error())
			} else {
				*b.DslQ = dslQuery
			}
		} else {
			*b.DslQ = nil
		}
	}
	return b
}

func (b *handleBinding) BindDSLField(field string) *handleBinding {

	if b.Error == nil {
		dsl_search := b.Context.Query(field)
		if dsl_search != "" {
			dslQuery, parseErr := ParseDSLSearch(dsl_search, b.Modal)
			if parseErr != nil {
				b.Error = errors.New(parseErr.Error())
			} else {
				*b.DslQ = dslQuery
			}
		} else {
			*b.DslQ = nil
		}
	}
	return b
}

func (b *handleBinding) BindPagination(paginationParams *types.PaginationParams) *handleBinding {
	if b.Error == nil {
		pagination, exists := b.Context.Get("pagination")
		if !exists {
			b.Error = errors.New("pagination not exist")
		} else {
			if pagParams, ok := pagination.(types.PaginationParams); ok {
				*paginationParams = pagParams
			} else {
				b.Error = errors.New("pagination is not of the correct type")
			}
		}
	}
	return b
}

func (b *handleBinding) Validate() *handleBinding {
	if b.Error == nil {
		validate := validator.New()
		err := validate.Struct(b.Modal)

		if err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				b.ErrorMessages = make(map[string]string)
				for _, errField := range validationErrors {
					b.ErrorMessages[errField.Field()] = errField.Translate(nil)
				}
				b.Error = errors.New("validation failed")
			} else {
				b.Error = err
			}
		}

	}
	return b
}

func (b *handleBinding) GetError() error {
	return b.Error
}

func (b *handleBinding) GetErrorMessages() map[string]string {
	return b.ErrorMessages
}

func ParseDSLSearch(dsl string, targetStruct interface{}) ([]types.QueryCondition, error) {
	var queryConditions []types.QueryCondition

	conditions := utils.CleanInput(dsl)

	for _, condition := range conditions {
		parts := strings.Split(condition, "|")

		if len(parts) != 3 {
			return nil, errors.New("invalid condition format")
		}

		queryConditions = append(queryConditions, types.QueryCondition{
			Field:    strings.TrimSpace(parts[0]),
			Operator: strings.TrimSpace(parts[1]),
			Value:    strings.TrimSpace(parts[2]),
		})
	}

	return queryConditions, nil
}
