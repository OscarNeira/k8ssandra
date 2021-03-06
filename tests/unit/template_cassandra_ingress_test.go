package unit_test

import (
	. "fmt"
	"path/filepath"

	"github.com/gruntwork-io/terratest/modules/helm"
	helmUtils "github.com/k8ssandra/k8ssandra/tests/unit/utils/helm"
	. "github.com/k8ssandra/k8ssandra/tests/unit/utils/traefik"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
)

var _ = Describe("Verify Cassandra ingress template", func() {
	var (
		helmChartPath string
		err           error
		ingress       traefik.IngressRouteTCP
	)

	BeforeEach(func() {
		helmChartPath, err = filepath.Abs(ChartsPath)
		Expect(err).To(BeNil())
		ingress = traefik.IngressRouteTCP{}
	})

	AfterEach(func() {
		err = nil
	})

	renderTemplate := func(options *helm.Options) error {
		return helmUtils.RenderAndUnmarshall("templates/cassandra/ingress.yaml",
			options, helmChartPath, HelmReleaseName,
			func(renderedYaml string) error {
				return helm.UnmarshalK8SYamlE(GinkgoT(), renderedYaml, &ingress)
			})
	}

	Context("by confirming it does not render when", func() {
		It("is implicitly disabled", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
			}
			Expect(renderTemplate(options)).ShouldNot(Succeed())
		})

		It("is explicitly disabled at the Ingress level", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"ingress.traefik.enabled": "false",
				},
			}
			Expect(renderTemplate(options)).ShouldNot(Succeed())
		})

		It("is explicitly disabled at the Cassandra level", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"ingress.traefik.enabled":           "true",
					"ingress.traefik.cassandra.enabled": "false",
				},
			}
			Expect(renderTemplate(options)).ShouldNot(Succeed())
		})

		It("is explicitly disabled at the Ingress level even when enabled at the Cassandra level", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"ingress.traefik.enabled":           "false",
					"ingress.traefik.cassandra.enabled": "true",
				},
			}
			Expect(renderTemplate(options)).ShouldNot(Succeed())
		})
	})

	Context("by rendering it when", func() {
		It("it is enabled and Stargate Cassandra ingress is disabled", func() {
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"ingress.traefik.enabled":                    "true",
					"ingress.traefik.cassandra.enabled":          "true",
					"ingress.traefik.stargate.cassandra.enabled": "false",
				},
			}

			Expect(renderTemplate(options)).To(Succeed())
			Expect(ingress.Kind).To(Equal("IngressRouteTCP"))

			VerifyTraefikTCPIngressRoute(ingress, "cassandra", "HostSNI(`*`)", Sprintf("%s-%s-service", HelmReleaseName, "dc1"), 9042)
		})

		It("it is enabled and Stargate Cassandra ingress is disabled with release name != cluster name", func() {
			clusterName := Sprintf("k8ssandracluster%s", UniqueIdSuffix)
			options := &helm.Options{
				KubectlOptions: defaultKubeCtlOptions,
				SetValues: map[string]string{
					"cassandra.clusterName":                      clusterName,
					"ingress.traefik.enabled":                    "true",
					"ingress.traefik.cassandra.enabled":          "true",
					"ingress.traefik.stargate.cassandra.enabled": "false",
				},
			}

			Expect(renderTemplate(options)).To(Succeed())
			Expect(ingress.Kind).To(Equal("IngressRouteTCP"))

			VerifyTraefikTCPIngressRoute(ingress, "cassandra", "HostSNI(`*`)", Sprintf("%s-%s-service", clusterName, "dc1"), 9042)
		})
	})
})
