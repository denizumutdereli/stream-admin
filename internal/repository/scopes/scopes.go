package scopes

import (
	"fmt"

	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/denizumutdereli/stream-admin/internal/repository/interpreters"
	"gorm.io/gorm"
)

type FilterParam struct {
	Field      string
	Comparison string
	Value      interface{}
}

func Paginate(page int, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func OrderBy(sortBy string, sortOrder string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sortBy != "" && sortOrder != "" {
			db = db.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
		}
		return db
	}
}

func WhereFinder(field string, comparison string, value interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if value != nil && field != "" && comparison != "" {
			return db.Where(fmt.Sprintf("%s %s ?", field, comparison), value)
		}
		return db
	}
}

func Limiter(limit uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if limit > 0 {
			return db.Limit(int(limit))
		}
		return db
	}
}

func LatestRecord(db *gorm.DB) *gorm.DB {
	return db.Order("id DESC").Limit(1)
}

func ApplySearchFilters(params interface{}, tableName string, isDslEnable bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		omitFields := []string{"dsl_search_operator", "dsl_search"}
		db = interpreters.ProcessFields(db, params, omitFields, tableName)

		if searcher, ok := params.(dsl.DSLSearcher); ok && isDslEnable {
			if ops := searcher.GetDSLSearchOperator(); ops != nil {
				for _, condition := range *ops {
					db = interpreters.DSLSearchOperator(db, &condition, tableName)
				}
			}
		}

		return db
	}
}
