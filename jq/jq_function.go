package jq

import (
	"fmt"
	"github.com/itchyny/gojq"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Function struct {
	Replacements []Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

func (p *Function) Apply(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, rp := range p.Replacements {
		source, err := SelectSource(rp, nodes...)
		if err != nil {
			return nil, err
		}

		if source == nil {
			continue
		}

		sm, err := source.Map()
		if err != nil {
			return nil, err
		}

		for _, t := range rp.Targets {
			targetNodes, err := SelectNodes(t.Selector, nodes...)
			if err != nil {
				return nil, err
			}

			for _, tn := range targetNodes {
				tm, err := tn.Map()
				if err != nil {
					return nil, err
				}

				for _, r := range t.Expressions {
					sq, err := gojq.Parse(r)
					if err != nil {
						return nil, fmt.Errorf("unable to parse expression %s, %w", r, err)
					}

					sqc, err := gojq.Compile(sq, gojq.WithVariables([]string{"$s"}))
					if err != nil {
						return nil, fmt.Errorf("unable to compile expression %s, %w", r, err)
					}

					sv, err := Run(sqc, tm, sm)
					if err != nil {
						return nil, err
					}

					b, err := yaml.Marshal(sv)
					if err != nil {
						return nil, err
					}

					fmt.Println(string(b))
				}
			}
		}
	}

	return nil, nil
}

func SelectSource(replacement Replacement, nodes ...*yaml.RNode) (*yaml.RNode, error) {
	resources, err := SelectNodes(replacement.Source.Selector, nodes...)
	if err != nil {
		return nil, err
	}

	query, err := gojq.Parse(replacement.Source.Expression)
	if err != nil {
		return nil, fmt.Errorf("unable to parse expression %s, %w", replacement.Source.Expression, err)
	}

	for _, r := range resources {
		data, err := r.Map()
		if err != nil {
			return nil, err
		}

		matches, err := Matches(query, data)
		if err != nil {
			return nil, err
		}

		if matches {
			return r, nil
		}
	}

	return nil, nil
}

func SelectNodes(s types.Selector, nodes ...*yaml.RNode) ([]*yaml.RNode, error) {
	var result []*yaml.RNode

	sr, err := types.NewSelectorRegex(&s)
	if err != nil {
		return nil, err
	}

	for _, r := range nodes {
		curId := resid.NewResIdWithNamespace(
			resid.GvkFromNode(r),
			r.GetName(),
			r.GetNamespace())

		if !sr.MatchNamespace(curId.EffectiveNamespace()) {
			continue
		}
		if !sr.MatchName(curId.Name) {
			continue
		}
		if !sr.MatchGvk(curId.Gvk) {
			continue
		}

		matched, err := r.MatchesLabelSelector(s.LabelSelector)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}

		matched, err = r.MatchesAnnotationSelector(s.AnnotationSelector)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}

		result = append(result, r)
	}

	return result, nil
}
