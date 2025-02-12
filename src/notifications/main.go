package notifications

import (
	"fmt"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
)

var _globalNotificationService *NotificationService

type RateSample struct {
	Timestamp time.Time
	Size      int64
	Progress  float64
}

type ProgressTracker struct {
	CurrentProgress float64
	LastUpdateTime  time.Time
	LastLogTime     time.Time // Track when we last logged a message
	Prefix          string
	StartTime       time.Time
	TotalSize       int64
	CurrentSize     int64
	IsComplete      bool
	RateSamples     []RateSample // Store last minute of samples
}

type NotificationService struct {
	ctx                   basecontext.ApiContext
	forceClearLine        bool
	clearLineOnUpdate     bool
	clearProgressOnUpdate bool
	Channel               chan NotificationMessage
	stopChan              chan bool
	activeProgress        map[string]*ProgressTracker // Track active progress notifications
	progressCounters      map[string]float64
	previousMessage       NotificationMessage
	CurrentMessage        NotificationMessage
}

// ProgressRate contains rate information for a progress notification
type ProgressRate struct {
	// BytesPerSecond is the transfer rate in bytes/second
	BytesPerSecond float64
	// RecentBytesPerSecond is calculated over the last minute
	RecentBytesPerSecond float64
	// ProgressPerSecond is the progress percentage change per second
	ProgressPerSecond float64
}

// Add this constant at the top with other constants
const (
	minUpdateInterval         = 500 * time.Millisecond // Minimum time between progress updates
	significantProgressChange = 1.0                    // Minimum progress change to force an update
)

// Add these types for prediction
type SpeedTrend struct {
	Increasing bool
	Stable     bool
	Factor     float64 // How much speed is changing
}

func New(ctx basecontext.ApiContext) *NotificationService {
	_globalNotificationService := &NotificationService{
		ctx:               ctx,
		Channel:           make(chan NotificationMessage),
		clearLineOnUpdate: false,
		activeProgress:    make(map[string]*ProgressTracker),
		progressCounters:  make(map[string]float64),
	}

	_globalNotificationService.Start()
	return _globalNotificationService
}

func Get() *NotificationService {
	if _globalNotificationService == nil {
		ctx := basecontext.NewBaseContext()
		_globalNotificationService = New(ctx)
	}
	return _globalNotificationService
}

func (p *NotificationService) EnableSingleLineOutput() *NotificationService {
	p.clearLineOnUpdate = true
	p.forceClearLine = true
	return p
}

func (p *NotificationService) SetContext(ctx basecontext.ApiContext) *NotificationService {
	p.ctx = ctx
	return p
}

func (p *NotificationService) ResetCounters(correlationId string) {
	if correlationId != "" {
		delete(p.progressCounters, correlationId)
	}
}

func (p *NotificationService) Notify(msg *NotificationMessage) {
	p.Channel <- *msg
}

func (p *NotificationService) NotifyInfo(msg string) {
	p.Notify(NewNotificationMessage(msg, NotificationMessageLevelInfo))
}

func (p *NotificationService) NotifyInfof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Notify(NewNotificationMessage(msg, NotificationMessageLevelInfo))
}

func (p *NotificationService) NotifyWarning(msg string) {
	p.Notify(NewNotificationMessage(msg, NotificationMessageLevelWarning))
}

func (p *NotificationService) NotifyWarningf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Notify(NewNotificationMessage(msg, NotificationMessageLevelWarning))
}

func (p *NotificationService) NotifyError(msg string) {
	p.Notify(NewNotificationMessage(msg, NotificationMessageLevelError))
}

func (p *NotificationService) NotifyErrorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Notify(NewNotificationMessage(msg, NotificationMessageLevelError))
}

func (p *NotificationService) NotifyDebug(msg string) {
	p.Notify(NewNotificationMessage(msg, NotificationMessageLevelDebug))
}

func (p *NotificationService) NotifyDebugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Notify(NewNotificationMessage(msg, NotificationMessageLevelDebug))
}

