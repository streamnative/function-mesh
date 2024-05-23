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

package spec

import (
	"strings"
	"testing"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/utils"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var yamlVal = string(v1alpha1.YAML)

func TestGetDownloadCommand(t *testing.T) {
	testOauth2 := &v1alpha1.OAuth2Config{
		Audience:      "test-audience",
		IssuerURL:     "test-issuer-url",
		KeySecretName: "test-private-key",
		KeySecretKey:  "auth.json",
	}
	testData := []struct {
		downloadPath     string
		componentPackage string
		tlsConfig        *v1alpha1.PulsarTLSConfig
		oauth2Config     *v1alpha1.OAuth2Config
		genericAuth      *v1alpha1.GenericAuth
		expectedCommand  []string
		hasPulsarctl     bool
	}{
		// test get the download command with package name
		{"function://public/default/test@v1", "function-package.jar", nil, nil, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"--admin-service-url", "$webServiceURL",
				"packages", "download", "function://public/default/test@v1", "--path", "function-package.jar",
			},
			true,
		},
		{"sink://public/default/test@v1", "sink-package.jar", nil, nil, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"--admin-service-url", "$webServiceURL",
				"packages", "download", "sink://public/default/test@v1", "--path", "sink-package.jar",
			},
			true,
		},
		{"source://public/default/test@v1", "source-package.jar", nil, nil, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"--admin-service-url", "$webServiceURL",
				"packages", "download", "source://public/default/test@v1", "--path", "source-package.jar",
			},
			true,
		},
		// test get the download command with normal name
		{"/test", "test.jar", nil, nil, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"--admin-service-url", "$webServiceURL",
				"functions", "download", "--path", "/test", "--destination-file", "test.jar",
			},
			true,
		},
		// test get the download command with a wrong package name
		{"source/public/default/test@v1", "source-package.jar", nil, nil, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"--admin-service-url", "$webServiceURL",
				"functions", "download", "--path", "source/public/default/test@v1", "--destination-file", "source-package.jar",
			},
			true,
		},
		{"source:/public/default/test@v1", "source-package.jar", nil, nil, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"--admin-service-url", "$webServiceURL",
				"functions", "download", "--path", "source:/public/default/test@v1", "--destination-file", "source-package.jar",
			},
			true,
		},
		// test get the download command with an oauth2 config
		{"function://public/default/test@v1", "function-package.jar", nil,
			&v1alpha1.OAuth2Config{
				Audience:      "test-audience",
				IssuerURL:     "test-issuer-url",
				KeySecretName: "test-private-key",
				KeySecretKey:  "auth.json",
			}, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"context",
				"set",
				"downloader",
				"--admin-service-url",
				"$webServiceURL",
				"--issuer-endpoint",
				"test-issuer-url",
				"--audience",
				"test-audience",
				"--key-file",
				"/etc/oauth2/auth.json",
				"&& " + PulsarctlExecutableFile,
				"oauth2",
				"activate",
				"&& " + PulsarctlExecutableFile,
				"packages", "download", "function://public/default/test@v1", "--path", "function-package.jar",
			},
			true,
		},
		// test get the download command with a tls config
		{"function://public/default/test@v1", "function-package.jar",
			&v1alpha1.PulsarTLSConfig{
				Enabled:              true,
				AllowInsecure:        false,
				HostnameVerification: true,
				CertSecretName:       "test-secret",
				CertSecretKey:        "test-key",
			}, nil, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"--admin-service-url", "$webServiceURL",
				"--tls-allow-insecure=false",
				"--tls-enable-hostname-verification=true",
				"--tls-trust-cert-path", "/etc/tls/pulsar-functions/test-key",
				"packages", "download", "function://public/default/test@v1", "--path", "function-package.jar",
			},
			true,
		},
		{"function://public/default/test@v1", "function-package.jar",
			&v1alpha1.PulsarTLSConfig{
				Enabled:              false,
				AllowInsecure:        false,
				HostnameVerification: true,
				CertSecretName:       "test-secret",
				CertSecretKey:        "test-key",
			}, nil, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"--admin-service-url", "$webServiceURL",
				"packages", "download", "function://public/default/test@v1", "--path", "function-package.jar",
			},
			true,
		},
		{"function://public/default/test@v1", "function-package.jar",
			&v1alpha1.PulsarTLSConfig{
				Enabled:              true,
				AllowInsecure:        true,
				HostnameVerification: false,
				CertSecretName:       "test-secret",
				CertSecretKey:        "test-key",
			}, nil, nil,
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				PulsarctlExecutableFile,
				"--admin-service-url", "$webServiceURL",
				"--tls-allow-insecure=true",
				"--tls-enable-hostname-verification=false",
				"--tls-trust-cert-path", "/etc/tls/pulsar-functions/test-key",
				"packages", "download", "function://public/default/test@v1", "--path", "function-package.jar",
			},
			true,
		},
		{"function://public/default/test@v1", "function-package.jar",
			&v1alpha1.PulsarTLSConfig{
				Enabled:              true,
				AllowInsecure:        true,
				HostnameVerification: false,
				CertSecretName:       "test-secret",
				CertSecretKey:        "test-key",
			}, nil, &v1alpha1.GenericAuth{
				ClientAuthenticationParameters: "auth-params",
				ClientAuthenticationPlugin:     "auth-plugin",
			},
			[]string{
				"export PATH=$PATH:/pulsar/bin && ",
				"( " + PulsarctlExecutableFile,
				"oauth2",
				"activate",
				"--auth-params",
				"'auth-params'",
				"|| true ) &&",
				PulsarctlExecutableFile,
				"--auth-plugin", "auth-plugin",
				"--auth-params", "'auth-params'",
				"--admin-service-url", "$webServiceURL",
				"--tls-allow-insecure=true",
				"--tls-enable-hostname-verification=false",
				"--tls-trust-cert-path", "/etc/tls/pulsar-functions/test-key",
				"packages", "download", "function://public/default/test@v1", "--path", "function-package.jar",
			},
			true,
		},
		{"http://aaa.bbb.ccc/test.jar", "function-package.jar",
			&v1alpha1.PulsarTLSConfig{
				Enabled:              true,
				AllowInsecure:        true,
				HostnameVerification: false,
				CertSecretName:       "test-secret",
				CertSecretKey:        "test-key",
			}, nil, nil,
			[]string{
				"wget",
				"http://aaa.bbb.ccc/test.jar",
				"-O",
				"function-package.jar",
			},
			true,
		},
		{"function://public/default/test@v1", "function-package.jar", nil, testOauth2, nil,
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"--auth-plugin", OAuth2AuthenticationPlugin,
				"--auth-params", testOauth2.AuthenticationParameters(),
				"packages", "download", "function://public/default/test@v1", "--path", "function-package.jar",
			},
			false,
		},
		{"sink://public/default/test@v1", "sink-package.jar", nil, nil, &v1alpha1.GenericAuth{
			ClientAuthenticationParameters: "auth-params",
			ClientAuthenticationPlugin:     "auth-plugin",
		},
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"--auth-plugin", "auth-plugin",
				"--auth-params", "'auth-params'",
				"packages", "download", "sink://public/default/test@v1", "--path", "sink-package.jar",
			},
			false,
		},
		{"source://public/default/test@v1", "source-package.jar", &v1alpha1.PulsarTLSConfig{
			Enabled:              true,
			AllowInsecure:        false,
			HostnameVerification: true,
			CertSecretName:       "test-secret",
			CertSecretKey:        "test-key",
		}, nil, nil,
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"--tls-enable-hostname-verification",
				"--tls-trust-cert-path", "/etc/tls/pulsar-functions/test-key",
				"packages", "download", "source://public/default/test@v1", "--path", "source-package.jar",
			},
			false,
		},
		// test get the download command with normal name
		{"/test", "test.jar", &v1alpha1.PulsarTLSConfig{
			Enabled:              true,
			AllowInsecure:        true,
			HostnameVerification: true,
			CertSecretName:       "test-secret",
			CertSecretKey:        "test-key",
		}, nil, nil,
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"--tls-allow-insecure", "--tls-enable-hostname-verification",
				"--tls-trust-cert-path", "/etc/tls/pulsar-functions/test-key",
				"functions", "download", "--path", "/test", "--destination-file", "test.jar",
			},
			false,
		},
		// test get the download command with a wrong package name
		{"source/public/default/test@v1", "source-package.jar", &v1alpha1.PulsarTLSConfig{
			Enabled:              true,
			AllowInsecure:        true,
			HostnameVerification: false,
			CertSecretName:       "test-secret",
			CertSecretKey:        "test-key",
		}, nil, nil,
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"--tls-allow-insecure",
				"--tls-trust-cert-path", "/etc/tls/pulsar-functions/test-key",
				"functions", "download", "--path", "source/public/default/test@v1", "--destination-file", "source-package.jar",
			},
			false,
		},
		{"source:/public/default/test@v1", "source-package.jar", &v1alpha1.PulsarTLSConfig{
			Enabled:              false,
			AllowInsecure:        true,
			HostnameVerification: false,
			CertSecretName:       "test-secret",
			CertSecretKey:        "test-key",
		}, nil, nil,
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"functions", "download", "--path", "source:/public/default/test@v1", "--destination-file", "source-package.jar",
			},
			false,
		},
	}

	for _, v := range testData {
		var authConfig v1alpha1.AuthConfig
		if v.oauth2Config != nil {
			authConfig.OAuth2Config = v.oauth2Config
		} else if v.genericAuth != nil {
			authConfig.GenericAuth = v.genericAuth
		}
		actualResult := GetDownloadCommand(v.downloadPath, v.componentPackage, v.hasPulsarctl, v.hasPulsarctl, false, false, v.tlsConfig, &authConfig)
		assert.Equal(t, v.expectedCommand, actualResult)
	}
}

