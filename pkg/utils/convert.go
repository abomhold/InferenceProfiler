package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ToFloat64 converts a value to float64, returning 0 on failure.
func ToFloat64(v interface{}) float64 {
	f, _ := ToFloat64Ok(v)
	return f
}

// ToFloat64Ok converts a value to float64, returning success status.
func ToFloat64Ok(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}

	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

// ToInt64 converts a value to int64, returning 0 on failure.
func ToInt64(v interface{}) int64 {
	i, _ := ToInt64Ok(v)
	return i
}

// ToInt64Ok converts a value to int64, returning success status.
func ToInt64Ok(v interface{}) (int64, bool) {
	if v == nil {
		return 0, false
	}

	switch n := v.(type) {
	case int64:
		return n, true
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case uint:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return int64(n), true
	case float32:
		// Only convert if it's a whole number
		if n == float32(int64(n)) {
			return int64(n), true
		}
	case float64:
		if n == float64(int64(n)) {
			return int64(n), true
		}
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		return i, err == nil
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	}
	return 0, false
}

// ToString converts a value to string.
func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToBool converts a value to bool, returning false on failure.
func ToBool(v interface{}) bool {
	b, _ := ToBoolOk(v)
	return b
}

// ToBoolOk converts a value to bool, returning success status.
func ToBoolOk(v interface{}) (bool, bool) {
	if v == nil {
		return false, false
	}

	switch b := v.(type) {
	case bool:
		return b, true
	case string:
		switch b {
		case "true", "1", "yes", "True", "TRUE", "Yes", "YES":
			return true, true
		case "false", "0", "no", "False", "FALSE", "No", "NO":
			return false, true
		}
	case int, int8, int16, int32, int64:
		return ToInt64(v) != 0, true
	case float32, float64:
		return ToFloat64(v) != 0, true
	}
	return false, false
}

// FormatValue converts any value to a string representation for CSV/TSV output.
func FormatValue(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case bool:
		return strconv.FormatBool(val)
	case uint64:
		return strconv.FormatUint(val, 10)
	case json.Number:
		return val.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