func (p *NotificationService) updateProgressTracker(tracker *ProgressTracker, progress float64, currentSize int64) {
	now := time.Now()

	// Only add a new sample if the size has changed
	if currentSize != tracker.CurrentSize || len(tracker.RateSamples) == 0 {
		newSample := RateSample{
			Timestamp: now,
			Size:      currentSize,
			Progress:  progress,
		}

		// Keep only samples from the last minute
		validSamples := []RateSample{newSample}
		for _, sample := range tracker.RateSamples {
			if now.Sub(sample.Timestamp) <= time.Minute {
				validSamples = append(validSamples, sample)
			}
		}
		tracker.RateSamples = validSamples
	}

	tracker.CurrentProgress = progress
	tracker.LastUpdateTime = now
	tracker.CurrentSize = currentSize
}

func (p *NotificationService) NotifyProgress(correlationId string, prefix string, progress float64) {
	msg := NewProgressNotificationMessage(correlationId, prefix, progress)

	// Create or update progress tracker
	tracker, exists := p.activeProgress[correlationId]
	if !exists {
		tracker = &ProgressTracker{
			StartTime:   time.Now(),
			Prefix:      prefix,
			IsComplete:  false,
			RateSamples: make([]RateSample, 0, 60),
			TotalSize:   msg.totalSize, // Make sure we capture the total size
		}
		p.activeProgress[correlationId] = tracker
	}

	p.updateProgressTracker(tracker, progress, msg.currentSize)

	if progress >= 100 {
		msg.Close()
		tracker.IsComplete = true
	}

	p.Notify(msg)
}

func (p *NotificationService) FinishProgress(correlationId string, prefix string) {
	if tracker, exists := p.activeProgress[correlationId]; exists {
		tracker.CurrentProgress = 100
		tracker.LastUpdateTime = time.Now()
		tracker.IsComplete = true
	}

	msg := NewProgressNotificationMessage(correlationId, prefix, 100)
	msg.Close()
	p.Notify(msg)
}

func (p *NotificationService) Stop() {
	p.stopChan <- true
}

func (p *NotificationService) Start() {
	// Start the cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-p.stopChan:
				return
			case <-ticker.C:
				p.CleanupStaleProgress(15 * time.Minute)
			}
		}
	}()

	// Existing message processing goroutine
	go func() {
		defer close(p.Channel)
		for {
			select {
			case <-p.stopChan:
				return
			case p.CurrentMessage = <-p.Channel:
				shouldLog := false

				if p.CurrentMessage.IsProgress {
					tracker, exists := p.activeProgress[p.CurrentMessage.CorrelationId()]
					if !exists {
						// New progress notification
						if p.CurrentMessage.CurrentProgress < 100 {
							tracker = &ProgressTracker{
								StartTime:       time.Now(),
								Prefix:          p.CurrentMessage.Message,
								CurrentProgress: p.CurrentMessage.CurrentProgress,
								LastUpdateTime:  time.Now(),
								CurrentSize:     p.CurrentMessage.currentSize,
								TotalSize:       p.CurrentMessage.totalSize,
								RateSamples:     make([]RateSample, 0, 60),
							}
							p.activeProgress[p.CurrentMessage.CorrelationId()] = tracker
							shouldLog = true
						}
					} else {
						// Update existing tracker and check if we should log
						p.updateProgressTracker(tracker, p.CurrentMessage.CurrentProgress, p.CurrentMessage.currentSize)
						shouldLog = p.shouldLogProgress(tracker, p.CurrentMessage.CurrentProgress)
					}

					// Clean up completed progress
					if p.CurrentMessage.Closed() || p.CurrentMessage.CurrentProgress >= 100 {
						delete(p.activeProgress, p.CurrentMessage.CorrelationId())
					}
				} else {
					// Non-progress messages
					if p.CurrentMessage.Message != "" {
						shouldLog = true
					}
				}

				if p.CurrentMessage.Message != p.previousMessage.Message && !p.forceClearLine {
					p.previousMessage = p.CurrentMessage
					p.clearLineOnUpdate = false
				}

				// if logging is disabled in the context, then we should not log
				if !p.ctx.Verbose() {
					shouldLog = false
				}

				if shouldLog {
					requestId := p.ctx.GetRequestId()
					printMsg := ""
					if requestId != "" {
						printMsg = fmt.Sprintf("[%s] ", requestId)
					}

					if p.CurrentMessage.IsProgress {
						// Use the new formatting for progress messages
						if tracker, exists := p.activeProgress[p.CurrentMessage.CorrelationId()]; exists {
							baseMsg := p.CurrentMessage.Message
							if baseMsg == "" {
								baseMsg = tracker.Prefix
							}
							printMsg += baseMsg + " "
							printMsg += formatProgressMessage(&p.CurrentMessage, tracker, p)
						} else {
							// Fallback for completed/cleaned up progress
							printMsg += fmt.Sprintf("%s (%.1f%%)",
								p.CurrentMessage.Message,
								p.CurrentMessage.CurrentProgress)
						}
					} else {
						printMsg += p.CurrentMessage.Message
					}

					if p.clearLineOnUpdate {
						ClearLine()
						fmt.Printf("\r%s", printMsg)
					} else {
						switch p.CurrentMessage.Level {
						case NotificationMessageLevelError:
							p.ctx.LogErrorf("%s", printMsg)
						case NotificationMessageLevelWarning:
							p.ctx.LogWarnf("%s", printMsg)
						case NotificationMessageLevelDebug:
							p.ctx.LogDebugf("%s", printMsg)
						default:
							p.ctx.LogInfof("%s", printMsg)
						}
					}
				}
			}
		}
	}()
}