func TestGetFunctionRunnerImage(t *testing.T) {
	javaRuntime := v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "test",
	}}
	image := getFunctionRunnerImage(&v1alpha1.FunctionSpec{Runtime: javaRuntime})
	assert.Equal(t, image, DefaultJavaRunnerImage)

	pythonRuntime := v1alpha1.Runtime{Python: &v1alpha1.PythonRuntime{
		Py:         "test.py",
		PyLocation: "test",
	}}
	image = getFunctionRunnerImage(&v1alpha1.FunctionSpec{Runtime: pythonRuntime})
	assert.Equal(t, image, DefaultPythonRunnerImage)

	goRuntime := v1alpha1.Runtime{Golang: &v1alpha1.GoRuntime{
		Go:         "test",
		GoLocation: "test",
	}}
	image = getFunctionRunnerImage(&v1alpha1.FunctionSpec{Runtime: goRuntime})
	assert.Equal(t, image, DefaultGoRunnerImage)
}

func TestGetSinkRunnerImage(t *testing.T) {
	spec := v1alpha1.SinkSpec{Runtime: v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "",
	}}}
	image := getSinkRunnerImage(&spec)
	assert.Equal(t, image, DefaultRunnerImage)

	spec = v1alpha1.SinkSpec{Runtime: v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "test",
	}}}
	image = getSinkRunnerImage(&spec)
	assert.Equal(t, image, DefaultRunnerImage)

	spec = v1alpha1.SinkSpec{Runtime: v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "sink://public/default/test",
	}}}
	image = getSinkRunnerImage(&spec)
	assert.Equal(t, image, DefaultJavaRunnerImage)

	spec = v1alpha1.SinkSpec{Runtime: v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "",
	}}, Image: "streamnative/pulsar-io-test:2.7.1"}
	image = getSinkRunnerImage(&spec)
	assert.Equal(t, image, "streamnative/pulsar-io-test:2.7.1")
}

