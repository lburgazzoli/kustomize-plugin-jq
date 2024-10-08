package jq_test

import (
	"github.com/lburgazzoli/kustomize-plugin-jq/jq"
	"os"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"strings"

	"testing"

	. "github.com/onsi/gomega"
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
        - '.spec.template.spec.containers[0].resources.limits = $s.spec.configuration.resources.fixed.resources.limits'
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
        resources:
          limits:
            cpu: 123m
            memory: 456i
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
        app.opendatahub.io/odh-component: "true"
        control-plane: odh-component
    spec:
      containers:
        - name: manager
          image: quay.io/opendatahub/odh-component:latest
          ports:
            - containerPort: 8080
          resources:
            limits:
              cpu: 500m
              memory: 2Gi
            requests:
              cpu: 10m
              memory: 64Mi
`

func TestJS(t *testing.T) {
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

	rw := &kio.ByteReadWriter{
		Writer:                os.Stdout,
		KeepReaderAnnotations: false,
		NoWrap:                true,
		FunctionConfig:        yaml.MustParse(c),
		Reader:                strings.NewReader(v),
	}

	err := framework.Execute(p, rw)
	g.Expect(err).To(BeNil())
}