func (p *NotificationService) Restart() {
	p.Stop()
	p.Start()
}

func ClearLine() {
	fmt.Printf("\r\033[K")
}

// CleanupNotifications removes all progress tracking for a specific correlation ID
func (p *NotificationService) CleanupNotifications(correlationId string) {
	if correlationId == "" {
		return
	}

	// Remove from active progress tracking
	delete(p.activeProgress, correlationId)

	// Reset previous message if it was for this correlation ID
	if p.previousMessage.correlationId == correlationId {
		p.previousMessage = NotificationMessage{}
	}

	// Reset current message if it was for this correlation ID
	if p.CurrentMessage.correlationId == correlationId {
		p.CurrentMessage = NotificationMessage{}
	}

	p.ctx.LogDebugf("Cleaned up notifications for correlation ID: %s", correlationId)
}

// GetActiveProgressCount returns the number of active progress notifications
func (p *NotificationService) GetActiveProgressCount() int {
	return len(p.activeProgress)
}

// GetActiveProgressIDs returns a slice of correlation IDs for active progress notifications
func (p *NotificationService) GetActiveProgressIDs() []string {
	ids := make([]string, 0, len(p.activeProgress))
	for id := range p.activeProgress {
		ids = append(ids, id)
	}
	return ids
}

// IsProgressActive checks if a progress notification is active for the given correlation ID
func (p *NotificationService) IsProgressActive(correlationId string) bool {
	_, exists := p.activeProgress[correlationId]
	return exists
}

// GetProgressStatus returns the current progress status for a given correlation ID
// Returns progress percentage and whether the progress exists
func (p *NotificationService) GetProgressStatus(correlationId string) (float64, bool) {
	if tracker, exists := p.activeProgress[correlationId]; exists {
		return tracker.CurrentProgress, true
	}
	return 0, false
}

// CleanupStaleProgress removes progress notifications that haven't been updated
// for longer than the specified duration
func (p *NotificationService) CleanupStaleProgress(staleDuration time.Duration) {
	now := time.Now()
	for id, tracker := range p.activeProgress {
		if now.Sub(tracker.LastUpdateTime) > staleDuration {
			p.ctx.LogDebugf("Cleaning up stale progress for correlation ID: %s (last update: %v)",
				id, tracker.LastUpdateTime)
			p.CleanupNotifications(id)
		}
	}
}

// GetProgressDuration returns the duration since the progress started
func (p *NotificationService) GetProgressDuration(correlationId string) (time.Duration, bool) {
	if tracker, exists := p.activeProgress[correlationId]; exists {
		return time.Since(tracker.StartTime), true
	}
	return 0, false
}

