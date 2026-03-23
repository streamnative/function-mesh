// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package controllers

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/utils"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var sourceReconciler *SourceReconciler
var sinkReconciler *SinkReconciler
var funcReconciler *FunctionReconciler
var testEnv *envtest.Environment
var testVpaCRDDir string

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	vpaCRDDir, err := vpaCRDDirectory()
	Expect(err).ToNot(HaveOccurred())
	testVpaCRDDir = vpaCRDDir
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			vpaCRDDir,
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = scheme.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = vpav1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		Metrics: server.Options{
			BindAddress: ":8090",
		},
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&FunctionMeshReconciler{
		Client: k8sManager.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("FunctionMesh"),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	gvFlags := &utils.GroupVersionFlags{
		WatchVPACRDs:               true,
		APIAutoscalingGroupVersion: utils.GroupVersionV2,
	}
	funcReconciler = &FunctionReconciler{
		Client:            k8sManager.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Function"),
		Scheme:            k8sManager.GetScheme(),
		GroupVersionFlags: gvFlags,
	}
	err = funcReconciler.SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	sourceReconciler = &SourceReconciler{
		Client:            k8sManager.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Source"),
		Scheme:            k8sManager.GetScheme(),
		GroupVersionFlags: gvFlags,
	}
	err = sourceReconciler.SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	sinkReconciler = &SinkReconciler{
		Client:            k8sManager.GetClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Sink"),
		Scheme:            k8sManager.GetScheme(),
		GroupVersionFlags: gvFlags,
	}
	err = sinkReconciler.SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
		gexec.KillAndWait(4 * time.Second)

		// Teardown the test environment once controller is fnished.
		// Otherwise from Kubernetes 1.21+, teardon timeouts waiting on
		// kube-apiserver to return
		err := testEnv.Stop()
		Expect(err).ToNot(HaveOccurred())
		if testVpaCRDDir != "" {
			Expect(os.RemoveAll(testVpaCRDDir)).To(Succeed())
		}
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

func vpaCRDDirectory() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "k8s.io/autoscaler/vertical-pod-autoscaler")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	sourcePath := filepath.Join(strings.TrimSpace(string(output)), "deploy", "vpa-v1-crd-gen.yaml")
	crdContents, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp("", "function-mesh-vpa-crd-*")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath.Join(tempDir, filepath.Base(sourcePath)), crdContents, 0o600); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}
