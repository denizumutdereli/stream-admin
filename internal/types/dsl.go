package types

type QueryCondition struct {
	Field    string
	Operator string
	Value    string
}

type SearchParameters struct {
	TableName string
	Params    []SearchFields
}

type SearchFields struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}
