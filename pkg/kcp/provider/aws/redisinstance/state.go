package redisinstance

import (
	"context"
	"fmt"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	redisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance/client"

	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
)

type State struct {
	types.State
	awsClient redisinstanceclient.ElastiCacheClient

	subnetGroup        *elasticacheTypes.CacheSubnetGroup
	parameterGroup     *elasticacheTypes.CacheParameterGroup
	elastiCacheCluster *elasticacheTypes.CacheCluster
}

type StateFactory interface {
	NewState(ctx context.Context, redisInstace types.State) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[redisinstanceclient.ElastiCacheClient]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[redisinstanceclient.ElastiCacheClient]
}

func (f *stateFactory) NewState(ctx context.Context, redisInstace types.State) (*State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", redisInstace.Scope().Spec.Scope.Aws.AccountId, awsconfig.AwsConfig.AssumeRoleName)

	c, err := f.skrProvider(
		ctx,
		redisInstace.Scope().Spec.Region,
		awsconfig.AwsConfig.AccessKeyId,
		awsconfig.AwsConfig.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return nil, err
	}

	return newState(redisInstace, c), nil
}

func newState(redisInstace types.State, elastiCacheClient redisinstanceclient.ElastiCacheClient) *State {
	return &State{
		State:     redisInstace,
		awsClient: elastiCacheClient,
	}
}