func TestGetSourceRunnerImage(t *testing.T) {
	spec := v1alpha1.SourceSpec{Runtime: v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "",
	}}}
	image := getSourceRunnerImage(&spec)
	assert.Equal(t, image, DefaultRunnerImage)

	spec = v1alpha1.SourceSpec{Runtime: v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "test",
	}}}
	image = getSourceRunnerImage(&spec)
	assert.Equal(t, image, DefaultRunnerImage)

	spec = v1alpha1.SourceSpec{Runtime: v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "sink://public/default/test",
	}}}
	image = getSourceRunnerImage(&spec)
	assert.Equal(t, image, DefaultJavaRunnerImage)

	spec = v1alpha1.SourceSpec{Runtime: v1alpha1.Runtime{Java: &v1alpha1.JavaRuntime{
		Jar:         "test.jar",
		JarLocation: "",
	}}, Image: "streamnative/pulsar-io-test:2.7.1"}
	image = getSourceRunnerImage(&spec)
	assert.Equal(t, image, "streamnative/pulsar-io-test:2.7.1")
}

func TestMakeGoFunctionCommand(t *testing.T) {
	function := makeGoFunctionSample(TestFunctionName)
	commands := MakeGoFunctionCommand("", "/pulsar/go-func", function)
	assert.Equal(t, commands[0], "bash")
	assert.Equal(t, commands[1], "-c")
	assert.True(t, strings.HasPrefix(commands[2], "SHARD_ID=${POD_NAME##*-} && echo shardId=${SHARD_ID}"))
	innerCommands := strings.Split(commands[2], "&&")
	assert.Equal(t, innerCommands[0], "SHARD_ID=${POD_NAME##*-} ")
	assert.Equal(t, innerCommands[1], " echo shardId=${SHARD_ID} ")
	assert.True(t, strings.HasPrefix(innerCommands[2], " GO_FUNCTION_CONF"))
	assert.Equal(t, innerCommands[3], " goFunctionConfigs=${GO_FUNCTION_CONF} ")
	assert.Equal(t, innerCommands[4], " echo goFunctionConfigs=\"'${goFunctionConfigs}'\" ")
	assert.Equal(t, innerCommands[5], " ls -l /pulsar/go-func ")
	assert.Equal(t, innerCommands[6], " chmod +x /pulsar/go-func ")
	assert.Equal(t, innerCommands[7], " exec /pulsar/go-func -instance-conf ${goFunctionConfigs}")
}

