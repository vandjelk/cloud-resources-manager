package v2

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
)

func rangeCheckBlockStatus(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	rangeCidr, _ := cidr.Parse(state.ObjAsIpRange().Spec.Cidr)
	for _, set := range state.vpc.CidrBlockAssociationSet {
		cdr, err := cidr.Parse(pointer.StringDeref(set.CidrBlock, ""))
		if err != nil {
			logger.Error(err, "Error parsing AWS CIDR: %w", err)
			continue
		}

		if util.CidrEquals(rangeCidr.CIDR(), cdr.CIDR()) &&
			// we must ignore disassociated sets
			!pie.Contains([]ec2Types.VpcCidrBlockStateCode{
				ec2Types.VpcCidrBlockStateCodeDisassociated,
				ec2Types.VpcCidrBlockStateCodeDisassociating,
			}, set.CidrBlockState.State) {
			state.associatedCidrBlock = &set
			break
		}
	}

	if state.associatedCidrBlock == nil {
		logger.Info("Matching AWS CIDR block not found")
		return nil, nil
	}

	logger.
		WithValues(
			"cidrBlockAssociationId", state.associatedCidrBlock.AssociationId,
			"cidrBlockAssociationState", state.associatedCidrBlock.CidrBlockState.State,
			"cidrBlockAssociationMessage", state.associatedCidrBlock.CidrBlockState.StatusMessage,
		).
		Info("Found matching AWS CIDR block")

	if pie.Contains([]ec2Types.VpcCidrBlockStateCode{
		ec2Types.VpcCidrBlockStateCodeAssociated,
		ec2Types.VpcCidrBlockStateCodeDisassociated,
		ec2Types.VpcCidrBlockStateCodeDisassociating,
	}, state.associatedCidrBlock.CidrBlockState.State) {
		return nil, nil
	}
	if state.associatedCidrBlock.CidrBlockState.State == ec2Types.VpcCidrBlockStateCodeAssociating {
		return composed.StopWithRequeueDelay(10 * time.Second), nil
	}

	return composed.PatchStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonCidrAssociationFailed,
			Message: fmt.Sprintf("CIDR block status state is %s", state.associatedCidrBlock.CidrBlockState.State),
		}).
		ErrorLogMessage("Failed patching KCP IpRange CidrAssociationFailed status").
		SuccessLogMsg("Forgetting KCP IpRange with unhandled CidrBlock state").
		Run(ctx, st)
}
