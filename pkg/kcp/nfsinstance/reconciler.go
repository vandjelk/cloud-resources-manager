package nfsinstance

import (
	"context"
	nfsinstance3 "github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/aws/nfsinstance"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/azure/nfsinstance"
	nfsinstance2 "github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/gcp/nfsinstance"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type NfsInstanceReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	awsStateFactory   nfsinstance3.StateFactory
	azureStateFactory nfsinstance.StateFactory
	gcpStateFactory   nfsinstance2.StateFactory
}

func NewNfsInstanceReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory nfsinstance3.StateFactory,
	azureStateFactory nfsinstance.StateFactory,
	gcpStateFactory nfsinstance2.StateFactory,
) *NfsInstanceReconciler {
	return &NfsInstanceReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpStateFactory:      gcpStateFactory,
	}
}

func (r *NfsInstanceReconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *NfsInstanceReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"nfsInstanceCommon",
				// common NfsInstance common actions here
				loadIpRange,
				// and now branch to provider specific flow
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(focal.AwsProviderPredicate, nfsinstance3.New(r.awsStateFactory)),
					composed.NewCase(focal.AzureProviderPredicate, nfsinstance.New(r.azureStateFactory)),
					composed.NewCase(focal.GcpProviderPredicate, nfsinstance2.New(r.gcpStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *NfsInstanceReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.NfsInstance{}),
	)
}
