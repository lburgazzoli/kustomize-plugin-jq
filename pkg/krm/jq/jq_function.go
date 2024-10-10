package jq

import (
	"errors"
	"fmt"
	"strings"

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
			return nil, fmt.Errorf("unable to map source for replacement %v: %w", rp, err)
		}

		for _, t := range rp.Targets {
			targetNodes, err := SelectNodes(t.Selector, nodes...)
			if err != nil {
				return nil, err
			}

			for _, tn := range targetNodes {
				for _, r := range t.Expressions {
					n := rp.Source.Name
					if n == "" {
						n = source.GetName()
					}

					n = strings.ToLower(n)
					n = strings.ReplaceAll(n, "-", "_")
					n = strings.ReplaceAll(n, ".", "_")

					if !strings.HasPrefix(n, "$") {
						n = "$" + n
					}

					v, err := Run(r, tn, map[string]any{
						n: sm,
					})

					if err != nil {
						return nil, err
					}

					*tn = *v
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

	if replacement.Source.Selector.Predicate == "" && len(nodes) == 1 {
		return nodes[0], nil
	}

	if replacement.Source.Selector.Predicate == "" && len(nodes) != 1 {
		return nil, errors.New("unable to determine source node")
	}

	query, err := gojq.Parse(replacement.Source.Selector.Predicate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse expression %s, %w", replacement.Source.Selector.Predicate, err)
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

func SelectNodes(s Selector, nodes ...*yaml.RNode) ([]*yaml.RNode, error) {
	result := make([]*yaml.RNode, 0)

	sr, err := types.NewSelectorRegex(&s.Selector)
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
