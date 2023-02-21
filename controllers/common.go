package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	computeapi "github.com/streamnative/function-mesh/api/compute/v1alpha2"
)

const (
	ResourceCreated = "Created"
	ResourceReady   = "Ready"
)

type ReconciliationHelper interface {
	FinalizeStatus()
	GetInstance() client.Object
	GetInstanceType() string
	GetState() *State
}

type State struct {
	FunctionState string
	SourceState   string
	SinkState     string

	OrphanedFunctionInstances []computeapi.Function
	OrphanedSinkInstances     []computeapi.Sink
	OrphanedSourceInstances   []computeapi.Source

	StatefulSetState string
	ServiceState     string
	HPAState         string
	VPAState         string

	notReadyMessage string
}

func (s *State) SetNotReadyMessage(msg string) {
	if s.notReadyMessage == "" {
		if len(msg) > 128 {
			msg = msg[0:128]
		}
		s.notReadyMessage = msg
	}
}

func (s *State) SetOrphanedFunction(function *computeapi.Function) {
	if s.OrphanedFunctionInstances == nil {
		s.OrphanedFunctionInstances = []computeapi.Function{}
	}
	s.OrphanedFunctionInstances = append(s.OrphanedFunctionInstances, *function)
}

func (s *State) SetOrphanedSource(source *computeapi.Source) {
	if s.OrphanedSourceInstances == nil {
		s.OrphanedSourceInstances = []computeapi.Source{}
	}
	s.OrphanedSourceInstances = append(s.OrphanedSourceInstances, *source)
}

func (s *State) SetOrphanedSink(sink *computeapi.Sink) {
	if s.OrphanedSinkInstances == nil {
		s.OrphanedSinkInstances = []computeapi.Sink{}
	}
	s.OrphanedSinkInstances = append(s.OrphanedSinkInstances, *sink)
}

type FunctionReconciliationHelper struct {
	Instance *computeapi.Function `json:"instance"`
	State    `json:",inline"`
}

func (helper *FunctionReconciliationHelper) FinalizeStatus() {
	status := &helper.Instance.Status
	status.ObservedGeneration = helper.Instance.Generation

	created := helper.State.StatefulSetState != "unready" && helper.State.ServiceState != "unready"
	if helper.Instance.Spec.MaxReplicas != nil {
		created = created && helper.State.HPAState != "unready"
	}
	if helper.Instance.Spec.Pod.VPA != nil {
		created = created && helper.State.VPAState != "unready"
	}
	createdCondition := &metav1.Condition{
		Type:               ResourceCreated,
		Status:             metav1.ConditionTrue,
		Reason:             "AllResourcesCreated",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: helper.Instance.Generation,
	}
	if !created {
		createdCondition.Status = metav1.ConditionFalse
		createdCondition.Reason = "NotCreated"
		createdCondition.Message = helper.notReadyMessage
	}
	meta.SetStatusCondition(&helper.Instance.Status.Conditions, *createdCondition)

	ready := helper.StatefulSetState == "ready" && helper.ServiceState == "ready"
	if helper.Instance.Spec.MaxReplicas != nil {
		ready = ready && helper.HPAState == "ready"
	}
	if helper.Instance.Spec.Pod.VPA != nil {
		ready = ready && helper.State.VPAState == "ready"
	}
	readyCondition := &metav1.Condition{
		Type:               ResourceReady,
		Status:             metav1.ConditionTrue,
		Reason:             "AllResourcesReady",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: helper.Instance.Generation,
	}
	if !(created && ready) {
		readyCondition.Status = metav1.ConditionFalse
		readyCondition.Reason = "NotReady"
	}
	meta.SetStatusCondition(&helper.Instance.Status.Conditions, *readyCondition)
}

func (helper *FunctionReconciliationHelper) GetInstance() client.Object {
	return helper.Instance
}

func (helper *FunctionReconciliationHelper) GetInstanceType() string {
	return "function"
}

func (helper *FunctionReconciliationHelper) GetState() *State {
	return &helper.State
}

type SinkReconciliationHelper struct {
	Instance *computeapi.Sink `json:"instance"`
	State    `json:",inline"`
}

func (helper *SinkReconciliationHelper) FinalizeStatus() {
	status := &helper.Instance.Status
	status.ObservedGeneration = helper.Instance.Generation

	created := helper.State.StatefulSetState != "unready" && helper.State.ServiceState != "unready"
	if helper.Instance.Spec.MaxReplicas != nil {
		created = created && helper.State.HPAState != "unready"
	}
	if helper.Instance.Spec.Pod.VPA != nil {
		created = created && helper.State.VPAState != "unready"
	}
	createdCondition := &metav1.Condition{
		Type:               ResourceCreated,
		Status:             metav1.ConditionTrue,
		Reason:             "AllResourcesCreated",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: helper.Instance.Generation,
	}
	if !created {
		createdCondition.Status = metav1.ConditionFalse
		createdCondition.Reason = "NotCreated"
		createdCondition.Message = helper.notReadyMessage
	}
	meta.SetStatusCondition(&helper.Instance.Status.Conditions, *createdCondition)

	ready := helper.StatefulSetState == "ready" && helper.ServiceState == "ready"
	if helper.Instance.Spec.MaxReplicas != nil {
		ready = ready && helper.HPAState == "ready"
	}
	if helper.Instance.Spec.Pod.VPA != nil {
		ready = ready && helper.State.VPAState == "ready"
	}
	readyCondition := &metav1.Condition{
		Type:               ResourceReady,
		Status:             metav1.ConditionTrue,
		Reason:             "AllResourcesReady",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: helper.Instance.Generation,
	}
	if !(created && ready) {
		readyCondition.Status = metav1.ConditionFalse
		readyCondition.Reason = "NotReady"
	}
	meta.SetStatusCondition(&helper.Instance.Status.Conditions, *readyCondition)
}

