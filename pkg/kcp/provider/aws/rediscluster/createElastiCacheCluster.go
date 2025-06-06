package rediscluster

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createElastiCacheCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisInstance := state.ObjAsRedisCluster()

	if state.elastiCacheReplicationGroup != nil {
		return nil, nil
	}

	logger.Info("Creating AWS ElastiCache")

	var authTokenSecetString *string = nil

	if state.authTokenValue != nil {
		authTokenSecetString = state.authTokenValue.SecretString
	}

	_, err := state.awsClient.CreateElastiCacheReplicationGroup(ctx, []types.Tag{
		{
			Key:   ptr.To(common.TagCloudManagerName),
			Value: ptr.To(state.Name().String()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerRemoteName),
			Value: ptr.To(redisInstance.Spec.RemoteRef.String()),
		},
		{
			Key:   ptr.To(common.TagScope),
			Value: ptr.To(redisInstance.Spec.Scope.Name),
		},
		{
			Key:   ptr.To(common.TagShoot),
			Value: ptr.To(state.Scope().Spec.ShootName),
		},
	}, client.CreateElastiCacheClusterOptions{
		Name:                       GetAwsElastiCacheClusterName(state.Obj().GetName()),
		SubnetGroupName:            ptr.Deref(state.subnetGroup.CacheSubnetGroupName, ""),
		ParameterGroupName:         ptr.Deref(state.parameterGroup.CacheParameterGroupName, ""),
		CacheNodeType:              redisInstance.Spec.Instance.Aws.CacheNodeType,
		EngineVersion:              redisInstance.Spec.Instance.Aws.EngineVersion,
		AutoMinorVersionUpgrade:    redisInstance.Spec.Instance.Aws.AutoMinorVersionUpgrade,
		AuthTokenSecretString:      authTokenSecetString,
		PreferredMaintenanceWindow: redisInstance.Spec.Instance.Aws.PreferredMaintenanceWindow,
		SecurityGroupIds:           []string{state.securityGroupId},
		ReplicasPerNodeGroup:       redisInstance.Spec.Instance.Aws.ReplicasPerShard,
		ShardCount:                 redisInstance.Spec.Instance.Aws.ShardCount,
		ClusterMode:                true,
		AutomaticFailoverEnabled:   true,
		MultiAZEnabled:             nil,
	})

	if err != nil {
		logger.Error(err, "Error creating AWS ElastiCache")
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to create RedisCluster",
		})
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed aws elasticache creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
