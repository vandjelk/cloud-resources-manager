package vpcpeering

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createRemoteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rx := range state.remoteRoutesToCreate {
		lll := logger.WithValues(
			"remoteRouteTableId", ptr.Deref(rx.RouteTableId, "xxx"),
			"destinationCidrBlock", ptr.Deref(rx.DestinationCidrBlock, "xxx"))

		err := state.remoteClient.CreateRoute(ctx, rx.RouteTableId, rx.DestinationCidrBlock, state.vpcPeering.VpcPeeringConnectionId)

		if err == nil {
			lll.Info("Remote route created")
			continue
		}

		lll.Error(err, "Error creating remote route")

		if awsmeta.IsErrorRetryable(err) {
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}

		successError := composed.StopWithRequeueDelay(util.Timing.T60000ms())

		if awsmeta.IsRouteNotSupported(err) {
			successError = composed.StopAndForget
		}

		changed := false

		msg, _ := awsmeta.GetErrorMessage(err, "")
		if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRoutes,
			Message: fmt.Sprintf("Failed updating routes for remote route table %s. %s", ptr.Deref(rx.RouteTableId, ""), msg),
		}) {
			changed = true
		}

		if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.StateWarning) {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateWarning)
			changed = true
		}

		// Do not update the status if nothing is changed
		if !changed {
			return successError, ctx
		}

		// User can recover by modifying routes
		return composed.PatchStatus(state.ObjAsVpcPeering()).
			ErrorLogMessage("Error updating VpcPeering status when updating remote routes").
			SuccessError(successError).
			Run(ctx, state)

	}

	return nil, ctx
}
