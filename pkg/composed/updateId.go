package composed

import (
	"context"
	"errors"
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func UpdateId(ctx context.Context, st State) (error, context.Context) {
	state := st
	logger := LoggerFromCtx(ctx)

	if MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	obj, ok := state.Obj().(ObjWithStatusId)
	if !ok {
		return errors.New("failed to convert to composed ObjWithId"), nil
	}

	if obj.Id() != "" {
		return nil, LoggerIntoCtx(ctx, logger.WithValues("id", obj.Id()))
	}

	id := uuid.NewString()

	if obj.GetLabels() == nil {
		obj.SetLabels(map[string]string{})
	}

	obj.GetLabels()[cloudresourcesv1beta1.LabelId] = id

	err := state.UpdateObj(ctx)

	if err != nil {
		return LogErrorAndReturn(err, "Error updating object with ID label", StopWithRequeue, ctx)
	}

	logger.Info("Object updated with ID label")

	obj.SetId(id)

	obj.SetDefaultState()

	ctx = LoggerIntoCtx(ctx, logger.WithValues("id", obj.Id()))

	_, ok = state.Obj().(ObjWithCloneForPatchStatus)

	if ok {
		err = state.PatchObjStatus(ctx)

		if err != nil {
			return LogErrorAndReturn(err, "Error patching object status with ID", StopWithRequeue, ctx)
		}

		return nil, ctx
	} else {
		err = state.UpdateObjStatus(ctx)

		if err != nil {
			return LogErrorAndReturn(err, "Error updating object status with ID", StopWithRequeue, ctx)
		}

		return StopWithRequeueDelay(util.Timing.T100ms()), ctx
	}
}
