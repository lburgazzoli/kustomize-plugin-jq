package jq

import (
	"fmt"
	"github.com/itchyny/gojq"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

var _ resmap.Transformer = (*Plugin)(nil)

type Plugin struct {
	Replacements []Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

func (p *Plugin) Config(h *resmap.PluginHelpers, c []byte) error {
	err := k8syaml.Unmarshal(c, p)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) Transform(m resmap.ResMap) error {
	for _, rp := range p.Replacements {
		resources, err := m.Select(rp.Source.Selector)
		if err != nil {
			return err
		}

		source, err := Select(rp, resources...)
		if err != nil {
			return err
		}

		if source == nil {
			continue
		}

	}

	return m.ApplyFilter(p.filter())
}

func (p *Plugin) Apply(r *resource.Resource) error {

	_, err := p.filter().Filter(
		[]*kyaml.RNode{&r.RNode},
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) filter() Filter {
	return Filter{
		Replacements: p.Replacements,
	}
}

func Select(replacement Replacement, resources ...*resource.Resource) (*resource.Resource, error) {
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
