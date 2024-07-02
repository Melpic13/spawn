package tools

import "fmt"

// ValidateInput performs minimal schema validation.
func ValidateInput(required []string, input map[string]interface{}) error {
	for _, key := range required {
		if _, ok := input[key]; !ok {
			return fmt.Errorf("validate input: missing required field %q", key)
		}
	}
	return nil
}
