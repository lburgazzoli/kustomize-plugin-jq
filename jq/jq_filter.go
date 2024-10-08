package jq

import (
	"fmt"
	"sigs.k8s.io/kustomize/kyaml/kio"

	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = (*Filter)(nil)

type Filter struct {
	Replacements []Replacement
}

func (f Filter) Filter(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
	return kio.FilterAll(kyaml.FilterFunc(f.run)).Filter(nodes)
}

func (f Filter) run(node *kyaml.RNode) (*kyaml.RNode, error) {
	m, err := node.Map()
	if err != nil {
		return node, nil
	}

	for _, rp := range f.Replacements {

		for _, t := range rp.Targets {
			meta, err := node.GetMeta()
			if err != nil {
				return node, err
			}

			if meta.TypeMeta != t.AsTypeMeta() {
				continue
			}
			if node.GetName() != t.Namespace {
				continue
			}
			if node.GetName() != t.Name {
				continue
			}
		}

		fmt.Println(">>> ", m)
		fmt.Println("### ", rp)
	}

	return nil, nil
}
