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

package v1alpha1

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func validateJavaRuntime(java *JavaRuntime, className string) []*field.Error {
	var allErrs field.ErrorList
	if java != nil {
		if className == "" {
			e := field.Invalid(field.NewPath("spec").Child("classname"), className, "class name cannot be empty")
			allErrs = append(allErrs, e)
		}
		if java.Jar == "" {
			e := field.Invalid(field.NewPath("spec").Child("java", "jar"), java.Jar,
				"jar cannot be empty in java runtime")
			allErrs = append(allErrs, e)
		}
		if java.JarLocation != "" {
			err := validPackageLocation(java.JarLocation)
			if err != nil {
				e := field.Invalid(field.NewPath("spec").Child("java", "jarLocation"), java.JarLocation, err.Error())
				allErrs = append(allErrs, e)
			}
		}
	}
	return allErrs
}

func validatePythonRuntime(python *PythonRuntime, className string) []*field.Error {
	var allErrs field.ErrorList
	if python != nil {
		if className == "" {
			e := field.Invalid(field.NewPath("spec").Child("classname"), className, "class name cannot be empty")
			allErrs = append(allErrs, e)
		}
		if python.Py == "" {
			e := field.Invalid(field.NewPath("spec").Child("python", "py"), python.Py,
				"py cannot be empty in python runtime")
			allErrs = append(allErrs, e)
		}
		if python.PyLocation != "" {
			err := validPackageLocation(python.PyLocation)
			if err != nil {
				e := field.Invalid(field.NewPath("spec").Child("python", "pyLocation"), python.PyLocation, err.Error())
				allErrs = append(allErrs, e)
			}
		}
	}
	return allErrs
}

func validateGolangRuntime(golang *GoRuntime) []*field.Error {
	var allErrs field.ErrorList
	if golang != nil {
		if golang.Go == "" {
			e := field.Invalid(field.NewPath("spec").Child("golang", "go"), golang.Go,
				"go cannot be empty in golang runtime")
			allErrs = append(allErrs, e)
		}
		if golang.GoLocation != "" {
			err := validPackageLocation(golang.GoLocation)
			if err != nil {
				e := field.Invalid(field.NewPath("spec").Child("golang", "goLocation"), golang.GoLocation, err.Error())
				allErrs = append(allErrs, e)
			}
		}
	}
	return allErrs
}

func validateReplicasAndMaxReplicas(replicas, maxReplicas *int32) []*field.Error {
	var allErrs field.ErrorList
	// TODO: allow 0 replicas, currently hpa's min value has to be 1
	if replicas == nil {
		e := field.Invalid(field.NewPath("spec").Child("replicas"), nil, "replicas cannot be nil")
		allErrs = append(allErrs, e)
	}

	if replicas != nil && *replicas <= 0 {
		e := field.Invalid(field.NewPath("spec").Child("replicas"), *replicas, "replicas cannot be zero or negative")
		allErrs = append(allErrs, e)
	}

	if maxReplicas != nil && replicas != nil && *replicas > *maxReplicas {
		e := field.Invalid(field.NewPath("spec").Child("maxReplicas"), *maxReplicas,
			"maxReplicas must be greater than or equal to replicas")
		allErrs = append(allErrs, e)
	}
	return allErrs
}

func validateResourceRequirement(requirements corev1.ResourceRequirements) *field.Error {
	if !validResourceRequirement(requirements) {
		return field.Invalid(field.NewPath("spec").Child("resources"), requirements, "resource requirement is invalid")
	}
	return nil
}

func validateTimeout(timeout int32, processingGuarantee ProcessGuarantee) *field.Error {
	if timeout != 0 && processingGuarantee == EffectivelyOnce {
		return field.Invalid(field.NewPath("spec").Child("timeout"), timeout,
			"message timeout can only be set for AtleastOnce processing guarantee")
	}
	return nil
}

func validateMaxMessageRetry(maxMessageRetry int32, processingGuarantee ProcessGuarantee,
	deadLetterTopic string) []*field.Error {
	var allErrs field.ErrorList
	if maxMessageRetry > 0 && processingGuarantee == EffectivelyOnce {
		e := field.Invalid(field.NewPath("spec").Child("maxMessageRetry"), maxMessageRetry,
			"MaxMessageRetries and Effectively once are not compatible")
		allErrs = append(allErrs, e)
	}

	if maxMessageRetry <= 0 && deadLetterTopic != "" {
		e := field.Invalid(field.NewPath("spec").Child("maxMessageRetry"), maxMessageRetry,
			"dead letter topic is set but max message retry is set to infinity")
		allErrs = append(allErrs, e)
	}
	return allErrs
}

