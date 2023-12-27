package database

type DataResult struct {
	Data interface{} `json:"data"`
}

func SingleDataResult(result interface{}) *DataResult {

	paginatedResult := &DataResult{
		Data: result,
	}

	return paginatedResult
}
