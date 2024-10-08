package jq

import (
	"fmt"

	"github.com/itchyny/gojq"
)

func MatchesE(expression string, data map[string]interface{}) (bool, error) {
	query, err := gojq.Parse(expression)
	if err != nil {
		return false, fmt.Errorf("unable to parse expression %s, %w", expression, err)
	}

	return Matches(query, data)
}

func Matches(query *gojq.Query, data map[string]interface{}) (bool, error) {
	it := query.Run(data)

	v, ok := it.Next()
	if !ok {
		return false, nil
	}

	if err, ok := v.(error); ok {
		return false, err
	}

	if match, ok := v.(bool); ok {
		return match, nil
	}

	return false, nil
}

func Run(query *gojq.Code, data map[string]interface{}, values ...any) (any, error) {
	it := query.Run(data, values...)

	v, ok := it.Next()
	if !ok {
		return nil, nil
	}

	if err, ok := v.(error); ok {
		return nil, err
	}

	return v, nil
}
