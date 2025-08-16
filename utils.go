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
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	var result []squirrel.Sqlizer
	typ := val.Type()

	for i := 0; i < val.Type().NumField(); i++ {
		if val.Field(i).IsNil() {
			continue
		}
		tag := typ.Field(i).Tag.Get(filterTag)
		if tag == "" {
			continue
		}

		result = append(result, squirrel.Eq{tag: val.Field(i).Interface()})
	}

	return result
}
