package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeering == nil {
		logger.Info("Local peering not loaded on deleting VpcPeering")
	} else if state.localTerminating {
		logger.Info("Local peering can't be deleted at this stage",
			"peeringStatusCode", string(state.vpcPeering.Status.Code),
			"peeringStatusMessage", ptr.Deref(state.vpcPeering.Status.Message, ""))
	}

	if !state.localPeeringDelete {
		logger.Info("Local peering can't be deleted since it's not loaded or terminating")
		return nil, ctx
	}

	logger.Info("Deleting VpcPeering")

	err := state.client.DeleteVpcPeeringConnection(ctx, state.vpcPeering.VpcPeeringConnectionId)

	if awsmeta.IsErrorRetryable(err) {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Failed to delete local peering", composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	logger.Info("VpcPeering deleted")

	return nil, ctx
}
