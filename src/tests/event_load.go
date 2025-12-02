package tests

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
	"github.com/Parallels/prl-devops-service/startup"
)

// RunEventLoadTest boots the service in API mode and broadcasts load-test events for WebSocket clients.
func RunEventLoadTest(ctx basecontext.ApiContext) error {
	startup.Init(ctx)

	cfg := config.Get()
	ctx.LogInfof("Starting Event Load Test API on port %s", cfg.ApiPort())
	startup.Start(ctx)

	listener := startup.InitApi()
	go listener.Start("Event Load Test API", "dev")

	apiPort := cfg.ApiPort()
	apiPrefix := cfg.ApiPrefix()
	if !strings.HasPrefix(apiPrefix, "/") {
		apiPrefix = "/" + apiPrefix
	}
	healthEndpoint := fmt.Sprintf("http://localhost:%s%s/health/probe", apiPort, strings.TrimSuffix(apiPrefix, "/"))

	readyCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for {
		req, err := http.NewRequestWithContext(readyCtx, http.MethodGet, healthEndpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create readiness request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ctx.LogInfof("Event Load Test API listener initialized")
				break
			}
		}

		if readyCtx.Err() != nil {
			return fmt.Errorf("timed out waiting for API initialization: %w", readyCtx.Err())
		}

		time.Sleep(500 * time.Millisecond)
	}

	func() {
		ticker := time.NewTicker(10 * time.Millisecond) // 100 events/sec
		defer ticker.Stop()

		seq := 0
		for range ticker.C {
			seq++

			vmPayload := models.VmAdded{
				VmID: fmt.Sprintf("vm-load-test-%d", seq),
				NewVm: models.ParallelsVM{
					ID:          fmt.Sprintf("seq-%d-ts-%d", seq, time.Now().UnixNano()),
					Name:        fmt.Sprintf("LoadTest-VM-%d", seq),
					Description: "Load test virtual machine with realistic payload size",
					State:       "running",
					OS:          "ubuntu-22.04",
					User:        "loadtest",
					HostId:      "load-test-host",
					Type:        "VM",
					Template:    "ubuntu-22.04-template",
				},
			}

			msg := models.NewEventMessage(constants.EventTypePDFM, "VM Added - Load Test", vmPayload)

			if err := eventemitter.Get().Broadcast(msg); err != nil {
				ctx.LogErrorf("Failed to broadcast load test message: %v", err)
			}
			if seq%1000 == 0 {
				ctx.LogInfof("Broadcasted %d VM events", seq)
			}
		}
	}()

	return nil
}
