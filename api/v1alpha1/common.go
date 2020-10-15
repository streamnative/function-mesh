package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Messaging struct {
	Pulsar *PulsarMessaging `json:"pulsar,omitempty"`
}

type PulsarMessaging struct {
	// The config map need to contain the following fields
	// webServiceURL
	// brokerServiceURL
	PulsarConfig string `json:"pulsarConfig,omitempty"`
}

type Runtime struct {
	Java   *JavaRuntime   `json:"java,omitempty"`
	Python *PythonRuntime `json:"python,omitempty"`
	Golang *GoRuntime     `json:"golang,omitempty"`
}

type JavaRuntime struct {
	Jar         string `json:"jar,omitempty"`
	JarLocation string `json:"jarLocation,omitempty"`
}

type PythonRuntime struct {
	Py string `json:"py,omitempty"`
}

type GoRuntime struct {
	Go string `json:"go,omitempty"`
}

type SecretRef struct {
	Path string `json:"path,omitempty"`
	Key  string `json:"key,omitempty"`
}

type Component string

const (
	StatefulSet Component = "StatefulSet"
	Service     Component = "Service"
	HPA         Component = "HorizontalPodAutoscaler"
)

// The `Status` of a given `Condition` and the `Action` needed to reach the `Status`
type ResourceCondition struct {
	Condition ResourceConditionType  `json:"condition,omitempty"`
	Status    metav1.ConditionStatus `json:"status,omitempty"`
	Action    ReconcileAction        `json:"action,omitempty"`
}

type ResourceConditionType string

const (
	FunctionReady ResourceConditionType = "FunctionReady"
	SourceReady   ResourceConditionType = "SourceReady"
	SinkReady     ResourceConditionType = "SinkReady"

	StatefulSetReady ResourceConditionType = "StatefulSetReady"
	ServiceReady     ResourceConditionType = "ServiceReady"
	HPAReady         ResourceConditionType = "HPAReady"
)

type ReconcileAction string

const (
	Create   ReconcileAction = "Create"
	Delete   ReconcileAction = "Delete"
	Update   ReconcileAction = "Update"
	Wait     ReconcileAction = "Wait"
	NoAction ReconcileAction = "NoAction"
)

const (
	ATLEAST_ONCE     string = "atleast_once"
	ATMOST_ONCE      string = "atmost_once"
	EFFECTIVELY_ONCE string = "effectively_once"

	DEFAULT_TENANT  string = "public"
	DEFAULT_CLUSTER string = "kubernetes"
)

func validResource(resources corev1.ResourceList) bool {
	// cpu & memory > 0 and storage >= 0
	return resources.Cpu().Sign() == 1 &&
		resources.Memory().Sign() == 1 &&
		resources.Storage().Sign() >= 0
}
