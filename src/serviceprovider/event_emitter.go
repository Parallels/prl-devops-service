package serviceprovider

import (
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
)

func GetEventEmitter() *eventemitter.EventEmitter {
	return eventemitter.Get()
}
