package jq

import (
	"sigs.k8s.io/kustomize/api/types"
)

type Configuration struct {
	Metadata ConfigurationMeta `json:"metadata" yaml:"metadata"`
	Spec     ConfigurationSpec `json:"spec"     yaml:"spec"`
}

type ConfigurationMeta struct {
	Name string `json:"name" yaml:"name"`
}
type ConfigurationSpec struct {
	Replacements []Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

type Replacement struct {
	Source  Source   `json:",inline,omitempty" yaml:",inline,omitempty"`
	Targets []Target `json:"targets,omitempty" yaml:"targets,omitempty"`
}

type Source struct {
	types.Selector `json:",inline,omitempty" yaml:",inline,omitempty"`
	Expression     string `json:"expression,omitempty" yaml:"expression,omitempty"`
}

type Target struct {
	types.Selector `json:",inline,omitempty" yaml:",inline,omitempty"`
	Expressions    []string `json:"expressions,omitempty" yaml:"expressions,omitempty"`
}
