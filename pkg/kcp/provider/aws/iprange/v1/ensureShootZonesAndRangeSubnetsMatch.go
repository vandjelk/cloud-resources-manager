package v1

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ensureShootZonesAndRangeSubnetsMatch(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	rangeSubnetCount := len(state.ObjAsIpRange().Status.Ranges)
	shootZonesCount := len(state.Scope().Spec.Scope.Aws.Network.Zones)
	if rangeSubnetCount != shootZonesCount {
		logger.
			WithValues(
				"rangeSubnetCount", rangeSubnetCount,
				"shootZonesCount", shootZonesCount,
			).
			Info("RangeSubnetCount different then shootZonesCount")

		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonShootAndVpcMismatch,
			Message: fmt.Sprintf("RangeSubnetCount %d different then shootZonesCount %d", rangeSubnetCount, shootZonesCount),
		})

		err := state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating IpRange status on shoot and vpc mismatch", composed.StopWithRequeue, ctx)
		}

		return composed.StopAndForget, nil
	}

	return nil, nil
}
