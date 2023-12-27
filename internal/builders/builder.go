package builders

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/prefix"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type builderService struct {
	config        *config.Config
	prefixService *prefix.Prefix
	logger        *zap.Logger
}

type BuilderService interface {
	NewHandleBinding(c *gin.Context, paramater interface{}, dDslQ *[]types.QueryCondition) HandleBinding
	SelectFields(fields []string, tableName string, targetStruct interface{}) (string, error)
	IsValidTableName(tableName string) bool
	FieldExistsInStruct(field string, targetStruct interface{}) bool
	ConvertStringToUnix(dateString string) (int64, error)
	ConvertUnixToString(unixTime int64) string
	ConvertDateFields(fieldNames []string, params map[string]interface{}, operation string)
	StructToMap(item interface{}) map[string]interface{}
	GetSearchParameters(t reflect.Type) ([]types.SearchFields, error)
	ConstructSearchParameters(modelType reflect.Type, tableName string) (types.SearchParameters, error)
}

func NewBuilder(config *config.Config) BuilderService {
	service := &builderService{
		config:        config,
		logger:        config.Logger,
		prefixService: config.PrefixService,
	}

	return service
}

func (b *builderService) SelectFields(fields []string, tableName string, targetStruct interface{}) (string, error) {
	if !b.IsValidTableName(tableName) {
		return "", fmt.Errorf("invalid table name %s", tableName)
	}

	if len(fields) == 0 {
		return fmt.Sprintf("%s.*", tableName), nil
	}

	columnNames := make([]string, len(fields))

	for i, fieldName := range fields {
		// colName, err := b.structFieldToColumnName(fieldName, targetStruct)
		// if err != nil {
		// 	return "", err
		// }
		columnNames[i] = fmt.Sprintf("%s.%s", tableName, fieldName)
	}

	query := strings.Join(columnNames, ", ")
	return query, nil
}

func (b *builderService) IsValidTableName(tableName string) bool {
	if tableName == "*" {
		return true
	}

	exist := b.prefixService.IsTableNameExists(tableName)
	if !exist {
		b.logger.Error("No table found for service:", zap.String("tableName", tableName))
	}

	return true
}

// TODO  snake_case problem!
func (b *builderService) FieldExistsInStruct(field string, targetStruct interface{}) bool {
	val := reflect.ValueOf(targetStruct)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return false
	}

	camelCaseField := utils.ToCamelCase(field)

	return val.FieldByName(camelCaseField).IsValid()
}

func (b *builderService) StructFieldToColumnName(field string, targetStruct interface{}) (string, error) {
	val := reflect.Indirect(reflect.ValueOf(targetStruct))
	fieldInfo, found := val.Type().FieldByName(field)

	if !found {
		return "", fmt.Errorf("field '%s' not found", field)
	}

	gormTag := fieldInfo.Tag.Get("gorm")
	if gormTag == "" {
		return "", fmt.Errorf("field '%s' does not have a gorm tag", field)
	}

	tagParts := strings.Split(gormTag, ";")
	columnParts := strings.Split(tagParts[0], ":")
	if len(columnParts) != 2 {
		return "", fmt.Errorf("unexpected gorm tag format for field '%s'", field)
	}

	return columnParts[1], nil
}

func (b *builderService) ConvertStringToUnix(dateString string) (int64, error) {
	t, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		t, err = time.Parse("2006-01-02", dateString)
		if err != nil {
			return 0, err
		}
	}
	return t.Unix(), nil
}

func (b *builderService) ConvertUnixToString(unixTime int64) string {
	t := time.Unix(unixTime, 0)
	return t.Format("2006-01-02")
}

func (b *builderService) ConvertDateFields(fieldNames []string, params map[string]interface{}, operation string) {
	for _, field := range fieldNames {
		switch operation {
		case "toUnix":
			if dateStr, ok := params[field].(string); ok {
				if unixTime, err := b.ConvertStringToUnix(dateStr); err == nil {
					params[field] = unixTime
				} else {
					fmt.Printf("error converting %s to %s: %v", field, operation, err)
				}
			}
		case "toString":
			if unixTime, ok := params[field].(int64); ok {
				params[field] = b.ConvertUnixToString(unixTime)
			} else if timestampz, ok := params[field].(time.Time); ok {
				params[field] = timestampz.Format("2006-01-02T15:04:05Z")
			}
		}
	}
}

func (b *builderService) StructToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	elem := reflect.ValueOf(obj).Elem()

	for i := 0; i < elem.NumField(); i++ {
		fieldType := elem.Type().Field(i)
		fieldVal := elem.Field(i).Interface()

		if fieldType.PkgPath != "" {
			continue
		}
		result[fieldType.Name] = fieldVal
	}
	return result
}

func (b *builderService) GetSearchParameters(t reflect.Type) ([]types.SearchFields, error) {
	params := []types.SearchFields{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		formTag := field.Tag.Get("form")
		if formTag != "" && formTag != "-" {
			dataType := field.Type.String()
			params = append(params, types.SearchFields{Name: formTag, DataType: dataType})
		}
	}
	return params, nil
}

func (b *builderService) ConstructSearchParameters(modelType reflect.Type, tableName string) (types.SearchParameters, error) {
	dataMap, err := b.GetSearchParameters(modelType)
	if err != nil {
		return types.SearchParameters{}, err
	}
	return types.SearchParameters{TableName: tableName, Params: dataMap}, nil
}
