package vpcpeering

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"testing"
)

func createState() *State {
	feature.InitializeFromStaticConfig(abstractions.NewOSEnvironment())

	clusterState := composed.NewStateCluster(nil, nil, nil, nil)
	baseStateFactory := composed.NewStateFactory(clusterState)

	obj := cloudcontrolv1beta1.NewVpcPeeringBuilder().
		WithRemoteRouteTableUpdateStrategy(cloudcontrolv1beta1.AwsRouteTableUpdateStrategyAuto).
		WithName("test-vpc-peering").Build()

	baseState := baseStateFactory.NewState(corev1.NamespacedName{Namespace: "kyma-system", Name: "test-vpc-peering"}, obj)

	focalStateFactory := focal.NewStateFactory()
	focalState := focalStateFactory.NewState(baseState)
	scope := cloudcontrolv1beta1.NewScopeBuilder().WithProvider(cloudcontrolv1beta1.ProviderAws).Build()
	scope.Spec.ShootName = "c-123456"
	scope.Spec.Scope.Aws = &cloudcontrolv1beta1.AwsScope{}
	focalState.SetScope(scope)

	return &State{
		State: types.NewState(focalState),
	}
}

func TestCreateNoNetworkTag(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := createState()

	s.remoteVpc = &ec2types.Vpc{
		Tags: []ec2types.Tag{},
	}

	err, ctx := calculate(ctx, s)

	assert.Nil(t, err)

	assert.False(t, s.hasShootTag)
	assert.False(t, s.localPeeringCreate)
	assert.False(t, s.localInitiating)
	assert.False(t, s.localTerminating)
	assert.False(t, s.remotePeeringAccept)
	assert.False(t, s.localActive)
}

func TestCreate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := createState()
	s.remoteVpc = &ec2types.Vpc{
		Tags: awsutil.Ec2Tags(s.Scope().Spec.ShootName, ""),
	}

	err, ctx := calculate(ctx, s)

	assert.Nil(t, err)

	assert.True(t, s.hasShootTag)
	assert.True(t, s.localPeeringCreate)
	assert.False(t, s.localInitiating)
	assert.False(t, s.localTerminating)
	assert.False(t, s.remotePeeringAccept)
	assert.False(t, s.localActive)
}

func TestInitiating(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := createState()

	s.remoteVpc = &ec2types.Vpc{
		Tags: awsutil.Ec2Tags(s.Scope().Spec.ShootName, ""),
	}

	s.vpcPeering = &ec2types.VpcPeeringConnection{
		Status: &ec2types.VpcPeeringConnectionStateReason{
			Code: ec2types.VpcPeeringConnectionStateReasonCodeInitiatingRequest,
		},
	}

	err, ctx := calculate(ctx, s)

	assert.Nil(t, err)

	assert.True(t, s.hasShootTag)
	assert.False(t, s.localPeeringCreate)
	assert.True(t, s.localInitiating)
	assert.False(t, s.localTerminating)
	assert.False(t, s.remotePeeringAccept)
	assert.False(t, s.localActive)
}

func TestAccept(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := createState()

	s.remoteVpc = &ec2types.Vpc{
		Tags: awsutil.Ec2Tags(s.Scope().Spec.ShootName, ""),
	}

	s.vpcPeering = &ec2types.VpcPeeringConnection{
		Status: &ec2types.VpcPeeringConnectionStateReason{
			Code: ec2types.VpcPeeringConnectionStateReasonCodePendingAcceptance,
		},
	}

	err, ctx := calculate(ctx, s)

	assert.Nil(t, err)

	assert.True(t, s.hasShootTag)
	assert.False(t, s.localPeeringCreate)
	assert.False(t, s.localInitiating)
	assert.False(t, s.localTerminating)
	assert.True(t, s.remotePeeringAccept)
	assert.False(t, s.localActive)
}

func TestActive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := createState()

	s.remoteVpc = &ec2types.Vpc{
		Tags: awsutil.Ec2Tags(s.Scope().Spec.ShootName, ""),
	}
	s.remoteVpc.CidrBlock = ptr.To("10.0.0.0/24")
	s.remoteVpc.CidrBlockAssociationSet = []ec2types.VpcCidrBlockAssociation{
		{
			CidrBlock: ptr.To("10.0.0.0/24"),
		},
		{
			CidrBlock: ptr.To("10.0.1.0/24"),
		},
	}

	s.vpc = &ec2types.Vpc{
		CidrBlock: ptr.To("10.0.2.0/24"),
		CidrBlockAssociationSet: []ec2types.VpcCidrBlockAssociation{
			{
				CidrBlock: ptr.To("10.0.2.0/24"),
			},
		},
	}

	s.vpcPeering = &ec2types.VpcPeeringConnection{
		Status: &ec2types.VpcPeeringConnectionStateReason{
			Code: ec2types.VpcPeeringConnectionStateReasonCodeActive,
		},
	}

	s.remoteVpcPeering = &ec2types.VpcPeeringConnection{}

	s.routeTables = []ec2types.RouteTable{
		{
			RouteTableId: ptr.To("rtb-1"),
			Routes: []ec2types.Route{
				{
					DestinationCidrBlock: ptr.To("0.0.0.0/0"),
				},
			},
		},
	}

	s.remoteRouteTables = []ec2types.RouteTable{
		{
			RouteTableId: ptr.To("rtb-1"),
			Routes: []ec2types.Route{
				{
					DestinationCidrBlock: ptr.To("0.0.0.0/0"),
				},
			},
		},
	}

	err, ctx := calculate(ctx, s)

	assert.Nil(t, err)

	assert.True(t, s.hasShootTag)
	assert.False(t, s.localPeeringCreate)
	assert.False(t, s.localInitiating)
	assert.False(t, s.localTerminating)
	assert.False(t, s.remotePeeringAccept)
	assert.True(t, s.localActive)

	assert.NotNil(t, s.remoteRoutesToCreate)
	assert.Equal(t, 1, len(s.remoteRoutesToCreate))
	assert.Equal(t, *s.remoteRoutesToCreate[0].RouteTableId, "rtb-1")
	assert.Equal(t, *s.remoteRoutesToCreate[0].DestinationCidrBlock, "10.0.2.0/24")

	assert.NotNil(t, s.localRoutesToCreate)
	assert.Equal(t, 2, len(s.localRoutesToCreate))

	assert.Nil(t, s.localRoutesToDelete)
}
