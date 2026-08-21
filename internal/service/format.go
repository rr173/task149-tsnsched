package service

import (
	"encoding/json"
	"example.com/task149/tsnsched/internal/model"
)

func EncodeValidation(v model.ValidationResult) ([]byte, error) { return json.Marshal(v) }
func ExplainValidation(v model.ValidationResult) []string {
	out := make([]string, 0, len(v.Violations))
	for _, item := range v.Violations {
		out = append(out, string(item.Kind)+": "+item.Detail)
	}
	return out
}
