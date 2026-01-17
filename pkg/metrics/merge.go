package metrics

import (
	"reflect"
	"strings"
)

// Merge combines multiple structs into a flat map.
// It handles both struct types and map types.
func Merge(items ...interface{}) Record {
	result := make(Record)
	for _, item := range items {
		if item == nil {
			continue
		}
		mergeValue(result, reflect.ValueOf(item))
	}
	return result
}

func mergeValue(result Record, v reflect.Value) {
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			keyStr := key.String()
			result[keyStr] = v.MapIndex(key).Interface()
		}
		return
	}

	// Handle structs
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Handle embedded structs recursively
		if field.Anonymous {
			mergeValue(result, fieldVal)
			continue
		}

		// Get JSON tag for field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Extract name from tag (handle "name,omitempty" format)
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			continue
		}

		// Skip nil pointers and empty slices
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			continue
		}
		if fieldVal.Kind() == reflect.Slice && fieldVal.Len() == 0 {
			continue
		}

		result[name] = fieldVal.Interface()
	}
}

// MergeRecords combines multiple Record maps into one.
func MergeRecords(records ...Record) Record {
	result := make(Record)
	for _, r := range records {
		for k, v := range r {
			result[k] = v
		}
	}
	return result
}
