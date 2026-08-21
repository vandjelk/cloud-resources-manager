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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AwsWebAclStatementSpec defines a reusable statement that can be referenced by WebACL rules
type AwsWebAclStatementSpec struct {
	// ManagedRuleGroup - Use AWS-managed rule sets
	// +kubebuilder:validation:Required
	ManagedRuleGroup *AwsWebAclManagedRuleGroupStatement `json:"managedRuleGroup"`
}

// AwsWebAclManagedRuleGroupStatement defines AWS-managed rule group configuration
// +kubebuilder:validation:XValidation:rule="self.vendorName == 'AWS' && (self.name == 'AWSManagedRulesCommonRuleSet' || self.name == 'AWSManagedRulesKnownBadInputsRuleSet' || self.name == 'AWSManagedRulesSQLiRuleSet' || self.name == 'AWSManagedRulesLinuxRuleSet' || self.name == 'AWSManagedRulesUnixRuleSet')", message="Only free AWS managed rules are supported: AWSManagedRulesCommonRuleSet, AWSManagedRulesKnownBadInputsRuleSet, AWSManagedRulesSQLiRuleSet, AWSManagedRulesLinuxRuleSet, AWSManagedRulesUnixRuleSet. Paid AWS rules and marketplace vendor rules require subscriptions in the service provider's AWS account."
type AwsWebAclManagedRuleGroupStatement struct {
	// VendorName (typically "AWS" for AWS managed rules)
	// +kubebuilder:validation:Required
	VendorName string `json:"vendorName"`

	// Name of the managed rule group
	// Common AWS managed rules:
	// - AWSManagedRulesCommonRuleSet
	// - AWSManagedRulesKnownBadInputsRuleSet
	// - AWSManagedRulesSQLiRuleSet
	// - AWSManagedRulesLinuxRuleSet
	// - AWSManagedRulesUnixRuleSet
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Version of the rule group (optional, uses latest if not specified)
	// +optional
	Version string `json:"version,omitempty"`

	// ExcludedRules to disable specific rules within the managed group
	// +optional
	ExcludedRules []AwsWebAclExcludedRule `json:"excludedRules,omitempty"`

	// RuleActionOverrides to override actions for specific rules
	// +optional
	RuleActionOverrides []AwsWebAclRuleActionOverride `json:"ruleActionOverrides,omitempty"`

	// ScopeDownStatementRef - Reference to external AwsWebAclStatement CRD by name
	// +optional
	ScopeDownStatementRef string `json:"scopeDownStatementRef,omitempty"`

	// ManagedRuleGroupConfigs - Vendor-specific configurations
	// +optional
	ManagedRuleGroupConfigs []AwsWebAclManagedRuleGroupConfig `json:"managedRuleGroupConfigs,omitempty"`
}

// AwsWebAclManagedRuleGroupConfig - Vendor-specific managed rule group configuration
type AwsWebAclManagedRuleGroupConfig struct {
	// LoginPath - Path for login page (for ATP/ACFP rule groups - not supported in MVP)
	// +optional
	LoginPath string `json:"loginPath,omitempty"`

	// PayloadType - Type of payload (for ATP/ACFP rule groups - not supported in MVP)
	// +optional
	// +kubebuilder:validation:Enum=JSON;FORM_ENCODED
	PayloadType string `json:"payloadType,omitempty"`
}

type AwsWebAclRuleActionOverride struct {
	// Name of the rule to override
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// ActionToUse to replace the original action
	// +kubebuilder:validation:Required
	ActionToUse *AwsWebAclRuleAction `json:"actionToUse"`
}

type AwsWebAclExcludedRule struct {
	// Name of the rule to exclude from the managed rule group
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// AwsWebAclStatementReference references an external AwsWebAclStatement CRD by name
type AwsWebAclStatementReference struct {
	// Name of the AwsWebAclStatement CRD to reference
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	Name string `json:"name"`
}

// AwsWebAclStatementStatus defines the observed state
type AwsWebAclStatementStatus struct {
	// Conditions for the statement lifecycle
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// State - High-level state of the statement
	// +optional
	State string `json:"state,omitempty"`

	// ObservedGeneration tracks the generation of the spec
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// AwsWebAclStatement is a reusable statement that can be referenced by WebACL rules
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type AwsWebAclStatement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsWebAclStatementSpec   `json:"spec,omitempty"`
	Status AwsWebAclStatementStatus `json:"status,omitempty"`
}

// AwsWebAclStatementList contains a list of AwsWebAclStatement
// +kubebuilder:object:root=true
type AwsWebAclStatementList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsWebAclStatement `json:"items"`
}

func (in *AwsWebAclStatement) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureWAF
}

func (in *AwsWebAclStatement) SpecificToProviders() []string {
	return []string{"aws"}
}

func init() {
	SchemeBuilder.Register(&AwsWebAclStatement{}, &AwsWebAclStatementList{})
}
