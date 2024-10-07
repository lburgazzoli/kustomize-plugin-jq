package jq

import "k8s.io/apimachinery/pkg/runtime/schema"

type Replacement struct {
	Source   Source
	Selector Selector
	Targets  []Target
}

type Source struct {
	GVK  schema.GroupVersionKind
	Name string
}

type Selector struct {
	Expression string
}

type Target struct {
	GVK          schema.GroupVersionKind
	Name         string
	Replacements []Replace
}

type Replace struct {
	Selector string
	With     string
}
