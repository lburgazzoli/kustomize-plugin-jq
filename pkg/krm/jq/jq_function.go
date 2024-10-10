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
		sources, err := SelectSources(rp, nodes...)
		if err != nil {
			return nil, err
		}

		for _, t := range rp.Targets {
			targetNodes, err := SelectNodes(t.Selector, nodes...)
			if err != nil {
				return nil, err
			}

			for _, tn := range targetNodes {
				for _, r := range t.Expressions {
					v, err := Run(r, tn, sources)
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

type ResolvedReplacement struct {
	R Replacement
	N *yaml.RNode
}

func SelectSources(replacement Replacement, nodes ...*yaml.RNode) (map[string]any, error) {
	answer := make(map[string]any)

	for _, s := range replacement.Sources {
		selected, err := SelectSource(s, nodes...)
		if err != nil {
			return nil, err
		}

		n := s.Name
		if n == "" {
			n = selected.GetName()
		}

		n = strings.ToLower(n)
		n = strings.ReplaceAll(n, "-", "_")
		n = strings.ReplaceAll(n, ".", "_")

		if !strings.HasPrefix(n, "$") {
			n = "$" + n
		}

		sm, err := selected.Map()
		if err != nil {
			return nil, fmt.Errorf("unable to map source for replacement %v: %w", replacement, err)
		}

		answer[n] = sm
	}

	return answer, nil
}

func SelectSource(source Source, nodes ...*yaml.RNode) (*yaml.RNode, error) {
	resources, err := SelectNodes(source.Selector, nodes...)
	if err != nil {
		return nil, err
	}

	if source.Selector.Predicate == "" && len(resources) == 1 {
		return nodes[0], nil
	}

	if source.Selector.Predicate == "" && len(resources) != 1 {
		return nil, errors.New("unable to determine source node")
	}

	query, err := gojq.Parse(source.Selector.Predicate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse expression %s, %w", source.Selector.Predicate, err)
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
