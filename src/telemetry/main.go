package telemetry

import (
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"github.com/amplitude/analytics-go/amplitude"
	"github.com/amplitude/analytics-go/amplitude/types"
)

var (
	globalTelemetryService *TelemetryService
	lock                   = &sync.Mutex{}
)

func New(ctx basecontext.ApiContext) *TelemetryService {
	svc := &TelemetryService{
		ctx:             ctx,
		EnableTelemetry: true,
		CallBackChan:    make(chan types.ExecuteResult),
	}

	// Getting the code inbuilt api key
	key := constants.AmplitudeApiKey
	if key == "" {
		// trying the api key from the environment variable
		cfg := config.Get()
		key = cfg.GetKey(constants.AmplitudeApiKeyEnvVar)
	}

	if key == "" {
		ctx.LogDebugf("[Telemetry] Telemetry disabled as no API key found")
		svc.EnableTelemetry = false
		return svc
	}

	config := amplitude.NewConfig(key)
	config.FlushQueueSize = 100
	config.FlushInterval = time.Second * 5
	// adding a callback to read what is the status
	config.ExecuteCallback = func(result types.ExecuteResult) {
		svc.Callback(result)
	}

	svc.client = amplitude.NewClient(config)

	globalTelemetryService = svc
	return svc
}

func Get() *TelemetryService {
	if globalTelemetryService == nil {
		lock.Lock()
		ctx := basecontext.NewBaseContext()
		globalTelemetryService = New(ctx)
		lock.Unlock()
		return globalTelemetryService
	}

	return globalTelemetryService
}

func TrackEvent(item TelemetryItem) {
	svc := Get()
	if !svc.EnableTelemetry {
		return
	}

	svc.TrackEvent(item)
}

func SendStartEvent(cmd string) {
	svc := Get()
	if !svc.EnableTelemetry {
		return
	}

	ctx := basecontext.NewRootBaseContext()
	sys := system.Get()
	os := sys.GetOperatingSystem()
	properties := map[string]interface{}{
		"version": system.VersionSvc.String(),
		"os":      os,
		"mode":    cmd,
	}

	TrackEvent(NewTelemetryItem(ctx, StartEvent, properties, nil))
}
