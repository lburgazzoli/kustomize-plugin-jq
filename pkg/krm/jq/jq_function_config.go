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
	Selector Selector `json:"selector,omitempty" yaml:"selector,omitempty"`
	Name     string   `json:"name,omitempty" yaml:"name,omitempty"`
}

type Selector struct {
	types.Selector `json:",inline,omitempty" yaml:",inline,omitempty"`
	Predicate      string `json:"predicate,omitempty" yaml:"predicate,omitempty"`
}

type Target struct {
	Selector    Selector `json:"selector,omitempty" yaml:"selector,omitempty"`
	Expressions []string `json:"expressions,omitempty" yaml:"expressions,omitempty"`
}
