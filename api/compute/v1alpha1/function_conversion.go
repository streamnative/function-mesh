package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/streamnative/function-mesh/api/compute/v1alpha2"
	apispec "github.com/streamnative/function-mesh/pkg/spec"
)

// ConvertTo converts this Function to the Hub version (v1alpha2).
func (src *Function) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.Function)
	dst.ObjectMeta = src.ObjectMeta
	if err := src.Spec.convertSpecTo(&dst.Spec); err != nil {
		return err
	}
	if err := src.Status.convertStatusTo(&dst.Status); err != nil {
		return err
	}
	return nil
}

func (src *FunctionSpec) convertSpecTo(dst *v1alpha2.FunctionSpec) error {
	dst.Name = src.Name
	dst.ClassName = src.ClassName
	dst.Tenant = src.Tenant
	dst.Namespace = src.Namespace
	dst.ClusterName = src.ClusterName
	dst.Replicas = src.Replicas
	dst.MinReplicas = src.MinReplicas
	dst.DownloaderImage = src.DownloaderImage
	dst.MaxReplicas = src.MaxReplicas
	dst.Input = src.Input
	dst.Output = src.Output
	dst.LogTopic = src.LogTopic
	dst.FuncConfig = src.FuncConfig
	dst.Resources = src.Resources
	dst.SecretsMap = src.SecretsMap
	dst.VolumeMounts = src.VolumeMounts
	dst.Timeout = src.Timeout
	dst.AutoAck = src.AutoAck
	dst.MaxMessageRetry = src.MaxMessageRetry
	dst.ProcessingGuarantee = src.ProcessingGuarantee
	dst.RetainOrdering = src.RetainOrdering
	dst.RetainKeyOrdering = src.RetainKeyOrdering
	dst.DeadLetterTopic = src.DeadLetterTopic
	dst.ForwardSourceMessageProperty = src.ForwardSourceMessageProperty
	dst.MaxPendingAsyncRequests = src.MaxPendingAsyncRequests
	dst.RuntimeFlags = src.RuntimeFlags
	dst.SubscriptionName = src.SubscriptionName
	dst.CleanupSubscription = src.CleanupSubscription
	dst.SubscriptionPosition = src.SubscriptionPosition
	dst.Pod = src.Pod
	dst.WindowConfig = src.WindowConfig
	dst.Messaging = src.Messaging
	dst.Runtime = src.Runtime
	dst.Image = src.Image
	dst.ImagePullPolicy = src.ImagePullPolicy
	dst.StateConfig = src.StateConfig
	return nil
}

func (src *FunctionStatus) convertStatusTo(dst *v1alpha2.FunctionStatus) error {
	dst.Replicas = src.Replicas
	dst.Selector = src.Selector
	dst.Conditions = []metav1.Condition{}

	var condType apispec.ResourceConditionType
	var condReason apispec.ResourceConditionReason
	for component, condition := range src.Conditions {
		switch component {
		case apispec.StatefulSet:
			condType = apispec.StatefulSetReady
			condReason = apispec.StatefulSetIsReady
		case apispec.HPA:
			condType = apispec.HPAReady
			condReason = apispec.HPAIsReady
		case apispec.Service:
			condType = apispec.ServiceReady
			condReason = apispec.ServiceIsReady
		case apispec.VPA:
			condType = apispec.VPAReady
			condReason = apispec.VPAIsReady
		default:
			return fmt.Errorf("unkown component: %s", component)
		}

		desiredCondition := v1alpha2.CreateCondition(&src.ObservedGeneration,
			condType, metav1.ConditionTrue, condReason, "")
		if condition.Action != NoAction {
			desiredCondition.Status = metav1.ConditionFalse
			desiredCondition.Reason = string(apispec.PendingCreation)
			desiredCondition.Message = "resource is not ready yet..."
		}
		dst.Conditions = append(dst.Conditions, desiredCondition)
	}
	return nil
}

// ConvertFrom converts from the Hub version (v1alpha2) to this version.
func (dst *Function) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha2.Function)
	src.ObjectMeta = dst.ObjectMeta
	if err := dst.Spec.convertSpecFrom(&src.Spec); err != nil {
		return err
	}
	if err := dst.Status.convertStatusFrom(&src.Status); err != nil {
		return err
	}
	return nil
}

func (dst *FunctionSpec) convertSpecFrom(src *v1alpha2.FunctionSpec) error {
	dst.Name = src.Name
	dst.ClassName = src.ClassName
	dst.Tenant = src.Tenant
	dst.Namespace = src.Namespace
	dst.ClusterName = src.ClusterName
	dst.Replicas = src.Replicas
	dst.MinReplicas = src.MinReplicas
	dst.DownloaderImage = src.DownloaderImage
	dst.MaxReplicas = src.MaxReplicas
	dst.Input = src.Input
	dst.Output = src.Output
	dst.LogTopic = src.LogTopic
	dst.FuncConfig = src.FuncConfig
	dst.Resources = src.Resources
	dst.SecretsMap = src.SecretsMap
	dst.VolumeMounts = src.VolumeMounts
	dst.Timeout = src.Timeout
	dst.AutoAck = src.AutoAck
	dst.MaxMessageRetry = src.MaxMessageRetry
	dst.ProcessingGuarantee = src.ProcessingGuarantee
	dst.RetainOrdering = src.RetainOrdering
	dst.RetainKeyOrdering = src.RetainKeyOrdering
	dst.DeadLetterTopic = src.DeadLetterTopic
	dst.ForwardSourceMessageProperty = src.ForwardSourceMessageProperty
	dst.MaxPendingAsyncRequests = src.MaxPendingAsyncRequests
	dst.RuntimeFlags = src.RuntimeFlags
	dst.SubscriptionName = src.SubscriptionName
	dst.CleanupSubscription = src.CleanupSubscription
	dst.SubscriptionPosition = src.SubscriptionPosition
	dst.Pod = src.Pod
	dst.WindowConfig = src.WindowConfig
	dst.Messaging = src.Messaging
	dst.Runtime = src.Runtime
	dst.Image = src.Image
	dst.ImagePullPolicy = src.ImagePullPolicy
	dst.StateConfig = src.StateConfig
	return nil
}

func (dst *FunctionStatus) convertStatusFrom(src *v1alpha2.FunctionStatus) error {
	dst.Conditions = map[apispec.Component]ResourceCondition{}

	var condType apispec.ResourceConditionType
	for _, componentType := range []apispec.Component{
		apispec.StatefulSet, apispec.Service, apispec.HPA, apispec.VPA} {
		switch componentType {
		case apispec.StatefulSet:
			condType = apispec.StatefulSetReady
		case apispec.Service:
			condType = apispec.ServiceReady
		case apispec.HPA:
			condType = apispec.HPAReady
		case apispec.VPA:
			condType = apispec.VPAReady
		default:
			return fmt.Errorf("unkown component type: %s", componentType)
		}

		if condition := meta.FindStatusCondition(src.Conditions, string(condType)); condition != nil {
			var action ReconcileAction
			if condition.Status == metav1.ConditionTrue {
				action = NoAction
			} else {
				action = Wait
			}
			dst.Conditions[componentType] = ResourceCondition{
				Condition: ResourceConditionType(condType),
				Status:    condition.Status,
				Action:    action,
			}
		}
	}
	return nil
}
