package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"k8s.io/apimachinery/pkg/api/meta"
)

func createRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	for _, rx := range state.localRoutesToCreate {
		lll := logger.WithValues(
			"routeTableId", ptr.Deref(rx.RouteTableId, "xxx"),
			"destinationCidrBlock", ptr.Deref(rx.DestinationCidrBlock, "xxx"))

		err := state.client.CreateRoute(ctx, rx.RouteTableId, rx.DestinationCidrBlock, state.remoteVpcPeering.VpcPeeringConnectionId)

		if err == nil {
			lll.Info("Route created")
			continue
		}

		lll.Error(err, "Error creating route")

		if awsmeta.IsErrorRetryable(err) {
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}

		changed := false
		if meta.RemoveStatusCondition(obj.Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
			changed = true
		}

		if meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRoutes,
			Message: "Failed creating route for local route table",
		}) {
			changed = true
		}

		if obj.Status.State != string(cloudcontrolv1beta1.StateError) {
			obj.Status.State = string(cloudcontrolv1beta1.StateError)
			changed = true
		}

		if changed {
			// User cannot recover from internal error
			return composed.PatchStatus(obj).
				ErrorLogMessage("Error updating VpcPeering status when creating local routes").
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}

		// prevents alternating error messages
		return composed.StopAndForget, ctx
	}

	for _, rx := range state.localRoutesToDelete {
		err := state.client.DeleteRoute(ctx, rx.RouteTableId, rx.DestinationCidrBlock)

		if awsmeta.IsErrorRetryable(err) {
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}

		lll := logger.WithValues(
			"routeTableId", ptr.Deref(rx.RouteTableId, "xxx"),
			"destinationCidrBlock", ptr.Deref(rx.DestinationCidrBlock, "xxx"),
		)

		if err != nil {
			lll.Error(err, "Error deleting orphan route")
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
		}

		lll.Info("Orphan route deleted")

	}

	return nil, ctx
}
