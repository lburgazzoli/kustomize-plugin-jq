package jq_test

import (
	"errors"
	"io"
	"strings"

	jq2 "github.com/lburgazzoli/kustomize-plugin-jq/pkg/krm/jq"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"testing"

	. "github.com/onsi/gomega"

	gyq "github.com/lburgazzoli/gomega-matchers/pkg/matchers/yq"
)

const c = `
apiVersion: kustomize.lburgazzoli.github.io/v1alpha1
kind: JQTransform
metadata:
  name: jq-transformer
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/lburgazzoli/kustomize-plugin-jq:latest
spec:
  replacements:
    - sources:
        - selector:
            group: components.lburgazzoli.github.io
            version: v1alpha1
            kind: Configuration
            name: 'foo-config'
          name: '$fc'
        - selector:
            version: v1
            kind: ConfigMap
            name: 'foo-cm'
          name: '$fcm'
      targets:
      - selector:
          group: apps
          version: v1
          kind: Deployment
          name: 'foo-deployment'
        expressions:
        - |
          if ( $fc.spec.configuration.resources.type == "fixed" )
          then
            . | (.spec.template.spec.containers[] | select(.name == "manager")).resources |= $fc.spec.configuration.resources.fixed.resources
              | .spec.replicas = $fc.spec.configuration.resources.fixed.replicas
          end
        - '.spec.template.spec.containers[0].image = $fcm.data.image'
`

const v = `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: foo-cm
data:
  image: "quay.io/lburgazzoli/component:v1.1"
---
apiVersion: components.lburgazzoli.github.io/v1alpha1
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
      control-plane: foo-component
  template:
    metadata:
      labels:
        app: foo-component
    spec:
      containers:
        - name: manager
          image: quay.io/lburgazzoli/component:latest
          resources:
            limits:
              cpu: 500m
              memory: 2Gi
            requests:
              cpu: 10m
              memory: 64Mi
`

func TestJQ(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)
	cfg := jq2.Configuration{}

	p := framework.SimpleProcessor{
		Config: &cfg,
		Filter: kio.FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			f := jq2.Function{
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
	g.Expect(items).To(HaveLen(3))

	g.Expect(items[2]).Should(
		WithTransform(gyq.Extract(`.spec.template.spec.containers[0]`),
			And(
				gyq.Match(`.image == "quay.io/lburgazzoli/component:v1.1"`),
			),
		),
	)
	g.Expect(items[2]).Should(
		WithTransform(gyq.Extract(`.spec.template.spec.containers[0].resources`),
			And(
				gyq.Match(`.limits.cpu == "123m"`),
				gyq.Match(`.limits.memory == "456Mi"`),
				gyq.Match(`.requests.cpu == "321m"`),
				gyq.Match(`.requests.memory == "654Mi"`),
			),
		),
	)
	g.Expect(items[2]).Should(
		WithTransform(gyq.Extract(`.spec`),
			And(
				gyq.Match(`.replicas == "1"`),
			),
		),
	)

	// t.Log("\n" + w.String())
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

func (in *DocumentSplitter) String() string {
	return in.buffer.String()
}
