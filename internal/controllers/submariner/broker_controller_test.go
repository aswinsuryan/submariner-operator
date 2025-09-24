/*
SPDX-License-Identifier: Apache-2.0

Copyright Contributors to the Submariner project.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package submariner_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/submariner-io/admiral/pkg/certificate"
	"github.com/submariner-io/admiral/pkg/fake"
	"github.com/submariner-io/admiral/pkg/resource"
	"github.com/submariner-io/admiral/pkg/syncer/broker"
	"github.com/submariner-io/admiral/pkg/syncer/test"
	"github.com/submariner-io/admiral/pkg/util"
	"github.com/submariner-io/submariner-operator/api/v1alpha1"
	submarinerController "github.com/submariner-io/submariner-operator/internal/controllers/submariner"
	controllerTest "github.com/submariner-io/submariner-operator/internal/controllers/test"
	"github.com/submariner-io/submariner-operator/pkg/discovery/globalnet"
	corev1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const brokerName = "test-broker"

var _ = Describe("Broker controller tests", func() {
	t := controllerTest.Driver{
		Namespace:    submarinerNamespace,
		ResourceName: brokerName,
	}

	var fakeDynClient *dynamicfake.FakeDynamicClient

	BeforeEach(func() {
		fakeDynClient = dynamicfake.NewSimpleDynamicClient(scheme.Scheme)

		resource.NewDynamicClient = func(_ *rest.Config) (dynamic.Interface, error) {
			return fakeDynClient, nil
		}

		util.BuildRestMapper = func(_ *rest.Config) (meta.RESTMapper, error) {
			return test.GetRESTMapperFor(&corev1.Secret{}), nil
		}
	})

	var brokerObject *v1alpha1.Broker

	BeforeEach(func() {
		t.BeforeEach()
		brokerObject = &v1alpha1.Broker{
			ObjectMeta: metav1.ObjectMeta{
				Name:      brokerName,
				Namespace: submarinerNamespace,
			},
			Spec: v1alpha1.BrokerSpec{
				GlobalnetCIDRRange:          "168.254.0.0/16",
				DefaultGlobalnetClusterSize: 8192,
				GlobalnetEnabled:            true,
			},
		}

		t.InitScopedClientObjs = []client.Object{brokerObject}
	})

	JustBeforeEach(func() {
		t.JustBeforeEach()

		t.Controller = &submarinerController.BrokerReconciler{
			ScopedClient:  t.ScopedClient,
			GeneralClient: t.GeneralClient,
			Config: &rest.Config{
				Host: "https://test-cluster",
			},
		}
	})

	It("should create the globalnet ConfigMap", func(ctx SpecContext) {
		t.AssertReconcileSuccess(ctx)

		globalnetInfo, _, err := globalnet.GetGlobalNetworks(ctx, t.GeneralClient, submarinerNamespace)
		Expect(err).To(Succeed())
		Expect(globalnetInfo.CIDR).To(Equal(brokerObject.Spec.GlobalnetCIDRRange))
		Expect(globalnetInfo.AllocationSize).To(Equal(brokerObject.Spec.DefaultGlobalnetClusterSize))
	})

	It("should create the CRDs", func(ctx SpecContext) {
		t.AssertReconcileSuccess(ctx)

		crd := &apiextensions.CustomResourceDefinition{}
		Expect(t.GeneralClient.Get(ctx, client.ObjectKey{Name: "clusters.submariner.io"}, crd)).To(Succeed())
		Expect(t.GeneralClient.Get(ctx, client.ObjectKey{Name: "endpoints.submariner.io"}, crd)).To(Succeed())
		Expect(t.GeneralClient.Get(ctx, client.ObjectKey{Name: "gateways.submariner.io"}, crd)).To(Succeed())
		Expect(t.GeneralClient.Get(ctx, client.ObjectKey{Name: "serviceimports.multicluster.x-k8s.io"}, crd)).To(Succeed())
	})

	It("should create the CA certificate secret", func(ctx SpecContext) {
		t.AssertReconcileSuccess(ctx)

		caSecretClient := fakeDynClient.Resource(corev1.SchemeGroupVersion.WithResource("secrets")).Namespace(submarinerNamespace)

		Eventually(func(g Gomega) {
			caSecretObj, err := caSecretClient.Get(ctx, "submariner-ca", metav1.GetOptions{})
			g.Expect(err).NotTo(HaveOccurred())

			caSecret := resource.MustFromUnstructured(caSecretObj, &corev1.Secret{})

			g.Expect(caSecret.Data).To(HaveKey("ca.crt"))
			g.Expect(caSecret.Data).To(HaveKey("ca.key"))
		}).Should(Succeed())
	})

	When("the Broker resource is deleted", func() {
		var signingRequestor certificate.SigningRequestor

		BeforeEach(func() {
			syncerConfig := broker.SyncerConfig{
				LocalNamespace:  "local-ns",
				LocalClusterID:  "east",
				LocalClient:     dynamicfake.NewSimpleDynamicClient(scheme.Scheme),
				BrokerNamespace: brokerObject.Namespace,
				BrokerClient:    fakeDynClient,
				RestMapper:      test.GetRESTMapperFor(&corev1.Secret{}),
			}

			var err error
			stopCh := make(chan struct{})

			signingRequestor, err = certificate.StartSigningRequestor(syncerConfig, stopCh)
			Expect(err).NotTo(HaveOccurred())

			DeferCleanup(func() {
				close(stopCh)
			})
		})

		It("should stop the certificate signer", func(ctx SpecContext) {
			t.AssertReconcileSuccess(ctx)

			Expect(t.ScopedClient.Delete(ctx, brokerObject)).To(Succeed())

			t.AssertReconcileSuccess(ctx)

			signedDataCh := make(chan map[string][]byte, 10)
			onSigned := func(data map[string][]byte) error {
				signedDataCh <- data
				return nil
			}

			Expect(signingRequestor.Issue(ctx, "test-secret", []string{"1.2.3.4"}, onSigned)).To(Succeed())
			Consistently(signedDataCh).ShouldNot(Receive())
		})
	})

	When("the Broker resource doesn't exist", func() {
		BeforeEach(func() {
			t.InitScopedClientObjs = nil
		})

		It("should return success", func(ctx SpecContext) {
			t.AssertReconcileSuccess(ctx)
		})
	})

	Context("Certificate management error handling", func() {
		It("should handle certificate signer creation errors gracefully", func(ctx SpecContext) {
			fake.FailOnAction(&fakeDynClient.Fake, "secrets", "create", nil, false)

			_, err := t.DoReconcile(ctx)
			Expect(err).To(HaveOccurred())
		})
	})
})
