package jq_test

import (
	"errors"
	"io"
	"strings"

	"github.com/lburgazzoli/kustomize-plugin-jq/jq"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"testing"

	. "github.com/onsi/gomega"

	gyq "github.com/lburgazzoli/gomega-matchers/pkg/matchers/yq"
)

const c = `
apiVersion: kustomize.opendatahub.io/v1alpha1
kind: JQTransform
metadata:
  name: jq-transformer
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ../../bin/jq-transform
spec:
  replacements:
    - source:
        group: components.opendatahub.io
        version: v1alpha1
        kind: Configuration
        name: 'foo-config'
        expression: '.spec.configuration.resources.type == "fixed"'
      targets:
      - group: apps
        version: v1
        kind: Deployment
        name: 'foo-deployment'
        expressions:
        - '(.spec.template.spec.containers[] | select(.name == "manager")).resources.limits |= $s.spec.configuration.resources.fixed.resources.limits'
        - '.spec.template.spec.containers[0].resources.requests = $s.spec.configuration.resources.fixed.resources.requests'
        - '.spec.replicas = $s.spec.configuration.resources.fixed.replicas'
`

const v = `
---
apiVersion: components.opendatahub.io/v1alpha1
kind: Configuration
metadata:
  name: foo-config
spec:
  configuration:
    resources:
      type: fixed
      fixed:
        replicas: 1
        resources:
          limits:
            cpu: 123m
            memory: 456Mi
          requests:
            cpu: 321m
            memory: 654Mi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      control-plane: odh-component
  template:
    metadata:
      labels:
        app: odh-component
    spec:
      containers:
        - name: manager
          image: quay.io/opendatahub/odh-component:latest
          resources:
            limits:
              cpu: 500m
              memory: 2Gi
            requests:
              cpu: 10m
              memory: 64Mi
`

func TestJS(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)
	cfg := jq.Configuration{}

	p := framework.SimpleProcessor{
		Config: &cfg,
		Filter: kio.FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			f := jq.Function{
				Replacements: cfg.Spec.Replacements,
			}

			return f.Apply(nodes)
		}),
	}

	w := DocumentSplitter{}

	rw := &kio.ByteReadWriter{
		Writer:                &w,
		KeepReaderAnnotations: false,
		NoWrap:                true,
		FunctionConfig:        yaml.MustParse(c),
		Reader:                strings.NewReader(v),
	}

	err := framework.Execute(p, rw)
	g.Expect(err).ToNot(HaveOccurred())

	items, err := w.Items()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(items).To(HaveLen(2))

	g.Expect(items[1]).Should(
		WithTransform(gyq.Extract(`.spec.template.spec.containers[0].resources`),
			And(
				gyq.Match(`.limits.cpu == "123m"`),
				gyq.Match(`.limits.memory == "456Mi"`),
				gyq.Match(`.requests.cpu == "321m"`),
				gyq.Match(`.requests.memory == "654Mi"`),
			),
		),
	)
	g.Expect(items[1]).Should(
		WithTransform(gyq.Extract(`.spec`),
			And(
				gyq.Match(`.replicas == "1"`),
			),
		),
	)
}

type DocumentSplitter struct {
	buffer strings.Builder
}

func (in *DocumentSplitter) Write(p []byte) (int, error) {
	return in.buffer.Write(p)
}

func (in *DocumentSplitter) Reset() {
	in.buffer.Reset()
}

func (in *DocumentSplitter) Items() ([]string, error) {
	items := make([]string, 0)

	r := strings.NewReader(in.buffer.String())
	dec := yaml.NewDecoder(r)

	for {
		var node yaml.Node

		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		item, err := yaml.Marshal(&node)
		if err != nil {
			return nil, err
		}

		items = append(items, string(item))
	}

	return items, nil
}
