package interpreters

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/denizumutdereli/stream-admin/internal/types"
	"gorm.io/gorm"
)

func ProcessFields(db *gorm.DB, params interface{}, omitFields []string, tableName string) *gorm.DB {
	val := reflect.ValueOf(params).Elem()

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		fieldType := val.Type().Field(i)

		formTagValue := fieldType.Tag.Get("form")
		if formTagValue == "" {
			continue
		}

		if valueField.Kind() == reflect.Ptr && !valueField.IsNil() && ShouldProcess(formTagValue, omitFields) {
			actualValue := valueField.Elem().Interface()
			db = ApplyFilters(db, formTagValue, actualValue, "=", tableName)
		}
	}

	return db
}

func ShouldProcess(fieldName string, omitFields []string) bool {
	for _, omit := range omitFields {
		if fieldName == omit {
			return false
		}
	}
	return true
}

func ApplyFilters(db *gorm.DB, field string, value interface{}, operator string, tableName string) *gorm.DB {
	if value == nil || value == "" {
		return db
	}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return db
	}
	return db.Debug().Where(fmt.Sprintf("%s.%s %s ?", tableName, field, operator), value)
}

func DSLSearchOperator(db *gorm.DB, condition *types.QueryCondition, tableName string) *gorm.DB {
	caseInsensitive := false

	if strings.HasSuffix(condition.Value, "\\i") {
		caseInsensitive = true
		condition.Value = strings.TrimSuffix(condition.Value, "\\i")
	}

	switch condition.Operator {
	case "=", "eq":
		db = ApplyFilters(db, condition.Field, condition.Value, "=", tableName)
	case ">", "gt":
		db = ApplyFilters(db, condition.Field, condition.Value, ">", tableName)
	case ">=", "gte":
		db = ApplyFilters(db, condition.Field, condition.Value, ">=", tableName)
	case "<", "lt":
		db = ApplyFilters(db, condition.Field, condition.Value, "<", tableName)
	case "<=", "lte":
		db = ApplyFilters(db, condition.Field, condition.Value, "<=", tableName)
	case "!=", "neq":
		db = ApplyFilters(db, condition.Field, condition.Value, "!=", tableName)
	case "contains", "cont", "inc":
		if caseInsensitive {
			db = ApplyFilters(db, condition.Field, "%"+strings.ToLower(condition.Value)+"%", "ILIKE", tableName)
		} else {
			db = ApplyFilters(db, condition.Field, "%"+condition.Value+"%", "LIKE", tableName)
		}
	case "notcontains", "nocont", "noinc":
		if caseInsensitive {
			db = ApplyFilters(db, condition.Field, "%"+strings.ToLower(condition.Value)+"%", "NOT ILIKE", tableName)
		} else {
			db = ApplyFilters(db, condition.Field, "%"+condition.Value+"%", "NOT LIKE", tableName)
		}
	case "start":
		if caseInsensitive {
			db = ApplyFilters(db, condition.Field, "%"+condition.Value, "LIKE", tableName)
		} else {
			db = ApplyFilters(db, condition.Field, "%"+condition.Value, "ILIKE", tableName)
		}
		db = ApplyFilters(db, condition.Field, "%"+condition.Value, "LIKE", tableName)
	case "end":

		if caseInsensitive {
			db = ApplyFilters(db, condition.Field, condition.Value+"%", "ILIKE", tableName)
		} else {
			db = ApplyFilters(db, condition.Field, condition.Value+"%", "LIKE", tableName)
		}

	default:
	}
	return db
}
