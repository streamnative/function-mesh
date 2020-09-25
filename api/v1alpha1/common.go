package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Runtime struct {
	Pulsar *PulsarRuntime `json:"pulsar,omitempty"`
}

type PulsarRuntime struct {
	PulsarConfig string `json:"pulsarConfig,omitempty"`
	// The config map need to contain the following fields
	// webServiceURL
	// brokerServiceURL
}

type Component string

const (
	StatefulSet Component = "StatefulSet"
	Service     Component = "Service"
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
)

type ReconcileAction string

const (
	Create   ReconcileAction = "Create"
	Delete   ReconcileAction = "Delete"
	Update   ReconcileAction = "Update"
	Wait     ReconcileAction = "Wait"
	NoAction ReconcileAction = "NoAction"
)
