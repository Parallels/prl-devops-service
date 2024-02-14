package telemetry

import (
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/amplitude/analytics-go/amplitude"
)

var (
	globalTelemetryService *TelemetryService
	lock                   = &sync.Mutex{}
)

func New(cxt basecontext.ApiContext) *TelemetryService {
	svc := &TelemetryService{
		ctx: cxt,
	}

	// Getting the code inbuilt api key
	key := constants.AmplitudeApiKey
	if key == "" {
		// trying the api key from the environment variable
		cfg := config.Get()
		key = cfg.GetKey(constants.AmplitudeApiKeyEnvVar)
	}

	if key == "" {
		svc.EnableTelemetry = false
		return svc
	}

	config := amplitude.NewConfig(key)
	config.FlushQueueSize = 100
	config.FlushInterval = time.Second * 5
	svc.client = amplitude.NewClient(config)
	svc.client.Flush()

	return svc
}

func Get() *TelemetryService {
	if globalTelemetryService == nil {
		lock.Lock()
		ctx := basecontext.NewBaseContext()
		New(ctx)
		lock.Unlock()
	}

	return globalTelemetryService
}

func TrackEvent(item TelemetryItem) error {
	svc := Get()
	if svc == nil {
		return fmt.Errorf("[Telemetry] unable to track event %v, telemetry service is not available", item.Type)
	}

	return svc.TrackEvent(item)
}