func validateRetainKeyOrdering(retainKeyOrdering bool, processingGuarantee ProcessGuarantee) *field.Error {
	if retainKeyOrdering && processingGuarantee == EffectivelyOnce {
		return field.Invalid(field.NewPath("spec").Child("retainKeyOrdering"), retainKeyOrdering,
			"when effectively once processing guarantee is specified, retain Key ordering cannot be set")
	}
	return nil
}

func validateRetainOrderingConflicts(retainKeyOrdering bool, retainOrdering bool) []*field.Error {
	var allErrs field.ErrorList
	if retainKeyOrdering && retainOrdering {
		e := field.Invalid(field.NewPath("spec").Child("retainKeyOrdering"), retainKeyOrdering,
			"only one of retain ordering or retain key ordering can be set")
		allErrs = append(allErrs, e)
		e = field.Invalid(field.NewPath("spec").Child("retainOrdering"), retainOrdering,
			"only one of retain ordering or retain key ordering can be set")
		allErrs = append(allErrs, e)
	}
	return allErrs
}

func validateFunctionConfig(config *Config) *field.Error {
	if config != nil {
		_, err := config.MarshalJSON()
		if err != nil {
			return field.Invalid(field.NewPath("spec").Child("funcConfig"), config,
				"function config is invalid: "+err.Error())
		}
	}
	return nil
}

func validateSinkConfig(config *Config) *field.Error {
	if config != nil {
		_, err := config.MarshalJSON()
		if err != nil {
			return field.Invalid(field.NewPath("spec").Child("sinkConfig"), config,
				"sink config is invalid: "+err.Error())
		}
	}
	return nil
}

func validateSourceConfig(config *Config) *field.Error {
	if config != nil {
		_, err := config.MarshalJSON()
		if err != nil {
			return field.Invalid(field.NewPath("spec").Child("sourceConfig"), config,
				"source config is invalid: "+err.Error())
		}
	}
	return nil
}

func validateSecretsMap(secrets map[string]SecretRef) *field.Error {
	if secrets != nil {
		_, err := json.Marshal(secrets)
		if err != nil {
			return field.Invalid(field.NewPath("spec").Child("secretsMap"), secrets,
				"secrets map is invalid: "+err.Error())
		}
	}
	return nil
}

func validateInputOutput(input *InputConf, output *OutputConf) []*field.Error {
	var allErrs field.ErrorList
	allInputTopics := []string{}
	if input != nil {
		allInputTopics = collectAllInputTopics(*input)
		if len(allInputTopics) == 0 {
			e := field.Invalid(field.NewPath("spec").Child("input"), *input,
				"No input topic(s) specified for the function")
			allErrs = append(allErrs, e)
		}

		for _, topic := range allInputTopics {
			err := isValidTopicName(topic)
			if err != nil {
				e := field.Invalid(field.NewPath("spec").Child("input"), *input,
					fmt.Sprintf("Input topic %s is invalid", topic))
				allErrs = append(allErrs, e)
			}
		}

		for topicName, conf := range input.SourceSpecs {
			if conf.ReceiverQueueSize != nil && *conf.ReceiverQueueSize < 0 {
				e := field.Invalid(field.NewPath("spec").Child("input", "sourceSpecs"),
					input.SourceSpecs, fmt.Sprintf("%s receiver queue size should be >= zero", topicName))
				allErrs = append(allErrs, e)
			}

			if conf.CryptoConfig != nil && conf.CryptoConfig.CryptoKeyReaderClassName == "" {
				e := field.Invalid(field.NewPath("spec").Child("input", "sourceSpecs"),
					input.SourceSpecs, fmt.Sprintf("%s cryptoKeyReader class name required", topicName))
				allErrs = append(allErrs, e)
			}
		}
	}

	if output != nil {
		if output.Topic != "" {
			err := isValidTopicName(output.Topic)
			if err != nil {
				e := field.Invalid(field.NewPath("spec").Child("output", "topic"), output.Topic,
					fmt.Sprintf("Output topic %s is invalid", output.Topic))
				allErrs = append(allErrs, e)
			}
			for _, v := range allInputTopics {
				if v == output.Topic {
					e := field.Invalid(field.NewPath("spec").Child("output", "topic"), output.Topic,
						fmt.Sprintf("Output topic %s is also being used as an input topic (topics must be one or the other)",
							output.Topic))
					allErrs = append(allErrs, e)
				}
			}
			if output.ProducerConf != nil && output.ProducerConf.CryptoConfig != nil {
				if output.ProducerConf.CryptoConfig.CryptoKeyReaderClassName == "" {
					e := field.Invalid(field.NewPath("spec").Child("output", "producerConf", "cryptoConfig",
						"cryptoKeyReaderClassName"),
						output.ProducerConf.CryptoConfig.CryptoKeyReaderClassName,
						"cryptoKeyReader class name required")
					allErrs = append(allErrs, e)
				}

				if len(output.ProducerConf.CryptoConfig.EncryptionKeys) == 0 {
					e := field.Invalid(field.NewPath("spec").Child("output", "producerConf", "cryptoConfig",
						"encryptionKeys"),
						output.ProducerConf.CryptoConfig.EncryptionKeys,
						"must provide encryption key name for crypto key reader")
					allErrs = append(allErrs, e)
				}
			}
		}
	}

	return allErrs
}

