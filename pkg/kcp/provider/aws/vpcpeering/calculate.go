package vpcpeering

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	peeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"k8s.io/utils/ptr"
)

func calculate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.IsMarkedForDeletion(state.Obj()) {

		// delete local routes
		if len(state.ObjAsVpcPeering().Status.Id) == 0 {
			state.localStatusIdEmpty = true
		} else {
			for _, t := range state.routeTables {
				for _, r := range t.Routes {
					if ptr.Equal(r.VpcPeeringConnectionId, &state.ObjAsVpcPeering().Status.Id) {
						state.localRoutesToDelete = append(state.localRoutesToDelete, route{
							RouteTableId:         t.RouteTableId,
							DestinationCidrBlock: r.DestinationCidrBlock,
						})
					}
				}
			}
		}

		// delete local peering
		if state.vpcPeering != nil {
			if state.localTerminating = awsutil.IsTerminated(state.vpcPeering); !state.localTerminating {
				state.localPeeringDelete = true
			}
		}

		if state.ObjAsVpcPeering().Spec.Details.DeleteRemotePeering {
			if len(state.ObjAsVpcPeering().Status.RemoteId) == 0 {
				state.remoteStatusIdEmpty = true
			} else {

				for _, t := range state.remoteRouteTables {
					shouldUpdateRouteTable := awsutil.ShouldUpdateRouteTable(t.Tags,
						state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy,
						state.Scope().Spec.ShootName)

					if !shouldUpdateRouteTable {
						continue
					}

					for _, r := range t.Routes {
						if ptr.Equal(r.VpcPeeringConnectionId, &state.ObjAsVpcPeering().Status.RemoteId) {
							state.remoteRoutesToDelete = append(state.remoteRoutesToDelete, route{
								RouteTableId:         t.RouteTableId,
								DestinationCidrBlock: r.DestinationCidrBlock,
							})
						}
					}
				}
			}

			// delete remote peering
			if state.remoteVpcPeering != nil {
				if state.remoteTerminating = awsutil.IsTerminated(state.remoteVpcPeering); !state.localTerminating {
					state.remotePeeringDelete = true
				}
			}
		}

	} else {

		if state.remoteVpcPeering == nil {
			// If VpcNetwork is found but tags don't match, user can recover by adding tag to remote VPC network, so, we are
			// adding stop with requeue delay of one minute.
			state.hasShootTag = awsutil.HasEc2Tag(state.remoteVpc.Tags, peeringconfig.VpcPeeringConfig.NetworkTag)
			if !state.hasShootTag {
				state.hasShootTag = awsutil.HasEc2Tag(state.remoteVpc.Tags, state.Scope().Spec.ShootName)
			}
		} else {
			state.hasShootTag = true
		}

		if !state.hasShootTag {
			return nil, ctx
		}

		if state.vpcPeering == nil {
			state.localPeeringCreate = true
		} else {
			state.localInitiating = state.vpcPeering.Status.Code == types.VpcPeeringConnectionStateReasonCodeInitiatingRequest

			if state.localInitiating {
				return nil, ctx
			}

			if state.localTerminating = awsutil.IsTerminatedOrDeleting(state.vpcPeering); state.localTerminating {
				return nil, ctx
			}

			if state.remoteVpcPeering == nil {
				state.remotePeeringAccept = true
			}

			state.localActive = state.vpcPeering.Status.Code == types.VpcPeeringConnectionStateReasonCodeActive

			if !state.localActive {
				return nil, ctx
			}

			for _, t := range state.routeTables {

				for _, cidrBlockAssociation := range state.remoteVpc.CidrBlockAssociationSet {

					cidrBlock := cidrBlockAssociation.CidrBlock

					routeExists := pie.Any(t.Routes, func(r types.Route) bool {
						return ptr.Equal(r.VpcPeeringConnectionId, state.vpcPeering.VpcPeeringConnectionId) &&
							ptr.Equal(r.DestinationCidrBlock, cidrBlock)
					})

					if routeExists {
						continue
					}

					state.localRoutesToCreate = append(state.localRoutesToCreate, route{
						DestinationCidrBlock: cidrBlock,
						RouteTableId:         t.RouteTableId})
				}

				if !feature.VpcPeeringSync.Value(ctx) {
					continue
				}

				peeringRoutes := pie.Filter(t.Routes, func(r types.Route) bool {
					return ptr.Equal(r.VpcPeeringConnectionId, state.vpcPeering.VpcPeeringConnectionId)
				})

				for _, r := range peeringRoutes {
					exists := pie.Any(state.remoteVpc.CidrBlockAssociationSet, func(a types.VpcCidrBlockAssociation) bool {
						return ptr.Equal(a.CidrBlock, r.DestinationCidrBlock)
					})

					if !exists {
						state.localRoutesToDelete = append(state.localRoutesToDelete, route{
							DestinationCidrBlock: r.DestinationCidrBlock,
							RouteTableId:         t.RouteTableId})
					}
				}
			}

			if awsutil.IsRouteTableUpdateStrategyNone(state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy) {
				return nil, ctx
			}

			for _, t := range state.remoteRouteTables {
				shouldUpdateRouteTable := awsutil.ShouldUpdateRouteTable(t.Tags,
					state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy,
					state.Scope().Spec.ShootName)

				routeExists := pie.Any(t.Routes, func(r types.Route) bool {
					return ptr.Equal(r.VpcPeeringConnectionId, state.vpcPeering.VpcPeeringConnectionId) &&
						ptr.Equal(r.DestinationCidrBlock, state.vpc.CidrBlock)
				})

				if routeExists {
					continue
				}

				if !shouldUpdateRouteTable {
					continue
				}

				state.remoteRoutesToCreate = append(state.remoteRoutesToCreate, route{
					DestinationCidrBlock: state.vpc.CidrBlock,
					RouteTableId:         t.RouteTableId})

			}
		}
	}

	return nil, ctx

}
