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

package main

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/streamnative/function-mesh/pkg/monitoring"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers"
	"github.com/streamnative/function-mesh/controllers/spec"
	"github.com/streamnative/function-mesh/pkg/webhook"
	"github.com/streamnative/function-mesh/utils"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	runtimeWebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(vpav1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(computev1alpha1.AddToScheme(scheme))

	monitoring.RegisterMetrics()
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr, pprofAddr string
	var leaderElectionID string
	var leaderElectionNamespace string
	var certDir string
	var healthProbeAddr string
	var enableLeaderElection, enablePprof bool
	var configFile string
	var watchedNamespace string
	var enableInitContainers bool
	var globalBackendConfig string
	var globalBackendConfigNamespace string
	var namespacedBackendConfig string
	flag.StringVar(&metricsAddr, "metrics-addr", lookupEnvOrString("METRICS_ADDR", ":8080"),
		"The address the metric endpoint binds to.")
	flag.StringVar(&leaderElectionID, "leader-election-id",
		lookupEnvOrString("LEADER_ELECTION_ID", "a3f45fce.functionmesh.io"),
		"the name of the configmap that leader election will use for holding the leader lock.")
	flag.StringVar(&leaderElectionNamespace, "leader-election-namespace",
		lookupEnvOrString("LEADER_ELECTION_NAMESPACE", ""),
		"the namespace in which the leader election configmap will be created")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", lookupEnvOrBool("ENABLE_LEADER_ELECTION", false),
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&healthProbeAddr, "health-probe-addr", lookupEnvOrString("HEALTH_PROBE_ADDR", ":8000"),
		"The address the healthz/readyz endpoint binds to.")
	flag.StringVar(&certDir, "cert-dir", lookupEnvOrString("CERT_DIR", ""),
		"CertDir is the directory that contains the server key and certificate.\n\tif not set, webhook server would look up the server key and certificate in\n\t{TempDir}/k8s-webhook-server/serving-certs. The server key and certificate\n\tmust be named tls.key and tls.crt, respectively.")
	flag.StringVar(&configFile, "config-file", lookupEnvOrString("CONFIG_FILE", ""),
		"config file path for controller manager")
	flag.StringVar(&watchedNamespace, "watched-namespace", lookupEnvOrString("WATCHED_NAMESPACE", ""),
		"Namespace if specified restricts the manager's cache to watch objects in the desired namespace. Defaults to all namespaces.")
	flag.BoolVar(&enablePprof, "enable-pprof", lookupEnvOrBool("ENABLE_PPROF", false),
		"Enable pprof for controller manager.")
	flag.StringVar(&pprofAddr, "pprof-addr", lookupEnvOrString("PPROF_ADDR", ":8090"),
		"The address the pprof binds to.")
	flag.BoolVar(&enableInitContainers, "enable-init-containers", lookupEnvOrBool("ENABLE_INIT_CONTAINERS", false),
		"Whether to use an init container to download package")
	flag.StringVar(&globalBackendConfig, "global-backend-config", lookupEnvOrString("GLOBAL_BACKEND_CONFIG", ""),
		"The global backend config name used for all functions&sinks&sources")
	flag.StringVar(&globalBackendConfigNamespace, "global-backend-config-namespace", lookupEnvOrString("GLOBAL_BACKEND_CONFIG_NAMESPACE", "default"),
		"The namespace of the global backend config name used for all functions&sinks&sources")
	flag.StringVar(&namespacedBackendConfig, "namespaced-backend-config", lookupEnvOrString("NAMESPACED_BACKEND_CONFIG", "backend-config"),
		"The backend config name used for functions&sinks&sources in the same namespace")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	utils.EnableInitContainers = enableInitContainers
	utils.GlobalBackendConfig = globalBackendConfig
	utils.GlobalBackendConfigNamespace = globalBackendConfigNamespace
	utils.NamespacedBackendConfig = namespacedBackendConfig

	// enable pprof
	if enablePprof {
		go func() {
			if err := http.ListenAndServe(pprofAddr, nil); err != nil {
				setupLog.Error(err, "unable to start pprof")
			}
		}()
	}

	if configFile != "" {
		err := spec.ParseControllerConfigs(configFile)
		if err != nil {
			setupLog.Error(err, "unable to parse the controller configs")
			os.Exit(1)
		}
	}

	options := ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress:  healthProbeAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: leaderElectionNamespace,
		LeaderElectionID:        leaderElectionID,
	}
	if watchedNamespace != "" {
		options.Cache = cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				watchedNamespace: {},
			},
		}
	}
	if certDir != "" {
		options.WebhookServer = &runtimeWebhook.DefaultServer{
			Options: runtimeWebhook.Options{
				Port:    9443,
				CertDir: certDir,
			},
		}
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// allow function mesh to be disabled and enable it by default
	// required because of https://github.com/operator-framework/operator-lifecycle-manager/issues/1523
	if os.Getenv("ENABLE_FUNCTION_MESH_CONTROLLER") != "false" {
		if err = (&controllers.FunctionMeshReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("FunctionMesh"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "FunctionMesh")
			os.Exit(1)
		}
	}
	watchFlags, err := checkGroupVersions(setupLog)
	if err != nil {
		setupLog.Error(err, "failed to check group versions")
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "failed to get rest client")
		os.Exit(1)
	}
	if err = (&controllers.FunctionReconciler{
		Client:            mgr.GetClient(),
		Config:            mgr.GetConfig(),
		RestClient:        clientset.CoreV1().RESTClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Function"),
		Scheme:            mgr.GetScheme(),
		GroupVersionFlags: &watchFlags,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Function")
		os.Exit(1)
	}
	if err = (&controllers.SourceReconciler{
		Client:            mgr.GetClient(),
		Config:            mgr.GetConfig(),
		RestClient:        clientset.CoreV1().RESTClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Source"),
		Scheme:            mgr.GetScheme(),
		GroupVersionFlags: &watchFlags,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Source")
		os.Exit(1)
	}
	if err = (&controllers.SinkReconciler{
		Client:            mgr.GetClient(),
		Config:            mgr.GetConfig(),
		RestClient:        clientset.CoreV1().RESTClient(),
		Log:               ctrl.Log.WithName("controllers").WithName("Sink"),
		Scheme:            mgr.GetScheme(),
		GroupVersionFlags: &watchFlags,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Sink")
		os.Exit(1)
	}

	// enable the webhook service by default
	// Disable function-mesh webhook with `ENABLE_WEBHOOKS=false` when we run locally.
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&webhook.FunctionWebhook{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Function")
			os.Exit(1)
		}
		if err = (&webhook.SourceWebhook{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Source")
			os.Exit(1)
		}
		if err = (&webhook.SinkWebhook{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Sink")
			os.Exit(1)
		}
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// checkGroupVersions will only enable the watch crd params if the related group version
// exists in the cluster
func checkGroupVersions(log logr.Logger) (utils.GroupVersionFlags, error) {
	watchFlags := utils.GroupVersionFlags{}
	client, err := discovery.NewDiscoveryClientForConfig(config.GetConfigOrDie())
	if err != nil {
		return watchFlags, err
	}

	groupVersions, err := utils.GetGroupVersions(client)
	if err != nil {
		return watchFlags, err
	}

	if groupVersions.HasGroupVersions(utils.GroupVersionsVPA) {
		log.Info("API group versions exists, watch vpa crd", "group versions",
			utils.GroupVersionsVPA)
		watchFlags.WatchVPACRDs = true
	}

	if groupVersions.HasGroupVersions(utils.GroupVersionAutoscalingV2) {
		watchFlags.APIAutoscalingGroupVersion = utils.GroupVersionV2
	} else if groupVersions.HasGroupVersions(utils.GroupVersionAutoscalingV2Beta2) {
		watchFlags.APIAutoscalingGroupVersion = utils.GroupVersionV2Beta2
	} else {
		return watchFlags, errors.New("autoscaling/v2beta2 or autoscaling/v2beta1 api group versions not found")
	}
	log.Info("API group versions exists ", "autoscaling group versions",
		watchFlags.APIAutoscalingGroupVersion)
	return watchFlags, nil
}

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			setupLog.Error(err, "unable to convert env: %s to bool", key)
		}
		return v
	}
	return defaultVal
}
