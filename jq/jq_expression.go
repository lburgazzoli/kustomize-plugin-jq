package jq

import (
	"fmt"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"strings"

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

func Run(expression string, node *yaml.RNode, values map[string]any) (*yaml.RNode, error) {
	data, err := node.Map()
	if err != nil {
		return nil, fmt.Errorf("unable to map target %v: %w", data, err)
	}

	sq, err := gojq.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("unable to parse expression %s, %w", expression, err)
	}

	keys := make([]string, 0, len(values))
	vals := make([]any, 0, len(values))

	for k, v := range values {
		if !strings.HasPrefix(k, "$") {
			k = "$" + k
		}

		k = strings.Replace(k, "-", "_", -1)

		keys = append(keys, k)
		vals = append(vals, v)
	}

	sqc, err := gojq.Compile(sq, gojq.WithVariables(keys))
	if err != nil {
		return nil, fmt.Errorf("unable to compile expression %s, %w", expression, err)
	}

	it := sqc.Run(data, vals...)

	v, ok := it.Next()
	if !ok {
		return nil, nil
	}

	if err, ok := v.(error); ok {
		return nil, err
	}

	ret, err := yaml.FromMap(v.(map[string]any))
	if err != nil {
		return nil, fmt.Errorf("unable to map target %v: %w", data, err)
	}

	return ret, nil
}
