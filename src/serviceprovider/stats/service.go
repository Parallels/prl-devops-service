package stats

import (
	"runtime"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
)
const emaAlpha = 0.2 // lower = smoother, higher = more reactive; 0.1–0.3 is a good range

// Broadcaster defines the capability required by StatsService to send messages.
type Broadcaster interface {
	BroadcastMessage(msg *models.EventMessage) error
	IsRunning() bool
}

// StatsMessage represents the payload for stats events
type StatsMessage struct {
	Memory        uint64  `json:"memory_bytes"`       // Allocated memory in bytes
  MemoryAlloc       uint64  `json:"memory_alloc_bytes"` // EMA-smoothed live alloc
	CPUUserTime   float64 `json:"cpu_user_seconds"`   // CPU user time in seconds
	CPUSystemTime float64 `json:"cpu_system_seconds"` // CPU system time in seconds
	CPUPercent  float64 `json:"cpu_percent"` // delta/elapsed — no more sawtooth
  Goroutines    int     `json:"goroutines"`         // Number of goroutines
  GoroutinesSmoothed int     `json:"goroutines_smoothed"` // EMA-smoothed
}

// StatsService handles stats collection and broadcasting
type StatsService struct {
	broadcaster Broadcaster
	isRunning   bool
	mu          sync.Mutex
	stopChan    chan struct{}
  // CPU delta tracking — added to fix Linux tick oscillation
  lastCPUUser  float64
  lastCPUSys   float64
  lastSampleAt time.Time
  // EMA smoothing for noisy metrics
  emaMemAlloc    float64
  emaGoroutines  float64
}

var (
	instance *StatsService
	once     sync.Once
)

// NewStatsService returns the singleton instance of StatsService
func NewStatsService(broadcaster Broadcaster) *StatsService {
	once.Do(func() {
		instance = &StatsService{
			broadcaster: broadcaster,
			stopChan:    make(chan struct{}),
		}
	})
	return instance
}

// Run starts the stats collection loop
func (s *StatsService) Run(ctx basecontext.ApiContext, interval time.Duration) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	ctx.LogInfof("[StatsService] Starting stats collection with interval %v", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			ctx.LogInfof("[StatsService] Stopping stats collection")
			s.mu.Lock()
			s.isRunning = false
			s.mu.Unlock()
			return
		case <-ticker.C:
			s.collectAndBroadcast(ctx)
		}
	}
}

// Stop stops the stats collection loop
func (s *StatsService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning {
		close(s.stopChan)
		s.isRunning = false
		// Re-create channel for next run
		s.stopChan = make(chan struct{})
	}
}

func (s *StatsService) collectAndBroadcast(ctx basecontext.ApiContext) {
	// If the event emitter has stopped, signal the run loop to exit
	if !s.broadcaster.IsRunning() {
		ctx.LogInfof("[StatsService] Event emitter is no longer running, stopping stats collection")
		s.Stop()
		return
	}

      now := time.Now()

    userTime, systemTime, err := getCPUTimes()
    if err != nil {
        ctx.LogWarnf("[StatsService] Failed to get CPU times: %v", err)
        return
    }

    // Compute CPU % as a rate: delta CPU time / elapsed wall time.
    // Raw cumulative rusage causes sawtooth on Linux due to scheduler tick resolution.
    var cpuPercent float64
    if !s.lastSampleAt.IsZero() {
        elapsed := now.Sub(s.lastSampleAt).Seconds()
        if elapsed > 0 {
            deltaUser := userTime - s.lastCPUUser
            deltaSys  := systemTime - s.lastCPUSys
            cpuPercent = ((deltaUser + deltaSys) / elapsed) * 100.0
        }
    }
    s.lastCPUUser  = userTime
    s.lastCPUSys   = systemTime
    s.lastSampleAt = now

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

      // EMA: seed on first sample, then smooth subsequent readings
    goroutines := float64(runtime.NumGoroutine())
    allocBytes := float64(memStats.Alloc)
    if s.emaMemAlloc == 0 {
        s.emaMemAlloc   = allocBytes
        s.emaGoroutines = goroutines
    } else {
        s.emaMemAlloc   = emaAlpha*allocBytes   + (1-emaAlpha)*s.emaMemAlloc
        s.emaGoroutines = emaAlpha*goroutines   + (1-emaAlpha)*s.emaGoroutines
    }

    stats := StatsMessage{
        Memory:     memStats.HeapSys,
        MemoryAlloc: uint64(s.emaMemAlloc),
        CPUPercent: cpuPercent,
        CPUUserTime: userTime,
        CPUSystemTime: systemTime,
        Goroutines: runtime.NumGoroutine(),
        GoroutinesSmoothed: int(s.emaGoroutines),
    }

	msg := models.NewEventMessage(constants.EventTypeStats, "System Stats", stats)

	if err := s.broadcaster.BroadcastMessage(msg); err != nil {
		ctx.LogWarnf("[StatsService] Failed to broadcast stats: %v", err)
	}
}
