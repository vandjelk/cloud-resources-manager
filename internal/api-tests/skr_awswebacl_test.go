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

package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: SKR AwsWebAcl", func() {

	It("Scenario: AwsWebAcl with StatementRef can be created", func() {
		const (
			stmtName   = "test-stmt"
			webaclName = "test-webacl"
		)

		stmt := &cloudresourcesv1beta1.AwsWebAclStatement{
			ObjectMeta: metav1.ObjectMeta{
				Name: stmtName,
			},
			Spec: cloudresourcesv1beta1.AwsWebAclStatementSpec{
				ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
					VendorName: "AWS",
					Name:       "AWSManagedRulesCommonRuleSet",
				},
			},
		}

		webacl := &cloudresourcesv1beta1.AwsWebAcl{
			ObjectMeta: metav1.ObjectMeta{
				Name: webaclName,
			},
			Spec: cloudresourcesv1beta1.AwsWebAclSpec{
				DefaultAction: cloudresourcesv1beta1.DefaultActionAllow(),
				VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
					CloudWatchMetricsEnabled: true,
					MetricName:               "test-webacl",
					SampledRequestsEnabled:   true,
				},
				Rules: []cloudresourcesv1beta1.AwsWebAclRule{
					{
						Name:           "test-rule",
						Priority:       0,
						OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
						StatementRef: &cloudresourcesv1beta1.AwsWebAclStatementReference{
							Name: stmtName,
						},
					},
				},
			},
		}

		By("When AwsWebAclStatement is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), stmt).
				Should(Succeed())
		})

		By("When AwsWebAcl is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl).
				Should(Succeed())
		})

		By("Then AwsWebAcl is created successfully", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl, NewObjActions()).
				Should(Succeed())
		})

		By("Cleanup", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), stmt).
				Should(Succeed())
		})
	})

	It("Scenario: AwsWebAcl with multiple rules can be created", func() {
		const (
			stmt1Name  = "test-stmt-1"
			stmt2Name  = "test-stmt-2"
			webaclName = "test-webacl-multi"
		)

		stmt1 := &cloudresourcesv1beta1.AwsWebAclStatement{
			ObjectMeta: metav1.ObjectMeta{
				Name: stmt1Name,
			},
			Spec: cloudresourcesv1beta1.AwsWebAclStatementSpec{
				ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
					VendorName: "AWS",
					Name:       "AWSManagedRulesCommonRuleSet",
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
					Name:       "AWSManagedRulesSQLiRuleSet",
				},
			},
		}

		webacl := &cloudresourcesv1beta1.AwsWebAcl{
			ObjectMeta: metav1.ObjectMeta{
				Name: webaclName,
			},
			Spec: cloudresourcesv1beta1.AwsWebAclSpec{
				DefaultAction: cloudresourcesv1beta1.DefaultActionAllow(),
				VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
					CloudWatchMetricsEnabled: true,
					MetricName:               "test-webacl-multi",
					SampledRequestsEnabled:   true,
				},
				Rules: []cloudresourcesv1beta1.AwsWebAclRule{
					{
						Name:           "common-rules",
						Priority:       0,
						OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
						StatementRef: &cloudresourcesv1beta1.AwsWebAclStatementReference{
							Name: stmt1Name,
						},
					},
					{
						Name:           "sql-injection-rules",
						Priority:       1,
						OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
						StatementRef: &cloudresourcesv1beta1.AwsWebAclStatementReference{
							Name: stmt2Name,
						},
					},
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

		By("When AwsWebAcl is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl).
				Should(Succeed())
		})

		By("Then AwsWebAcl is created successfully", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl, NewObjActions()).
				Should(Succeed())
		})

		By("Cleanup", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), stmt1).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), stmt2).
				Should(Succeed())
		})
	})
})
