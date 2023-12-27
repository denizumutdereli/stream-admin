package validation

import (
	"time"

	"github.com/go-playground/validator"
)

var IsInteger validator.Func = func(fl validator.FieldLevel) bool {
	_, ok := fl.Field().Interface().(int)
	return ok
}

var BookableDate validator.Func = func(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if ok {
		today := time.Now()
		if today.After(date) {
			return false
		}
	}
	return true
}
