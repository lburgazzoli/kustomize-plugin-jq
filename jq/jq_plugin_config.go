package jq

import (
	"sigs.k8s.io/kustomize/api/types"
)

type Replacement struct {
	Source  Source   `json:",inline,omitempty" yaml:",inline,omitempty"`
	Targets []Target `json:"targets,omitempty" yaml:"targets,omitempty"`
}

type Source struct {
	types.Selector `json:",inline,omitempty" yaml:",inline,omitempty"`
	Expression     string `json:"expression,omitempty" yaml:"expression,omitempty"`
}

type Selector struct {
	Expression string `json:"expression,omitempty" yaml:"expression,omitempty"`
}

type Target struct {
	types.Selector `json:",inline,omitempty" yaml:",inline,omitempty"`
	Replacements   []Replace `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

type Replace struct {
	Target string `json:"field,omitempty" yaml:"field,omitempty"`
	Source string `json:"value,omitempty" yaml:"value,omitempty"`
}
