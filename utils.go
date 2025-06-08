package datagate

import (
	"reflect"

	"github.com/Masterminds/squirrel"
)

const (
	dbTag     = "db"
	filterTag = "filter"
	insertTag = "insert"
)

func extractStructFieldsByTag(s interface{}, tagName string) map[string]interface{} {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]interface{})
	typ := val.Type()

	for i := 0; i < val.Type().NumField(); i++ {
		tag := typ.Field(i).Tag.Get(tagName)
		if tag == "" {
			continue
		}

		result[tag] = val.Field(i).Interface()
	}

	return result
}

func buildSqlFiltersFromStruct(s interface{}) []squirrel.Sqlizer {
	var result []squirrel.Sqlizer

	values := extractStructFieldsByTag(s, filterTag)
	for key, value := range values {
		result = append(result, squirrel.Eq{key: value})
	}

	return result
}
