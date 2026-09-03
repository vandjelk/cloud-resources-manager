/*
Copyright 2023.

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

package v1beta1

import (
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AwsWebAclSpec defines the desired state of AwsWebAcl
type AwsWebAclSpec struct {
	// Data contains the complete WebACL configuration in AWS API JSON format.
	// This should match the structure of AWS WAFv2 CreateWebACLInput/UpdateWebACLInput
	// excluding the Name and Scope fields which are set automatically.
	//
	// The JSON should include:
	// - DefaultAction: {"Allow": {}} or {"Block": {}}
	// - Rules: array of rule definitions
	// - VisibilityConfig: CloudWatch metrics configuration
	//
	// Example:
	// {
	//   "DefaultAction": {"Allow": {}},
	//   "Rules": [{
	//     "Name": "RateLimitRule",
	//     "Priority": 1,
	//     "Statement": {
	//       "RateBasedStatement": {
	//         "Limit": 2000,
	//         "AggregateKeyType": "IP"
	//       }
	//     },
	//     "Action": {"Block": {}},
	//     "VisibilityConfig": {
	//       "SampledRequestsEnabled": true,
	//       "CloudWatchMetricsEnabled": true,
	//       "MetricName": "RateLimitRule"
	//     }
	//   }],
	//   "VisibilityConfig": {
	//     "SampledRequestsEnabled": true,
	//     "CloudWatchMetricsEnabled": true,
	//     "MetricName": "MyWebACL"
	//   }
	// }
	//
	// See: https://docs.aws.amazon.com/waf/latest/APIReference/API_CreateWebACL.html
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Data string `json:"data"`
}

// AwsWebAclStatus defines the observed state of AwsWebAcl.
type AwsWebAclStatus struct {
	// ARN of the WebACL
	// +optional
	Arn string `json:"arn,omitempty"`

	// Capacity units consumed by the WebACL
	// +optional
	Capacity int64 `json:"capacity,omitempty"`

	// List of status conditions to indicate the status of a AwsWebAcl.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Capacity",type="integer",JSONPath=".status.capacity"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// AwsWebAcl is the Schema for the awswebacls API
type AwsWebAcl struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of AwsWebAcl
	// +required
	Spec AwsWebAclSpec `json:"spec,omitempty"`
	// status defines the observed state of AwsWebAcl
	// +optional
	Status AwsWebAclStatus `json:"status,omitempty"`
}

func (in *AwsWebAcl) ObservedGeneration() int64 {
	return in.Status.ObservedGeneration
}

func (in *AwsWebAcl) SetObservedGeneration(i int64) {
	in.Status.ObservedGeneration = i
}

func (in *AwsWebAcl) GetStatus() any {
	return &in.Status
}

func (in *AwsWebAcl) SetStatusProviderError(msg string) {
	in.Status.State = ReasonProviderError
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProviderError,
		Message:            msg,
	})
}

func (in *AwsWebAcl) SetStatusReady() {
	in.Status.State = StateReady
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Generation,
		Reason:             ReasonReady,
		Message:            ReasonReady,
	})
}

func (in *AwsWebAcl) SetStatusProcessing() {
	in.Status.State = StateProcessing
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProcessing,
		Message:            ReasonProcessing,
	})
}

func (in *AwsWebAcl) SetStatusDeleteWhileUsed(msg string) {
	in.Status.State = ReasonDeleteWhileUsed
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeDeleteWhileUsed,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Status.ObservedGeneration,
		Reason:             ReasonDeleteWhileUsed,
		Message:            msg,
	})
}

func (in *AwsWebAcl) RemoveStatusDeleteWhileUsed() {
	in.Status.State = StateDeleting
	meta.RemoveStatusCondition(&in.Status.Conditions, ConditionTypeDeleteWhileUsed)
}

func (in *AwsWebAcl) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *AwsWebAcl) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *AwsWebAcl) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureWAF
}

func (in *AwsWebAcl) SpecificToProviders() []string {
	return []string{"aws"}
}

func (in *AwsWebAcl) State() string {
	return in.Status.State
}

func (in *AwsWebAcl) SetState(v string) {
	in.Status.State = v
}

// +kubebuilder:object:root=true

// AwsWebAclList contains a list of AwsWebAcl
type AwsWebAclList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsWebAcl `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsWebAcl{}, &AwsWebAclList{})
}
