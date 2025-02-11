package notifications

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
)

var _globalNotificationService *NotificationService

type NotificationService struct {
	ctx                   basecontext.ApiContext
	forceClearLine        bool
	clearLineOnUpdate     bool
	clearProgressOnUpdate bool
	Channel               chan NotificationMessage
	stopChan              chan bool
	progressCounters      map[string]float64
	previousMessage       NotificationMessage
	CurrentMessage        NotificationMessage
}

func New(ctx basecontext.ApiContext) *NotificationService {
	_globalNotificationService := &NotificationService{
		ctx:               ctx,
		Channel:           make(chan NotificationMessage),
		clearLineOnUpdate: false,
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

func (p *NotificationService) NotifyProgress(correlationId string, prefix string, progress float64) {
	p.Notify(NewProgressNotificationMessage(correlationId, prefix, progress))
}

func (p *NotificationService) FinishProgress(correlationId string, prefix string) {
	msg := NewProgressNotificationMessage(correlationId, prefix, 100)
	msg.Close()
	p.Notify(msg)
}

func (p *NotificationService) Stop() {
	p.stopChan <- true
}

func (p *NotificationService) Start() {
	go func() {
		defer close(p.Channel)
		for {
			select {
			case <-p.stopChan:
				return
			case p.CurrentMessage = <-p.Channel:
				progress := 0.0
				shouldLog := false
				if p.CurrentMessage.IsProgress {
					if val, ok := p.progressCounters[p.CurrentMessage.CorrelationId()]; ok {
						progress = val
					} else {
						progress = 0
						p.progressCounters[p.CurrentMessage.CorrelationId()] = progress
					}

					currentProgressStr, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", p.CurrentMessage.CurrentProgress), 64)
					progressStr, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", progress), 64)
					if currentProgressStr > progressStr {
						p.progressCounters[p.CurrentMessage.CorrelationId()] = p.CurrentMessage.CurrentProgress
						shouldLog = true
					}

					if p.CurrentMessage.Closed() {
						p.ResetCounters(p.CurrentMessage.CorrelationId())
					}
				} else {
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
					printMsg += p.CurrentMessage.Message
					if p.CurrentMessage.IsProgress {
						p.CurrentMessage.lastNotificationTime = time.Now()
						printMsg += fmt.Sprintf(" (%.1f%%)", p.CurrentMessage.CurrentProgress)
						eta := ""
						if p.CurrentMessage.TotalSize() > 0 && p.CurrentMessage.CurrentSize() > 0 {
							currentSizeUnit := "b"
							totalSizeUnit := "b"
							currentSize := float64(p.CurrentMessage.CurrentSize())
							totalSize := float64(p.CurrentMessage.TotalSize())
							if currentSize > 1024 {
								currentSizeUnit = "kb"
								currentSize = currentSize / 1024
							}
							if currentSize > 1024 {
								currentSizeUnit = "mb"
								currentSize = currentSize / 1024
							}
							if currentSize > 1024 {
								currentSizeUnit = "gb"
								currentSize = currentSize / 1024
							}
							// total size
							if totalSize > 1024 {
								totalSizeUnit = "kb"
								totalSize = totalSize / 1024
							}
							if totalSize > 1024 {
								totalSizeUnit = "mb"
								totalSize = totalSize / 1024
							}
							if totalSize > 1024 {
								totalSizeUnit = "gb"
								totalSize = totalSize / 1024
							}
							if !p.CurrentMessage.startingTime.IsZero() {
								eta = fmt.Sprintf(" ETA: %s", calculateETA(p.CurrentMessage.startingTime, p.CurrentMessage.CurrentSize(), p.CurrentMessage.TotalSize()))
							}

							printMsg += fmt.Sprintf(" [%.2f %v/%.2f %v]%s", currentSize, currentSizeUnit, totalSize, totalSizeUnit, eta)
						}
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

	// Reset progress counter
	delete(p.progressCounters, correlationId)

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
