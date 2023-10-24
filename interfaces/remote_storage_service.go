package interfaces

import "Parallels/pd-api-service/models"

type RemoteStorageService interface {
	Push(r *models.PushRemoteMachineRequest) error
	Pull(r *models.PullRemoteMachineRequest) error
}
