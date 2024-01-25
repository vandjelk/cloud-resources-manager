package iprange

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	awsiprange "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/iprange"
	azureiprange "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/azure/iprange"
	gcpiprange "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/iprange"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type IPRangeReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	awsStateFactory   awsiprange.StateFactory
	azureStateFactory azureiprange.StateFactory
	gcpStateFactory   gcpiprange.StateFactory
}

func NewIPRangeReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory awsiprange.StateFactory,
	azureStateFactory azureiprange.StateFactory,
	gcpStateFactory gcpiprange.StateFactory,
) *IPRangeReconciler {
	return &IPRangeReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpStateFactory:      gcpStateFactory,
	}
}

func (r *IPRangeReconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *IPRangeReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"ipRangeCommon",
				// common IpRange common actions here
				// ... none so far
				// and now branch to provider specific flow
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(focal.AwsProviderPredicate, awsiprange.New(r.awsStateFactory)),
					composed.NewCase(focal.AzureProviderPredicate, azureiprange.New(r.azureStateFactory)),
					composed.NewCase(focal.GcpProviderPredicate, gcpiprange.New(r.gcpStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *IPRangeReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.IpRange{}),
	)
}