const TestClusterName string = "test-pulsar"
const TestFunctionName string = "test-function"
const TestNameSpace string = "default"

func makeSampleObjectMeta(name string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      name,
		Namespace: TestNameSpace,
		UID:       "dead-beef", // uid not generate automatically with fake k8s
	}
}

func makeGoFunctionSample(functionName string) *v1alpha1.Function {
	maxPending := int32(1000)
	replicas := int32(1)
	minReplicas := int32(1)
	maxReplicas := int32(5)
	trueVal := true
	return &v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "compute.functionmesh.io/v1alpha1",
		},
		ObjectMeta: *makeSampleObjectMeta(functionName),
		Spec: v1alpha1.FunctionSpec{
			Name:        functionName,
			Tenant:      "public",
			ClusterName: TestClusterName,
			Input: v1alpha1.InputConf{
				Topics: []string{
					"persistent://public/default/go-function-input-topic",
				},
			},
			Output: v1alpha1.OutputConf{
				Topic: "persistent://public/default/go-function-output-topic",
			},
			LogTopic:                     "persistent://public/default/go-function-logs",
			Timeout:                      0,
			MaxMessageRetry:              0,
			ForwardSourceMessageProperty: &trueVal,
			Replicas:                     &replicas,
			MinReplicas:                  &minReplicas,
			MaxReplicas:                  &maxReplicas,
			AutoAck:                      &trueVal,
			MaxPendingAsyncRequests:      &maxPending,
			Messaging: v1alpha1.Messaging{
				Pulsar: &v1alpha1.PulsarMessaging{
					PulsarConfig: TestClusterName,
				},
			},
			Runtime: v1alpha1.Runtime{
				Golang: &v1alpha1.GoRuntime{
					Go: "/pulsar/go-func",
				},
			},
		},
	}
}

