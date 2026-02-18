package stats

import (
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
)

// Broadcaster defines the capability required by StatsService to send messages.
type Broadcaster interface {
	BroadcastMessage(msg *models.EventMessage) error
}

// StatsMessage represents the payload for stats events
type StatsMessage struct {
	Memory        uint64  `json:"memory_bytes"`       // Allocated memory in bytes
	CpuUserTime   float64 `json:"cpu_user_seconds"`   // CPU user time in seconds
	CpuSystemTime float64 `json:"cpu_system_seconds"` // CPU system time in seconds
	Goroutines    int     `json:"goroutines"`         // Number of goroutines
}

// StatsService handles stats collection and broadcasting
type StatsService struct {
	broadcaster Broadcaster
	isRunning   bool
	mu          sync.Mutex
	stopChan    chan struct{}
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
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var rUsage syscall.Rusage
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &rUsage)
	if err != nil {
		ctx.LogWarnf("[StatsService] Failed to get rusage: %v", err)
		return
	}

	// Convert Timeval to seconds (float64)
	userTime := float64(rUsage.Utime.Sec) + float64(rUsage.Utime.Usec)/1e6
	systemTime := float64(rUsage.Stime.Sec) + float64(rUsage.Stime.Usec)/1e6

	stats := StatsMessage{
		Memory:        memStats.Alloc,
		CpuUserTime:   userTime,
		CpuSystemTime: systemTime,
		Goroutines:    runtime.NumGoroutine(),
	}

	msg := models.NewEventMessage(constants.EventTypeStats, "System Stats", stats)

	if err := s.broadcaster.BroadcastMessage(msg); err != nil {
		ctx.LogWarnf("[StatsService] Failed to broadcast stats: %v", err)
	}
}