// GetProgressRate calculates transfer and progress rates for a given correlation ID
func (p *NotificationService) GetProgressRate(correlationId string) (*ProgressRate, bool) {
	tracker, exists := p.activeProgress[correlationId]
	if !exists || tracker.TotalSize <= 0 {
		return nil, false
	}

	now := time.Now()
	totalDuration := now.Sub(tracker.StartTime).Seconds()
	if totalDuration <= 0 {
		return nil, false
	}

	rate := &ProgressRate{}

	// Calculate overall rate from the start
	rate.BytesPerSecond = float64(tracker.CurrentSize) / totalDuration
	rate.ProgressPerSecond = tracker.CurrentProgress / totalDuration

	// For recent rate, use a window of the last few samples but weight them
	if len(tracker.RateSamples) >= 2 {
		// Calculate rate using the total progress since start
		firstSample := tracker.RateSamples[0]
		lastSample := tracker.RateSamples[len(tracker.RateSamples)-1]

		totalBytes := lastSample.Size - firstSample.Size
		sampleDuration := lastSample.Timestamp.Sub(firstSample.Timestamp).Seconds()

		if sampleDuration > 0 {
			rate.RecentBytesPerSecond = float64(totalBytes) / sampleDuration
		} else {
			rate.RecentBytesPerSecond = rate.BytesPerSecond
		}
	} else {
		rate.RecentBytesPerSecond = rate.BytesPerSecond
	}

	return rate, true
}

// PredictTimeRemaining estimates the time remaining based on recent progress
func (p *NotificationService) PredictTimeRemaining(correlationId string) (time.Duration, bool) {
	tracker, exists := p.activeProgress[correlationId]
	if !exists || tracker.TotalSize <= 0 {
		return 0, false
	}

	rate, ok := p.GetProgressRate(correlationId)
	if !ok || rate.RecentBytesPerSecond <= 0 {
		return 0, false
	}

	trend := p.analyzeSpeedTrend(tracker.RateSamples)

	remainingBytes := float64(tracker.TotalSize - tracker.CurrentSize)
	currentSpeed := rate.RecentBytesPerSecond

	// Adjust prediction based on trend
	var remainingSeconds float64
	if trend.Stable {
		remainingSeconds = remainingBytes / currentSpeed
	} else {
		// Use trend factor to adjust prediction
		adjustedSpeed := currentSpeed * trend.Factor
		// Take average of current and trend-adjusted predictions
		remainingSeconds = (remainingBytes/currentSpeed + remainingBytes/adjustedSpeed) / 2
	}

	// Apply bounds to avoid wild predictions
	if remainingSeconds < 0 {
		remainingSeconds = 0
	} else if remainingSeconds > float64(24*time.Hour/time.Second) {
		remainingSeconds = float64(24 * time.Hour / time.Second)
	}

	return time.Duration(remainingSeconds * float64(time.Second)), true
}

