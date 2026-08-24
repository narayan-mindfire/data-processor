package utils

import "strings"

// GetNestedField extracts a value from a map using dot-notation path
func GetNestedField(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for i, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		if i == len(parts)-1 {
			return val, true
		}
		current = val
	}
	return nil, false
}

func SetNestedField(data map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}

		next, exists := current[part]
		if !exists {
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		} else {
			nextMap, ok := next.(map[string]interface{})
			if !ok {
				// If we hit a non-map leaf before the end of the path, overwrite it with a new map
				newMap := make(map[string]interface{})
				current[part] = newMap
				current = newMap
			} else {
				current = nextMap
			}
		}
	}
}