func (helper *SinkReconciliationHelper) GetInstance() client.Object {
	return helper.Instance
}

func (helper *SinkReconciliationHelper) GetInstanceType() string {
	return "sink"
}

func (helper *SinkReconciliationHelper) GetState() *State {
	return &helper.State
}

type SourceReconciliationHelper struct {
	Instance *computeapi.Source `json:"instance"`
	State    `json:",inline"`
}

func (helper *SourceReconciliationHelper) FinalizeStatus() {
	status := &helper.Instance.Status
	status.ObservedGeneration = helper.Instance.Generation

	created := helper.State.StatefulSetState != "unready" && helper.State.ServiceState != "unready"
	if helper.Instance.Spec.MaxReplicas != nil {
		created = created && helper.State.HPAState != "unready"
	}
	if helper.Instance.Spec.Pod.VPA != nil {
		created = created && helper.State.VPAState != "unready"
	}
	createdCondition := &metav1.Condition{
		Type:               ResourceCreated,
		Status:             metav1.ConditionTrue,
		Reason:             "AllResourcesCreated",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: helper.Instance.Generation,
	}
	if !created {
		createdCondition.Status = metav1.ConditionFalse
		createdCondition.Reason = "NotCreated"
		createdCondition.Message = helper.notReadyMessage
	}
	meta.SetStatusCondition(&helper.Instance.Status.Conditions, *createdCondition)

	ready := helper.StatefulSetState == "ready" && helper.ServiceState == "ready"
	if helper.Instance.Spec.MaxReplicas != nil {
		ready = ready && helper.HPAState == "ready"
	}
	if helper.Instance.Spec.Pod.VPA != nil {
		ready = ready && helper.State.VPAState == "ready"
	}
	readyCondition := &metav1.Condition{
		Type:               ResourceReady,
		Status:             metav1.ConditionTrue,
		Reason:             "AllResourcesReady",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: helper.Instance.Generation,
	}
	if !(created && ready) {
		readyCondition.Status = metav1.ConditionFalse
		readyCondition.Reason = "NotReady"
	}
	meta.SetStatusCondition(&helper.Instance.Status.Conditions, *readyCondition)
}

func (helper *SourceReconciliationHelper) GetInstance() client.Object {
	return helper.Instance
}

func (helper *SourceReconciliationHelper) GetInstanceType() string {
	return "source"
}

func (helper *SourceReconciliationHelper) GetState() *State {
	return &helper.State
}

type MeshReconciliationHelper struct {
	Instance *computeapi.FunctionMesh `json:"instance"`
	State    `json:",inline"`
}

func (helper *MeshReconciliationHelper) FinalizeStatus() {
	status := &helper.Instance.Status
	status.ObservedGeneration = helper.Instance.Generation

	created := helper.State.FunctionState != "unready" &&
		helper.State.SourceState != "unready" &&
		helper.State.SinkState != "unready"
	createdCondition := &metav1.Condition{
		Type:               ResourceCreated,
		Status:             metav1.ConditionTrue,
		Reason:             "AllResourcesCreated",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: helper.Instance.Generation,
	}
	if !created {
		createdCondition.Status = metav1.ConditionFalse
		createdCondition.Reason = "NotCreated"
		createdCondition.Message = helper.notReadyMessage
	}
	meta.SetStatusCondition(&helper.Instance.Status.Conditions, *createdCondition)

	ready := helper.FunctionState == "ready" && helper.SourceState == "ready" && helper.SinkState == "ready"
	readyCondition := &metav1.Condition{
		Type:               ResourceReady,
		Status:             metav1.ConditionTrue,
		Reason:             "AllResourcesReady",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: helper.Instance.Generation,
	}
	if !(created && ready) {
		readyCondition.Status = metav1.ConditionFalse
		readyCondition.Reason = "NotReady"
	}
	meta.SetStatusCondition(&helper.Instance.Status.Conditions, *readyCondition)
}

func (helper *MeshReconciliationHelper) GetInstance() client.Object {
	return helper.Instance
}

func (helper *MeshReconciliationHelper) GetInstanceType() string {
	return "mesh"
}

func (helper *MeshReconciliationHelper) GetState() *State {
	return &helper.State
}

func MakeFunctionReconciliationHelper(function *computeapi.Function) *FunctionReconciliationHelper {
	return &FunctionReconciliationHelper{
		Instance: function,
		State:    State{},
	}
}

func MakeSinkReconciliationHelper(sink *computeapi.Sink) *SinkReconciliationHelper {
	return &SinkReconciliationHelper{
		Instance: sink,
		State:    State{},
	}
}

func MakeSourceReconciliationHelper(source *computeapi.Source) *SourceReconciliationHelper {
	return &SourceReconciliationHelper{
		Instance: source,
		State:    State{},
	}
}

func MakeMeshReconciliationHelper(mesh *computeapi.FunctionMesh) *MeshReconciliationHelper {
	return &MeshReconciliationHelper{
		Instance: mesh,
		State:    State{},
	}
}

// SaveStatus will trigger resource object update to save the current status
// conditions
func SaveStatus(ctx context.Context, logger logr.Logger, c client.StatusClient, helper ReconciliationHelper) {
	instance := helper.GetInstance()
	instanceType := helper.GetInstanceType()
	logger.Info(fmt.Sprintf("Updating status on %s", instanceType),
		"resource version", instance.GetResourceVersion())

	helper.FinalizeStatus()
	err := c.Status().Update(ctx, instance)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to update status on %s", instanceType), instanceType, instance)
	} else {
		logger.Info(fmt.Sprintf("Updating status on %s", instanceType),
			"resource version", instance.GetResourceVersion())
	}
}
