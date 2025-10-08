package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitVpcPeeringActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.localActive {
		return nil, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
