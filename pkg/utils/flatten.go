package utils

import (
	"fmt"
	"reflect"
)

func Flatten(v any) map[string]any {
	result := make(map[string]any)
	flattenReflect(reflect.ValueOf(v), "", result)
	return result
}

func flattenReflect(v reflect.Value, prefix string, result map[string]any) {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		flattenStruct(v, prefix, result)
	case reflect.Map:
		flattenMap(v, prefix, result)
	case reflect.Slice, reflect.Array:
		flattenSlice(v, prefix, result)
	default:
		if prefix != "" {
			result[prefix] = v.Interface()
		}
	}
}

func flattenStruct(v reflect.Value, prefix string, result map[string]any) {
	t := v.Type()

	if isMetricType(t) {
		vField := v.FieldByName("V")
		tField := v.FieldByName("T")
		if prefix != "" {
			result[prefix] = vField.Interface()
			result[prefix+"T"] = tField.Interface()
		}
		return
	}

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		key := sf.Name
		if tag, ok := sf.Tag.Lookup("json"); ok {
			parts := splitFirst(tag, ',')
			if parts != "" && parts != "-" {
				key = parts
			}
		}

		childPrefix := key
		if prefix != "" {
			childPrefix = prefix + capitalize(key)
		}

		flattenReflect(v.Field(i), childPrefix, result)
	}
}

func flattenMap(v reflect.Value, prefix string, result map[string]any) {
	for _, key := range v.MapKeys() {
		k := fmt.Sprintf("%v", key.Interface())
		childPrefix := k
		if prefix != "" {
			childPrefix = prefix + capitalize(k)
		}
		flattenReflect(v.MapIndex(key), childPrefix, result)
	}
}

func flattenSlice(v reflect.Value, prefix string, result map[string]any) {
	for i := 0; i < v.Len(); i++ {
		childPrefix := fmt.Sprintf("%s%d", prefix, i)
		flattenReflect(v.Index(i), childPrefix, result)
	}
}

func isMetricType(t reflect.Type) bool {
	if t.NumField() != 2 {
		return false
	}
	vField, vOk := t.FieldByName("V")
	tField, tOk := t.FieldByName("T")
	if !vOk || !tOk {
		return false
	}
	if tField.Type.Kind() != reflect.Int64 {
		return false
	}
	switch vField.Type.Kind() {
	case reflect.Int64, reflect.Float64, reflect.String:
		return true
	}
	return false
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'a' && b[0] <= 'z' {
		b[0] -= 'a' - 'A'
	}
	return string(b)
}

func splitFirst(s string, sep byte) string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return s[:i]
		}
	}
	return s
}
