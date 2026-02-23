package logs

import (
	"strings"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"

	log "github.com/cjlapao/common-go-logger"
)

// Broadcaster defines the capability required by LogService to send messages.
type Broadcaster interface {
	BroadcastMessage(msg *models.EventMessage) error
}

// LogMessage represents the payload for log events
type LogMessage struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// LogService handles log streaming to the event emitter
type LogService struct {
	broadcaster    Broadcaster
	isRunning      bool
	mu             sync.Mutex
	subscriptionId string
}

var (
	instance *LogService
	once     sync.Once
)

// NewLogService returns the singleton instance of LogService
func NewLogService(broadcaster Broadcaster) *LogService {
	once.Do(func() {
		instance = &LogService{
			broadcaster: broadcaster,
		}
	})
	return instance
}

// Run starts the log stream to the emitter
func (s *LogService) Run(ctx basecontext.ApiContext) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	ctx.LogInfof("[LogService] Starting log streaming to Event Emitter")

	subscriptionId := "emitter-log-sub"
	onMessage := func(msg log.LogMessage) {
		// Filter out noisy or recursive messages
		if strings.Contains(msg.Message, "[Hub]") || strings.Contains(msg.Message, "[StatsService]") {
			return
		}

		logMsg := LogMessage{
			Level:   msg.Level,
			Message: msg.Message,
			Time:    msg.Timestamp,
		}
		eventMsg := models.NewEventMessage(constants.EventTypeSystemLogs, "System Log", logMsg)

		if err := s.broadcaster.BroadcastMessage(eventMsg); err != nil {
			// Intentionally empty as we do not want to trigger another log loop
		}
	}
	s.subscriptionId = ctx.Logger().OnMessage(subscriptionId, onMessage)
}

// Stop stops the log streaming
func (s *LogService) Stop(ctx basecontext.ApiContext) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning {
		if s.subscriptionId != "" {
			ctx.Logger().RemoveMessageHandler(s.subscriptionId)
			s.subscriptionId = ""
		}
		s.isRunning = false
	}
}
