package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func remoteRoutesDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !state.ObjAsVpcPeering().Spec.Details.DeleteRemotePeering {
		logger.Info("Skipping route deletion")
		return nil, ctx
	}

	if state.remoteStatusIdEmpty {
		logger.Info("Skip deleting remote routes since VpcPeering.Status.RemoteId is empty")
		return nil, ctx
	}

	for _, rx := range state.remoteRoutesToDelete {

		err := state.remoteClient.DeleteRoute(ctx, rx.RouteTableId, rx.DestinationCidrBlock)

		lll := logger.WithValues(
			"remoteRouteTableId", ptr.Deref(rx.RouteTableId, "xxx"),
			"destinationCidrBlock", ptr.Deref(rx.DestinationCidrBlock, "xxx"),
		)

		if err != nil {
			if awsmeta.IsErrorRetryable(err) {
				return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
			}

			lll.Error(err, "Error deleting remote route")
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
		}

		lll.Info("Remote route deleted")

	}

	return nil, ctx
}