func TestGeneratePodVolumes(t *testing.T) {
	type args struct {
		podVolumes    []corev1.Volume
		producerConf  *v1alpha1.ProducerConfig
		consumerConfs map[string]v1alpha1.ConsumerConfig
		trustCert     *v1alpha1.PulsarTLSConfig
		authConfig    *v1alpha1.AuthConfig
		logConf       map[int32]*v1alpha1.RuntimeLogConfig
	}
	tests := []struct {
		name string
		args args
		want []corev1.Volume
	}{
		{
			name: "keep given pod volumes",
			args: args{
				podVolumes: []corev1.Volume{{
					Name: "test-volume",
				}},
			},
			want: []corev1.Volume{{
				Name: "test-volume",
			}},
		},
		{
			name: "generate pod volumes from producer conf",
			args: args{
				producerConf: &v1alpha1.ProducerConfig{
					CryptoConfig: &v1alpha1.CryptoConfig{
						CryptoSecrets: []v1alpha1.CryptoSecret{{
							SecretName: "test-producer-secret",
							SecretKey:  "test-producer-key",
						}},
					},
				},
			},
			want: []corev1.Volume{{
				Name: "test-producer-secret-test-producer-key",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: "test-producer-secret",
						Items: []corev1.KeyToPath{
							{
								Key:  "test-producer-key",
								Path: "test-producer-key",
							},
						},
					},
				},
			}},
		},
		{
			name: "generate pod volumes from consumer conf",
			args: args{
				podVolumes: nil,
				producerConf: &v1alpha1.ProducerConfig{
					CryptoConfig: &v1alpha1.CryptoConfig{
						CryptoSecrets: []v1alpha1.CryptoSecret{{
							SecretName: "test-producer-secret",
							SecretKey:  "test-producer-key",
						}},
					},
				},
				consumerConfs: map[string]v1alpha1.ConsumerConfig{
					"test-consumer": {
						CryptoConfig: &v1alpha1.CryptoConfig{
							CryptoSecrets: []v1alpha1.CryptoSecret{{
								SecretName: "test-consumer-secret",
								SecretKey:  "test-consumer-key",
							}},
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "test-producer-secret-test-producer-key",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-producer-secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "test-producer-key",
									Path: "test-producer-key",
								},
							},
						},
					},
				},
				{
					Name: "test-consumer-secret-test-consumer-key",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-consumer-secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "test-consumer-key",
									Path: "test-consumer-key",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "generate pod volumes from trust cert",
			args: args{
				podVolumes: nil,
				producerConf: &v1alpha1.ProducerConfig{
					CryptoConfig: &v1alpha1.CryptoConfig{
						CryptoSecrets: []v1alpha1.CryptoSecret{{
							SecretName: "test-producer-secret",
							SecretKey:  "test-producer-key",
						}},
					},
				},
				consumerConfs: map[string]v1alpha1.ConsumerConfig{
					"test-consumer": {
						CryptoConfig: &v1alpha1.CryptoConfig{
							CryptoSecrets: []v1alpha1.CryptoSecret{{
								SecretName: "test-consumer-secret",
								SecretKey:  "test-consumer-key",
							}},
						},
					},
				},
				trustCert: &v1alpha1.PulsarTLSConfig{
					Enabled:              true,
					AllowInsecure:        true,
					HostnameVerification: true,
					CertSecretName:       "test-trust-secret",
					CertSecretKey:        "test-trust-key",
				},
			},
			want: []corev1.Volume{
				{
					Name: "test-trust-secret-test-trust-key",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-trust-secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "test-trust-key",
									Path: "test-trust-key",
								},
							},
						},
					},
				},
				{
					Name: "test-producer-secret-test-producer-key",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-producer-secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "test-producer-key",
									Path: "test-producer-key",
								},
							},
						},
					},
				},
				{
					Name: "test-consumer-secret-test-consumer-key",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-consumer-secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "test-consumer-key",
									Path: "test-consumer-key",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "generate pod volumes from authConfig",
			args: args{
				authConfig: &v1alpha1.AuthConfig{
					OAuth2Config: &v1alpha1.OAuth2Config{
						Audience:      "test-aud",
						IssuerURL:     "test-issuer",
						KeySecretName: "test-private-key-secret",
						KeySecretKey:  "test-private-key",
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "test-private-key-secret-test-private-key",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-private-key-secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "test-private-key",
									Path: "test-private-key",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "generate pod volumes from runtime log configs",
			args: args{
				logConf: map[int32]*v1alpha1.RuntimeLogConfig{
					javaRuntimeLog: {
						LogConfig: &v1alpha1.LogConfig{
							Name: "test-log-config",
							Key:  "java-xml",
						},
					},
					pythonRuntimeLog: {
						LogConfig: &v1alpha1.LogConfig{
							Name: "test-log-config",
							Key:  "python-ini",
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "test-log-config-java-log-conf",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-log-config",
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "java-xml",
									Path: "java_instance_log4j.xml",
								},
							},
						},
					},
				},
				{
					Name: "test-log-config-python-log-conf",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-log-config",
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "python-ini",
									Path: "python_instance_logging.ini",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "generate pod volumes from runtime log configs with log4j type",
			args: args{
				logConf: map[int32]*v1alpha1.RuntimeLogConfig{
					javaRuntimeLog: {
						JavaLog4JConfigFileType: (*v1alpha1.JavaLog4JConfigFileType)(&yamlVal),
						LogConfig: &v1alpha1.LogConfig{
							Name: "test-log-config",
							Key:  "java-yaml",
						},
					},
					pythonRuntimeLog: {
						LogConfig: &v1alpha1.LogConfig{
							Name: "test-log-config",
							Key:  "python-ini",
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "test-log-config-java-log-conf",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-log-config",
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "java-yaml",
									Path: "java_instance_log4j.yaml",
								},
							},
						},
					},
				},
				{
					Name: "test-log-config-python-log-conf",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-log-config",
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "python-ini",
									Path: "python_instance_logging.ini",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want,
				GeneratePodVolumes(
					tt.args.podVolumes,
					tt.args.producerConf,
					tt.args.consumerConfs,
					tt.args.trustCert,
					tt.args.authConfig,
					tt.args.logConf,
					v1alpha1.RUNTIME,
				), "GeneratePodVolumes(%v, %v, %v, %v)", tt.args.podVolumes, tt.args.producerConf, tt.args.consumerConfs, tt.args.trustCert)
		})
	}
}

