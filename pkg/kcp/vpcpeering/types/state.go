package types

import (
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
)

type State interface {
	focal.State
	ObjAsVpcPeering() *v1beta1.VpcPeering
	LocalNetwork() *v1beta1.Network
	RemoteNetwork() *v1beta1.Network
	SetLocalNetwork(*v1beta1.Network)
	SetRemoteNetwork(*v1beta1.Network)
}

type state struct {
	focal.State
	localNetwork  *v1beta1.Network
	remoteNetwork *v1beta1.Network
}

func (s *state) ObjAsVpcPeering() *v1beta1.VpcPeering {
	return s.Obj().(*v1beta1.VpcPeering)
}

func (s *state) LocalNetwork() *v1beta1.Network {
	return s.localNetwork
}

func (s *state) RemoteNetwork() *v1beta1.Network {
	return s.remoteNetwork
}

func (s *state) SetLocalNetwork(network *v1beta1.Network) {
	s.localNetwork = network
}

func (s *state) SetRemoteNetwork(network *v1beta1.Network) {
	s.remoteNetwork = network
}

func NewState(focalState focal.State) State {
	return &state{State: focalState}
}
