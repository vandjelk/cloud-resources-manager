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

package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AwsWebAcl Controller", func() {
	It("Scenario: SKR AwsWebAcl with StatementRef is created then deleted", func() {

		awsAccountLocal := infra.AwsMock().NewAccount()
		defer awsAccountLocal.Delete()

		scope := &cloudcontrolv1beta1.Scope{}

		scopeName := "e08b6fe8-9628-4601-8351-7d443a078606"

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(scopeName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, awsAccountLocal.AccountId(), WithName(scopeName))).To(Succeed())
		})

		Expect(scope.Namespace).To(Equal(infra.KCP().Namespace()))
		Expect(scope.Name).To(Equal(scopeName))

		objName := "test-webacl"
		stmt1Name := "test-stmt-1"
		stmt2Name := "test-stmt-2"
		infra.ScopeProvider().Add(scopeprovider.MatchingObj(objName, scope))

		By("And Given scope is ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		// Create AwsWebAclStatement resources first
		stmt1 := &cloudresourcesv1beta1.AwsWebAclStatement{
			ObjectMeta: metav1.ObjectMeta{
				Name: stmt1Name,
			},
			Spec: cloudresourcesv1beta1.AwsWebAclStatementSpec{
				ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
					VendorName: "AWS",
					Name:       "AWSManagedRulesKnownBadInputsRuleSet",
				},
			},
		}

		stmt2 := &cloudresourcesv1beta1.AwsWebAclStatement{
			ObjectMeta: metav1.ObjectMeta{
				Name: stmt2Name,
			},
			Spec: cloudresourcesv1beta1.AwsWebAclStatementSpec{
				ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
					VendorName: "AWS",
					Name:       "AWSManagedRulesCommonRuleSet",
				},
			},
		}

		By("When AwsWebAclStatements are created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), stmt1).
				Should(Succeed())
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), stmt2).
				Should(Succeed())
		})

		awsWebAcl := &cloudresourcesv1beta1.AwsWebAcl{}

		By("When AwsWebAcl is created", func() {
			awsWebAcl.Spec = cloudresourcesv1beta1.AwsWebAclSpec{
				DefaultAction: cloudresourcesv1beta1.DefaultActionAllow(),
				Description:   "Web ACL for test application with AWS managed rule sets",
				VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
					CloudWatchMetricsEnabled: true,
					MetricName:               "TestAppWAFMetrics",
					SampledRequestsEnabled:   true,
				},
				Rules: []cloudresourcesv1beta1.AwsWebAclRule{
					{
						Name:           "AWS-AWSManagedRulesKnownBadInputsRuleSet",
						Priority:       0,
						OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
						StatementRef: &cloudresourcesv1beta1.AwsWebAclStatementReference{
							Name: stmt1Name,
						},
						VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
							CloudWatchMetricsEnabled: true,
							MetricName:               "AWS-AWSManagedRulesKnownBadInputsRuleSet",
							SampledRequestsEnabled:   true,
						},
					},
					{
						Name:           "AWS-AWSManagedRulesCommonRuleSet",
						Priority:       1,
						OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
						StatementRef: &cloudresourcesv1beta1.AwsWebAclStatementReference{
							Name: stmt2Name,
						},
						VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
							CloudWatchMetricsEnabled: true,
							MetricName:               "AWS-AWSManagedRulesCommonRuleSet",
							SampledRequestsEnabled:   true,
						},
					},
				},
			}

			Eventually(CreateAwsWebAcl).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsWebAcl,
					WithName(objName),
				).
				Should(Succeed())
		})

		By("Then AwsWebAcl gets to Ready condition eventually", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsWebAcl,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then AwsWebAcl status has ARN populated", func() {
			Expect(awsWebAcl.Status.Arn).NotTo(BeEmpty())
			Expect(awsWebAcl.Status.Capacity).To(BeNumerically(">", 0))
		})

		By("When AwsWebAcl is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})

		By("Then AwsWebAcl does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})

		By("Cleanup AwsWebAclStatements", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), stmt1).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), stmt2).
				Should(Succeed())
		})
	})
})
