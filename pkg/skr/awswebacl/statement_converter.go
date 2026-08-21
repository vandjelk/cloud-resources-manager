package awswebacl

import (
	"context"
	"fmt"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// convertRuleStatement converts a rule's StatementRef to AWS WAF Statement
func convertRuleStatement(ctx context.Context, kcpClient client.Client, rule *cloudresourcesv1beta1.AwsWebAclRule) (*wafv2types.Statement, error) {
	if rule.StatementRef == nil {
		return nil, fmt.Errorf("rule must have statementRef set")
	}

	// Fetch the AwsWebAclStatement CRD
	stmt := &cloudresourcesv1beta1.AwsWebAclStatement{}
	if err := kcpClient.Get(ctx, client.ObjectKey{Name: rule.StatementRef.Name}, stmt); err != nil {
		return nil, fmt.Errorf("failed to get AwsWebAclStatement %s: %w", rule.StatementRef.Name, err)
	}

	return convertStatement(stmt)
}

// convertStatement converts an AwsWebAclStatement CRD to AWS WAF format
func convertStatement(stmt *cloudresourcesv1beta1.AwsWebAclStatement) (*wafv2types.Statement, error) {
	result := &wafv2types.Statement{}

	if stmt.Spec.ManagedRuleGroup == nil {
		return nil, fmt.Errorf("statement must have ManagedRuleGroup set")
	}

	managedStmt, err := convertManagedRuleGroupStatement(stmt.Spec.ManagedRuleGroup)
	if err != nil {
		return nil, err
	}
	result.ManagedRuleGroupStatement = managedStmt
	return result, nil
}
