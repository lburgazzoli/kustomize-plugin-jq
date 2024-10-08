package main

import (
	"github.com/lburgazzoli/kustomize-plugin-jq/jq"
	"os"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	c := jq.Configuration{}

	p := framework.SimpleProcessor{
		Config: &c,
		Filter: kio.FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			f := jq.Function{
				Replacements: c.Spec.Replacements,
			}

			return f.Apply(nodes)
		}),
	}

	cmd := command.Build(p, command.StandaloneEnabled, false)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