func TestGenerateContainerVolumeMounts(t *testing.T) {
	utils.EnableInitContainers = true
	type args struct {
		volumeMounts  []corev1.VolumeMount
		producerConf  *v1alpha1.ProducerConfig
		consumerConfs map[string]v1alpha1.ConsumerConfig
		trustCert     *v1alpha1.PulsarTLSConfig
		authConfig    *v1alpha1.AuthConfig
		logConf       map[int32]*v1alpha1.RuntimeLogConfig
		javaRuntime   *v1alpha1.JavaRuntime
	}
	tests := []struct {
		name string
		args args
		want []corev1.VolumeMount
	}{
		{
			name: "generate volume mounts from producerConf",
			args: args{
				producerConf: &v1alpha1.ProducerConfig{
					CryptoConfig: &v1alpha1.CryptoConfig{
						CryptoSecrets: []v1alpha1.CryptoSecret{{
							SecretName: "test-producer-secret",
							SecretKey:  "test-producer-key",
							AsVolume:   "/test-producer",
						}},
					},
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "test-producer-secret-test-producer-key",
					MountPath: "/test-producer",
				},
			},
		},
		{
			name: "not generate volume mounts if AsVolume is not set",
			args: args{
				producerConf: &v1alpha1.ProducerConfig{
					CryptoConfig: &v1alpha1.CryptoConfig{
						CryptoSecrets: []v1alpha1.CryptoSecret{{
							SecretName: "test-producer-secret",
							SecretKey:  "test-producer-key",
						}},
					},
				},
			},
			want: []corev1.VolumeMount{},
		},
		{
			name: "generate volume mounts from consumerConf",
			args: args{
				producerConf: &v1alpha1.ProducerConfig{
					CryptoConfig: &v1alpha1.CryptoConfig{
						CryptoSecrets: []v1alpha1.CryptoSecret{{
							SecretName: "test-producer-secret",
							SecretKey:  "test-producer-key",
							AsVolume:   "/test-producer",
						}},
					},
				},
				consumerConfs: map[string]v1alpha1.ConsumerConfig{
					"test-consumer": {
						CryptoConfig: &v1alpha1.CryptoConfig{
							CryptoSecrets: []v1alpha1.CryptoSecret{{
								SecretName: "test-consumer-secret",
								SecretKey:  "test-consumer-key",
								AsVolume:   "/test-consumer",
							}},
						},
					},
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "test-producer-secret-test-producer-key",
					MountPath: "/test-producer",
				},
				{
					Name:      "test-consumer-secret-test-consumer-key",
					MountPath: "/test-consumer",
				},
			},
		},
		{
			name: "generate volume mounts from trustCert",
			args: args{
				producerConf: &v1alpha1.ProducerConfig{
					CryptoConfig: &v1alpha1.CryptoConfig{
						CryptoSecrets: []v1alpha1.CryptoSecret{{
							SecretName: "test-producer-secret",
							SecretKey:  "test-producer-key",
							AsVolume:   "/test-producer",
						}},
					},
				},
				consumerConfs: map[string]v1alpha1.ConsumerConfig{
					"test-consumer": {
						CryptoConfig: &v1alpha1.CryptoConfig{
							CryptoSecrets: []v1alpha1.CryptoSecret{{
								SecretName: "test-consumer-secret",
								SecretKey:  "test-consumer-key",
								AsVolume:   "/test-consumer",
							}},
						},
					},
				},
				trustCert: &v1alpha1.PulsarTLSConfig{
					Enabled:              true,
					AllowInsecure:        true,
					HostnameVerification: true,
					CertSecretName:       "test-trust-secret",
					CertSecretKey:        "test-trust-key",
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "test-trust-secret-test-trust-key",
					MountPath: "/etc/tls/pulsar-functions",
				},
				{
					Name:      "test-producer-secret-test-producer-key",
					MountPath: "/test-producer",
				},
				{
					Name:      "test-consumer-secret-test-consumer-key",
					MountPath: "/test-consumer",
				},
			},
		},
		{
			name: "generate volume mounts from runtime config",
			args: args{
				producerConf: &v1alpha1.ProducerConfig{
					CryptoConfig: &v1alpha1.CryptoConfig{
						CryptoSecrets: []v1alpha1.CryptoSecret{{
							SecretName: "test-producer-secret",
							SecretKey:  "test-producer-key",
							AsVolume:   "/test-producer",
						}},
					},
				},
				consumerConfs: map[string]v1alpha1.ConsumerConfig{
					"test-consumer": {
						CryptoConfig: &v1alpha1.CryptoConfig{
							CryptoSecrets: []v1alpha1.CryptoSecret{{
								SecretName: "test-consumer-secret",
								SecretKey:  "test-consumer-key",
								AsVolume:   "/test-consumer",
							}},
						},
					},
				},
				trustCert: &v1alpha1.PulsarTLSConfig{
					Enabled:              true,
					AllowInsecure:        true,
					HostnameVerification: true,
					CertSecretName:       "test-trust-secret",
					CertSecretKey:        "test-trust-key",
				},
				javaRuntime: &v1alpha1.JavaRuntime{
					Jar:         "test.jar",
					JarLocation: "/test-jar-location",
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "test-trust-secret-test-trust-key",
					MountPath: "/etc/tls/pulsar-functions",
				},
				{
					Name:      "test-producer-secret-test-producer-key",
					MountPath: "/test-producer",
				},
				{
					Name:      "test-consumer-secret-test-consumer-key",
					MountPath: "/test-consumer",
				},
			},
		},
		{
			name: "generate volume mounts from authSecret",
			args: args{
				authConfig: &v1alpha1.AuthConfig{
					OAuth2Config: &v1alpha1.OAuth2Config{
						Audience:      "test-aud",
						IssuerURL:     "test-issuer",
						KeySecretName: "test-private-key-secret",
						KeySecretKey:  "test-private-key",
					},
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "test-private-key-secret-test-private-key",
					MountPath: "/etc/oauth2",
				},
			},
		},
		{
			name: "generate volume mounts from runtime log config",
			args: args{
				logConf: map[int32]*v1alpha1.RuntimeLogConfig{
					javaRuntimeLog: {
						LogConfig: &v1alpha1.LogConfig{
							Name: "test-log-config",
							Key:  "java-xml",
						},
					},
					pythonRuntimeLog: {
						LogConfig: &v1alpha1.LogConfig{
							Name: "test-log-config",
							Key:  "python-ini",
						},
					},
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "test-log-config-java-log-conf",
					MountPath: "/pulsar/conf/java-log/",
				},
				{
					Name:      "test-log-config-python-log-conf",
					MountPath: "/pulsar/conf/python-log/",
				},
			},
		},
		{
			name: "generate volume mounts from runtime log config with logj4 types",
			args: args{
				logConf: map[int32]*v1alpha1.RuntimeLogConfig{
					javaRuntimeLog: {
						JavaLog4JConfigFileType: (*v1alpha1.JavaLog4JConfigFileType)(&yamlVal),
						LogConfig: &v1alpha1.LogConfig{
							Name: "test-log-config",
							Key:  "java-yaml",
						},
					},
					pythonRuntimeLog: {
						LogConfig: &v1alpha1.LogConfig{
							Name: "test-log-config",
							Key:  "python-ini",
						},
					},
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "test-log-config-java-log-conf",
					MountPath: "/pulsar/conf/java-log/",
				},
				{
					Name:      "test-log-config-python-log-conf",
					MountPath: "/pulsar/conf/python-log/",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want,
				GenerateContainerVolumeMounts(
					tt.args.volumeMounts,
					tt.args.producerConf,
					tt.args.consumerConfs,
					tt.args.trustCert,
					tt.args.authConfig,
					tt.args.logConf,
					v1alpha1.RUNTIME), "GenerateContainerVolumeMounts(%v, %v, %v, %v)", tt.args.volumeMounts, tt.args.producerConf, tt.args.consumerConfs, tt.args.trustCert)
		})
	}
}
