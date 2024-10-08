package jq_test

import (
	"github.com/lburgazzoli/kustomize-plugin-jq/jq"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"

	"sigs.k8s.io/kustomize/api/provider"

	"testing"

	. "github.com/onsi/gomega"
)

const c = `
apiVersion: components.opendatahub.io/v1alpha1
kind: Configuration
metadata:
  name: foo-config
spec:
  configuration:
	resources:
	  fixed:
	    resources:
          limits:
            cpu: 100m
            memory: 1Gi
`

const d = `
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

func TestJQ(t *testing.T) {
	g := NewWithT(t)

	p := jq.Plugin{
		Replacements: []jq.Replacement{
			{
				Source: jq.Source{
					Selector: types.Selector{
						ResId: resid.ResId{
							Gvk:  resid.NewGvk("components.opendatahub.io", "v1alpha1", "Configuration"),
							Name: "foo-config",
						},
					},
					Expression: `.spec.configuration.resources.type == "fixed"`,
				},
				Targets: []jq.Target{
					{
						Selector: types.Selector{
							ResId: resid.ResId{
								Gvk:  resid.NewGvk("apps", "v1", "Deployment"),
								Name: "foo-deployment",
							},
						},
						Replacements: []jq.Replace{
							{
								Source: ".spec.configuration.resources.fixed.resources.limits",
								Target: ".spec.template.spec.containers | select(.name == \"controller\") | .resources.limits",
							},
						},
					},
				},
			},
		},
	}

	factory := provider.NewDefaultDepProvider().GetResourceFactory()

	res, err := factory.FromBytes([]byte(d))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).NotTo(BeNil())

	err = p.Apply(res)
	g.Expect(err).NotTo(HaveOccurred())
}
