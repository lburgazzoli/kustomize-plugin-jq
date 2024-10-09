package main

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	kjq "github.com/lburgazzoli/kustomize-plugin-jq/pkg/krm/jq"
)

func main() {
	c := kjq.Configuration{}

	p := framework.SimpleProcessor{
		Config: &c,
		Filter: kio.FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			f := kjq.Function{
				Replacements: c.Spec.Replacements,
			}

			return f.Apply(nodes)
		}),
	}

	cmd := command.Build(p, command.StandaloneDisabled, false)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
