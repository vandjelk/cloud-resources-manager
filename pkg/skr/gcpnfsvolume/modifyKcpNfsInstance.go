package gcpnfsvolume

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyKcpNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR GcpNfsVolume is marked for deletion, do not create mirror in KCP
		return nil, nil
	}

	if state.KcpNfsInstance == nil {
		return createKcpNfsInstance(ctx, state, logger.WithValues("operation", "createKcpNfsInstance"))
	} else if state.IsChanged() {
		return updateKcpNfsInstance(ctx, state, logger.WithValues("operation", "updateKcpNfsInstance"))
	}

	return nil, nil

}

func createKcpNfsInstance(ctx context.Context, state *State, logger logr.Logger) (error, context.Context) {
	state.KcpNfsInstance = &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.NewString(),
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      state.Name().Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: state.Name().Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.NfsInstanceSpec{
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: state.ObjAsGcpNfsVolume().Namespace,
				Name:      state.ObjAsGcpNfsVolume().Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.KcpIpRange.Name,
			},
			Instance: cloudcontrolv1beta1.NfsInstanceInfo{
				Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
					Location:      state.ObjAsGcpNfsVolume().Spec.Location,
					Tier:          cloudcontrolv1beta1.GcpFileTier(state.ObjAsGcpNfsVolume().Spec.Tier),
					FileShareName: state.ObjAsGcpNfsVolume().Spec.FileShareName,
					CapacityGb:    state.ObjAsGcpNfsVolume().Spec.CapacityGb,
					ConnectMode:   cloudcontrolv1beta1.PRIVATE_SERVICE_ACCESS,
				},
			},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpNfsInstance)
	if err != nil {
		logger.Error(err, "Error creating KCP NfsInstance")
		return composed.StopWithRequeue, nil
	}
	logger.
		WithValues("kcpNfsInstanceName", state.KcpNfsInstance.Name).
		Info("KCP NFS instance created")

	// Update the object with the ID of the KCP NfsInstance
	state.ObjAsGcpNfsVolume().Status.Id = state.KcpNfsInstance.Name
	state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeProcessing
	return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}

func updateKcpNfsInstance(ctx context.Context, state *State, logger logr.Logger) (error, context.Context) {
	modified := state.KcpNfsInstance.DeepCopy()
	// As of now, only CapacityGb is mutable
	modified.Spec.Instance.Gcp.CapacityGb = state.ObjAsGcpNfsVolume().Spec.CapacityGb
	err := state.KcpCluster.K8sClient().Update(ctx, modified)

	if err != nil {
		logger.Error(err, "Error updating KCP NfsInstance")
		return composed.StopWithRequeue, nil
	}
	logger.
		WithValues("kcpNfsInstanceName", state.KcpNfsInstance.Name).
		Info("KCP NFS instance got updated")

	// Update the object with the ID of the KCP NfsInstance
	state.ObjAsGcpNfsVolume().Status.Id = state.KcpNfsInstance.Name
	state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeProcessing
	return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
		RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
