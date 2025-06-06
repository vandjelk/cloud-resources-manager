package vpcpeering

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	client       awsvpcpeeringclient.Client
	remoteClient awsvpcpeeringclient.Client
	provider     awsclient.SkrClientProvider[awsvpcpeeringclient.Client]

	awsAccessKeyid     string
	awsSecretAccessKey string
	roleName           string

	vpc              *ec2types.Vpc
	vpcPeering       *ec2types.VpcPeeringConnection
	remoteVpc        *ec2types.Vpc
	remoteVpcPeering *ec2types.VpcPeeringConnection

	routeTables       []ec2types.RouteTable
	remoteRouteTables []ec2types.RouteTable

	localNetwork  *cloudcontrolv1beta1.Network
	remoteNetwork *cloudcontrolv1beta1.Network
}

type StateFactory interface {
	NewState(ctx context.Context, state vpcpeeringtypes.State, logger logr.Logger) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[awsvpcpeeringclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[awsvpcpeeringclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {
	roleName := awsconfig.AwsConfig.Peering.AssumeRoleName
	awsAccessKeyId := awsconfig.AwsConfig.Peering.AccessKeyId
	awsSecretAccessKey := awsconfig.AwsConfig.Peering.SecretAccessKey

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", vpcPeeringState.Scope().Spec.Scope.Aws.AccountId, roleName)

	logger.WithValues(
		"awsRegion", vpcPeeringState.Scope().Spec.Region,
		"awsRole", roleArn,
	).Info("Assuming AWS role")

	c, err := f.skrProvider(
		ctx,
		vpcPeeringState.Scope().Spec.Scope.Aws.AccountId,
		vpcPeeringState.Scope().Spec.Region,
		awsAccessKeyId,
		awsSecretAccessKey,
		roleArn,
	)

	if err != nil {
		return nil, err
	}

	return newState(vpcPeeringState, c, f.skrProvider, awsAccessKeyId, awsSecretAccessKey, roleName), nil
}

func newState(vpcPeeringState vpcpeeringtypes.State,
	client awsvpcpeeringclient.Client,
	provider awsclient.SkrClientProvider[awsvpcpeeringclient.Client],
	key string,
	secret string,
	roleName string) *State {
	return &State{
		State:              vpcPeeringState,
		client:             client,
		provider:           provider,
		awsAccessKeyid:     key,
		awsSecretAccessKey: secret,
		roleName:           roleName,
	}
}