// GetFormattedTimeRemaining returns a human-readable prediction of remaining time
func (p *NotificationService) GetFormattedTimeRemaining(correlationId string) string {
	duration, ok := p.PredictTimeRemaining(correlationId)
	if !ok {
		return "calculating..."
	}

	seconds := int(duration.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// FormatTransferRate converts bytes per second to a human-readable string
func FormatTransferRate(bytesPerSecond float64) string {
	units := []string{"B/s", "KB/s", "MB/s", "GB/s"}
	size := bytesPerSecond
	unitIndex := 0

	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}

// GetFormattedProgressRate returns a human-readable string of the progress rate
func (p *NotificationService) GetFormattedProgressRate(correlationId string) string {
	rate, exists := p.GetProgressRate(correlationId)
	if !exists {
		return "N/A"
	}

	var result string
	if rate.BytesPerSecond > 0 {
		currentRate := FormatTransferRate(rate.RecentBytesPerSecond)
		avgRate := FormatTransferRate(rate.BytesPerSecond)
		result = fmt.Sprintf("Current: %s, Average: %s", currentRate, avgRate)
	}

	if rate.ProgressPerSecond > 0 {
		if result != "" {
			result += ", "
		}
		result += fmt.Sprintf("%.2f%% per second", rate.ProgressPerSecond)
	}

	return result
}

// Update the shouldLog check in Start() method
func (p *NotificationService) shouldLogProgress(tracker *ProgressTracker, currentProgress float64) bool {
	now := time.Now()

	// Always show first and last updates
	if tracker.LastLogTime.IsZero() || currentProgress >= 100 {
		tracker.LastLogTime = now
		return true
	}

	timeSinceLastLog := now.Sub(tracker.LastLogTime)
	progressChange := currentProgress - tracker.CurrentProgress

	// Log if any of these conditions are met:
	// 1. Minimum time has passed (regardless of progress change)
	// 2. We have significant progress change
	// 3. We have any progress change and it's been at least 100ms
	shouldLog := timeSinceLastLog >= minUpdateInterval ||
		progressChange >= significantProgressChange ||
		(progressChange > 0 && timeSinceLastLog >= 100*time.Millisecond)

	if shouldLog {
		tracker.LastLogTime = now
	}

	return shouldLog
}

// Add to GetProgressRate method
func (p *NotificationService) analyzeSpeedTrend(samples []RateSample) SpeedTrend {
	if len(samples) < 3 {
		return SpeedTrend{Stable: true, Factor: 1.0}
	}

	// Calculate speed changes over time
	changes := make([]float64, len(samples)-1)
	for i := len(samples) - 1; i > 0; i-- {
		duration := samples[i].Timestamp.Sub(samples[i-1].Timestamp).Seconds()
		if duration > 0 {
			currentSpeed := float64(samples[i].Size-samples[i-1].Size) / duration
			prevSpeed := float64(samples[i-1].Size-samples[i-2].Size) /
				samples[i-1].Timestamp.Sub(samples[i-2].Timestamp).Seconds()
			if prevSpeed > 0 {
				changes[len(samples)-1-i] = currentSpeed / prevSpeed
			}
		}
	}

	// Analyze trend
	trend := SpeedTrend{Stable: true, Factor: 1.0}
	if len(changes) >= 2 {
		// Calculate weighted average of changes
		var weightedSum, weightTotal float64
		for i, change := range changes {
			weight := float64(i + 1) // More recent changes have higher weight
			weightedSum += change * weight
			weightTotal += weight
		}
		avgChange := weightedSum / weightTotal

		// Determine trend
		if avgChange > 1.1 {
			trend.Increasing = true
			trend.Stable = false
			trend.Factor = avgChange
		} else if avgChange < 0.9 {
			trend.Increasing = false
			trend.Stable = false
			trend.Factor = avgChange
		}
	}

	return trend
}

// Update the message formatting in Start() method
func formatProgressMessage(msg *NotificationMessage, tracker *ProgressTracker, p *NotificationService) string {
	var parts []string

	// Basic progress
	parts = append(parts, fmt.Sprintf("%.1f%%", msg.CurrentProgress))

	// Size information
	if msg.TotalSize() > 0 && msg.CurrentSize() > 0 {
		current := formatSize(float64(msg.CurrentSize()))
		total := formatSize(float64(msg.TotalSize()))
		parts = append(parts, fmt.Sprintf("[%s/%s]", current, total))

		// Calculate rate from the start
		rate, ok := p.GetProgressRate(msg.CorrelationId())
		if ok && rate.BytesPerSecond > 0 {
			parts = append(parts, FormatTransferRate(rate.BytesPerSecond))

			// Calculate ETA using the consistent rate from start
			eta := calculateETA(tracker.StartTime, tracker.CurrentSize, tracker.TotalSize)
			if eta != "calculating..." {
				parts = append(parts, fmt.Sprintf("ETA: %s", eta))
			}
		}
	}

	return strings.Join(parts, " ")
}

// Helper function to format sizes
func formatSize(bytes float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	unitIndex := 0

	for bytes >= 1024 && unitIndex < len(units)-1 {
		bytes /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.2f %s", bytes, units[unitIndex])
}

// Add the missing function for calculating ETA (used by the old code)
func calculateETA(startTime time.Time, currentSize int64, totalSize int64) string {
	if currentSize <= 0 || totalSize <= 0 || startTime.IsZero() {
		return "calculating..."
	}

	elapsed := time.Since(startTime)
	if elapsed <= 0 {
		return "calculating..."
	}

	bytesPerSecond := float64(currentSize) / elapsed.Seconds()
	if bytesPerSecond <= 0 {
		return "calculating..."
	}

	remainingBytes := totalSize - currentSize
	remainingSeconds := float64(remainingBytes) / bytesPerSecond

	duration := time.Duration(remainingSeconds) * time.Second
	if duration < 0 {
		duration = 0
	}

	if duration.Hours() >= 24 {
		return fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60)
	} else if duration.Hours() >= 1 {
		return fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60)
	} else if duration.Minutes() >= 1 {
		return fmt.Sprintf("%dm %ds", int(duration.Minutes()), int(duration.Seconds())%60)
	}
	return fmt.Sprintf("%ds", int(duration.Seconds()))
}