func validateLogTopic(logTopic string) *field.Error {
	if logTopic != "" {
		err := isValidTopicName(logTopic)
		if err != nil {
			return field.Invalid(field.NewPath("spec").Child("logTopic"), logTopic,
				fmt.Sprintf("Log topic %s is invalid", logTopic))
		}
	}
	return nil
}

func validateDeadLetterTopic(deadLetterTopic string) *field.Error {
	if deadLetterTopic != "" {
		err := isValidTopicName(deadLetterTopic)
		if err != nil {
			return field.Invalid(field.NewPath("spec").Child("deadLetterTopic"), deadLetterTopic,
				fmt.Sprintf("DeadLetter topic %s is invalid", deadLetterTopic))
		}
	}
	return nil
}

func validateAutoAck(autoAck *bool) *field.Error {
	if autoAck == nil {
		return field.Invalid(field.NewPath("spec").Child("autoAck"), autoAck, "autoAck cannot be nil")
	}
	return nil
}

func validateStatefulFunctionConfigs(statefulFunctionConfigs *Stateful, runtime Runtime) *field.Error {
	if statefulFunctionConfigs != nil {
		if statefulFunctionConfigs.Pulsar != nil {
			if isGolangRuntime(runtime) {
				return field.Invalid(field.NewPath("spec").Child("statefulConfig"), runtime.Golang,
					"Golang function do not support stateful function yet")
			}
			if statefulFunctionConfigs.Pulsar.ServiceURL == "" {
				return field.Invalid(field.NewPath("spec").Child("statefulConfig", "pulsar", "serviceUrl"),
					statefulFunctionConfigs.Pulsar.ServiceURL, "serviceUrl cannot be empty")
			}
		}
	}
	return nil
}

func isGolangRuntime(runtime Runtime) bool {
	return runtime.Golang != nil && runtime.Python == nil && runtime.Java == nil
}

func validateWindowConfigs(windowConfig *WindowConfig) *field.Error {
	if windowConfig != nil {
		if windowConfig.WindowLengthDurationMs == nil && windowConfig.WindowLengthCount == nil {
			return field.Invalid(field.NewPath("spec").Child("windowConfig"), windowConfig,
				"Window length is not specified")
		}
		if windowConfig.WindowLengthDurationMs != nil && windowConfig.WindowLengthCount != nil {
			return field.Invalid(field.NewPath("spec").Child("windowConfig"), windowConfig,
				"Window length for time and count are set! Please set one or the other")
		}
		if windowConfig.WindowLengthCount != nil && *windowConfig.WindowLengthCount <= 0 {
			return field.Invalid(field.NewPath("spec").Child("windowConfig"), windowConfig.WindowLengthCount,
				"Window length must be positive")
		}
		if windowConfig.WindowLengthDurationMs != nil && *windowConfig.WindowLengthDurationMs <= 0 {
			return field.Invalid(field.NewPath("spec").Child("windowConfig"), windowConfig.WindowLengthDurationMs,
				"Window length must be positive")
		}
		if windowConfig.SlidingIntervalCount != nil && *windowConfig.SlidingIntervalCount <= 0 {
			return field.Invalid(field.NewPath("spec").Child("windowConfig"), windowConfig.SlidingIntervalCount,
				"Sliding interval must be positive")
		}
		if windowConfig.SlidingIntervalDurationMs != nil && *windowConfig.SlidingIntervalDurationMs <= 0 {
			return field.Invalid(field.NewPath("spec").Child("windowConfig"), windowConfig.SlidingIntervalDurationMs,
				"Sliding interval must be positive")
		}
		if windowConfig.TimestampExtractorClassName != nil {
			if windowConfig.MaxLagMs != nil && *windowConfig.MaxLagMs <= 0 {
				return field.Invalid(field.NewPath("spec").Child("windowConfig"), windowConfig.MaxLagMs,
					"Lag duration must be positive")
			}
			if windowConfig.WatermarkEmitIntervalMs != nil && *windowConfig.WatermarkEmitIntervalMs <= 0 {
				return field.Invalid(field.NewPath("spec").Child("windowConfig"), windowConfig.WatermarkEmitIntervalMs,
					"Watermark interval must be positive")
			}
		}
	}
	return nil
}
