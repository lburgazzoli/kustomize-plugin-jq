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

//nolint:gocognit,cyclop
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
			return nil, fmt.Errorf("unable to map source for replacement %v: %w", rp, err)
		}

		for _, t := range rp.Targets {
			targetNodes, err := SelectNodes(t.Selector, nodes...)
			if err != nil {
				return nil, err
			}

			for _, tn := range targetNodes {
				for _, r := range t.Expressions {
					tm, err := tn.Map()
					if err != nil {
						return nil, fmt.Errorf("unable to map target %v: %w", tn, err)
					}

					sq, err := gojq.Parse(r)
					if err != nil {
						return nil, fmt.Errorf("unable to parse expression %s, %w", r, err)
					}

					sqc, err := gojq.Compile(sq, gojq.WithVariables([]string{"$s"}))
					if err != nil {
						return nil, fmt.Errorf("unable to compile expression %s, %w", r, err)
					}

					v, err := Run(sqc, tm, sm)
					if err != nil {
						return nil, err
					}

					n, err := yaml.FromMap(v.(map[string]any))
					if err != nil {
						return nil, fmt.Errorf("unable to map target %v: %w", tn, err)
					}

					*tn = *n
				}
			}
		}
	}

	return nodes, nil
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
			return nil, fmt.Errorf("unable to map resource %v, %w", r, err)
		}

		matches, err := Matches(query, data)
		if err != nil {
			return nil, fmt.Errorf("unable to match resource %v, %w", r, err)
		}

		if matches {
			return r, nil
		}
	}

	return nil, nil
}

func SelectNodes(s types.Selector, nodes ...*yaml.RNode) ([]*yaml.RNode, error) {
	result := make([]*yaml.RNode, 0)

	sr, err := types.NewSelectorRegex(&s)
	if err != nil {
		return nil, fmt.Errorf("unable to create selector regex, %w", err)
	}

	for _, r := range nodes {
		curID := resid.NewResIdWithNamespace(
			resid.GvkFromNode(r),
			r.GetName(),
			r.GetNamespace())

		if !sr.MatchNamespace(curID.EffectiveNamespace()) {
			continue
		}

		if !sr.MatchName(curID.Name) {
			continue
		}

		if !sr.MatchGvk(curID.Gvk) {
			continue
		}

		matched, err := r.MatchesLabelSelector(s.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("unable to match label selector, %w", err)
		}

		if !matched {
			continue
		}

		matched, err = r.MatchesAnnotationSelector(s.AnnotationSelector)
		if err != nil {
			return nil, fmt.Errorf("unable to match annotation selector, %w", err)
		}

		if !matched {
			continue
		}

		result = append(result, r)
	}

	return result, nil
}
