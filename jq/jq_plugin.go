package jq

import (
	"fmt"
	"sigs.k8s.io/kustomize/api/resource"
)

type Plugin struct {
	Replacements []Replacement
}

func (p *Plugin) Apply(r *resource.Resource) error {

	for _, rp := range p.Replacements {

		for _, t := range rp.Targets {
			meta, err := r.GetMeta()
			if err != nil {
				return err
			}

			if meta.APIVersion != t.GVK.GroupVersion().String() {
				continue
			}
			if meta.Kind != t.GVK.Kind {
				continue
			}
			if r.GetName() != t.Name {
				continue
			}
		}

		fmt.Println(rp)
	}

	return nil
}
