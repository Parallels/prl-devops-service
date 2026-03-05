package notifications

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
)

var _globalNotificationService *NotificationService

type RateSample struct {
	Timestamp time.Time
	Size      int64
	Progress  float64
}

type JobStep struct {
	Name          string
	Weight        float64 // Percentage weight of this step, e.g. 40.0 for 40%
	Parallel      bool
	HasPercentage bool
}

type JobWorkflow struct {
	Message      string
	Steps        []JobStep
	StepProgress map[string]float64
	StepMessage  map[string]string
	StepError    map[string]string
	StepValue    map[string]int64
	StepTotal    map[string]int64
	StepFilename map[string]string
	StepState    map[string]constants.JobState
}

type ProgressTracker struct {
	CurrentProgress        float64
	JobPercentage          float64
	LastJobPercentageSent  float64
	LastStepPercentageSent float64
	LastLogPercentageSent  float64
	JobId                  string
	CurrentAction          string
	CurrentActionStep      string
	Filename               string
	LastUpdateTime         time.Time
	LastLogTime            time.Time // Track when we last logged a message
	Prefix                 string
	StartTime              time.Time
	TotalSize              int64
	CurrentSize            int64
	IsComplete             bool
	RateSamples            []RateSample // Store last minute of samples
}

type NotificationService struct {
	ctx                   basecontext.ApiContext
	forceClearLine        bool
	clearLineOnUpdate     bool
	clearProgressOnUpdate bool
	Channel               chan NotificationMessage
	stopChan              chan bool
	activeProgress        map[string]*ProgressTracker // Track active progress notifications
	activeWorkflows       map[string]JobWorkflow      // Track expected steps and weights per JobId
	progressCounters      map[string]float64
	previousMessage       NotificationMessage
	CurrentMessage        NotificationMessage
	mu                    sync.RWMutex // Protects activeProgress map
	queue                 map[string]map[string]NotificationMessage
	qMu                   sync.Mutex

	OnUpdateJobSteps        func(jobId string, steps []data_models.JobStep)
	OnUpdateJobProgress     func(jobId string, percent int, status string)
	OnUpdateJobMessage      func(jobId string, message string)
	OnUpdateJobResultRecord func(jobId string, recordId string, recordType string)
	// OnUpdateJobProgressAndSteps is called instead of the two separate callbacks when available.
	// It writes progress + step snapshot in a single atomic operation, avoiding the race where
	// a separate UpdateJobProgress read could see stale (empty) steps.
	OnUpdateJobProgressAndSteps func(jobId string, percent int, state string, steps []data_models.JobStep)
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
	minSampleInterval         = 2 * time.Second        // Minimum time between rate samples
	uiProgressPushThreshold   = 1.0                    // 1% step change threshold to trigger UI JobManager socket via Hub
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
		activeWorkflows:   make(map[string]JobWorkflow),
		progressCounters:  make(map[string]float64),
		queue:             make(map[string]map[string]NotificationMessage),
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

// RegisterJobWorkflow defines the expected steps and their percent-weights for a given JobId.
// This allows the NotificationService to automatically compute the overall 0-100% JobPercentage
// by combining the current action's completion with the weights.
func (p *NotificationService) RegisterJobWorkflow(jobId string, steps []JobStep) {
	if jobId == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	stepProgress := make(map[string]float64)
	stepMessage := make(map[string]string)
	stepError := make(map[string]string)
	stepValue := make(map[string]int64)
	stepTotal := make(map[string]int64)
	stepFilename := make(map[string]string)
	stepState := make(map[string]constants.JobState)

	if existing, exists := p.activeWorkflows[jobId]; exists {
		if existing.StepProgress != nil {
			stepProgress = existing.StepProgress
		}
		if existing.StepMessage != nil {
			stepMessage = existing.StepMessage
		}
		if existing.StepError != nil {
			stepError = existing.StepError
		}
		if existing.StepValue != nil {
			stepValue = existing.StepValue
		}
		if existing.StepTotal != nil {
			stepTotal = existing.StepTotal
		}
		if existing.StepFilename != nil {
			stepFilename = existing.StepFilename
		}
		if existing.StepState != nil {
			stepState = existing.StepState
		}
	}

	p.activeWorkflows[jobId] = JobWorkflow{
		Message:      "", // Message is not passed to RegisterJobWorkflow, so it should be empty or handled elsewhere
		Steps:        steps,
		StepProgress: stepProgress,
		StepMessage:  stepMessage,
		StepError:    stepError,
		StepValue:    stepValue,
		StepTotal:    stepTotal,
		StepFilename: stepFilename,
		StepState:    stepState,
	}
}

func (p *NotificationService) ResetCounters(correlationId string) {
	if correlationId != "" {
		delete(p.progressCounters, normalizeCorrelationID(correlationId))
	}
}

func (p *NotificationService) Notify(msg *NotificationMessage) {
	p.Channel <- *msg
}

// NotifyJobMessage sends a root-level job message without assigning it to any specific step.
// This is useful for JobStateInit messages and general overall status messages.
func (p *NotificationService) NotifyJobMessage(jobId string, msg string) {
	nMsg := NewNotificationMessage(msg, NotificationMessageLevelInfo).
		SetJobId(jobId).
		SetCurrentAction("") // Empty action marks it as a root-level Job message
	p.Notify(nMsg)
}

type JobLogger struct {
	ns     *NotificationService
	jobId  string
	action string
}

func (p *NotificationService) WithJob(jobId, action string) *JobLogger {
	return &JobLogger{
		ns:     p,
		jobId:  jobId,
		action: action,
	}
}

func (l *JobLogger) NotifyInfo(msg string) {
	nMsg := NewNotificationMessage(msg, NotificationMessageLevelInfo).
		SetJobId(l.jobId).
		SetCurrentAction(l.action)
	l.ns.Notify(nMsg)
}

func (l *JobLogger) NotifyInfof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	nMsg := NewNotificationMessage(msg, NotificationMessageLevelInfo).
		SetJobId(l.jobId).
		SetCurrentAction(l.action)
	l.ns.Notify(nMsg)
}

func (l *JobLogger) NotifyWarning(msg string) {
	nMsg := NewNotificationMessage(msg, NotificationMessageLevelWarning).
		SetJobId(l.jobId).
		SetCurrentAction(l.action)
	l.ns.Notify(nMsg)
}

func (l *JobLogger) NotifyWarningf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	nMsg := NewNotificationMessage(msg, NotificationMessageLevelWarning).
		SetJobId(l.jobId).
		SetCurrentAction(l.action)
	l.ns.Notify(nMsg)
}

func (l *JobLogger) NotifyError(msg string) {
	nMsg := NewNotificationMessage(msg, NotificationMessageLevelError).
		SetJobId(l.jobId).
		SetCurrentAction(l.action)
	l.ns.Notify(nMsg)
}

func (l *JobLogger) NotifyErrorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	nMsg := NewNotificationMessage(msg, NotificationMessageLevelError).
		SetJobId(l.jobId).
		SetCurrentAction(l.action)
	l.ns.Notify(nMsg)
}

func (l *JobLogger) NotifyDebug(msg string) {
	nMsg := NewNotificationMessage(msg, NotificationMessageLevelDebug).
		SetJobId(l.jobId).
		SetCurrentAction(l.action)
	l.ns.Notify(nMsg)
}

func (l *JobLogger) NotifyDebugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	nMsg := NewNotificationMessage(msg, NotificationMessageLevelDebug).
		SetJobId(l.jobId).
		SetCurrentAction(l.action)
	l.ns.Notify(nMsg)
}

// LogInfof implements a standard interface matching ApiContext and forwards to NotifyInfof
func (l *JobLogger) LogInfof(format string, args ...interface{}) {
	l.NotifyInfof(format, args...)
}

// LogErrorf implements a standard interface matching ApiContext and forwards to NotifyErrorf
func (l *JobLogger) LogErrorf(format string, args ...interface{}) {
	l.NotifyErrorf(format, args...)
}

// LogDebugf implements a standard interface matching ApiContext and forwards to NotifyDebugf
func (l *JobLogger) LogDebugf(format string, args ...interface{}) {
	l.NotifyDebugf(format, args...)
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

	// Only add a new sample if enough time has passed and size has changed
	if (len(tracker.RateSamples) == 0 || now.Sub(tracker.RateSamples[len(tracker.RateSamples)-1].Timestamp) >= minSampleInterval) &&
		currentSize != tracker.CurrentSize {

		newSample := RateSample{
			Timestamp: now,
			Size:      currentSize,
			Progress:  progress,
		}

		// Keep only last 5 samples to reduce memory and calculation overhead
		if len(tracker.RateSamples) >= 5 {
			tracker.RateSamples = tracker.RateSamples[1:]
		}
		tracker.RateSamples = append(tracker.RateSamples, newSample)
	}

	if progress > tracker.CurrentProgress {
		tracker.CurrentProgress = progress
	}
	tracker.LastUpdateTime = now
	if currentSize > tracker.CurrentSize {
		tracker.CurrentSize = currentSize
	}
}

func (p *NotificationService) NotifyProgress(correlationId string, prefix string, progress float64) {
	if correlationId == "" {
		return
	}
	msg := NewProgressNotificationMessage(correlationId, prefix, progress)
	encodedID := msg.CorrelationId()

	// Create or update progress tracker
	p.mu.Lock()
	tracker, exists := p.activeProgress[encodedID]
	if !exists {
		tracker = &ProgressTracker{
			StartTime:   time.Now(),
			Prefix:      prefix,
			IsComplete:  false,
			RateSamples: make([]RateSample, 0, 60),
			TotalSize:   msg.totalSize, // Make sure we capture the total size
		}
		p.activeProgress[encodedID] = tracker
	}

	currentSize := msg.currentSize
	if currentSize == 0 {
		currentSize = tracker.CurrentSize
	}
	p.updateProgressTracker(tracker, progress, currentSize)

	if progress >= 100 {
		msg.Close()
		tracker.IsComplete = true
	}
	p.mu.Unlock()

	p.Notify(msg)
}

func (p *NotificationService) FinishProgress(correlationId string, prefix string) {
	if correlationId == "" {
		return
	}
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.Lock()
	if tracker, exists := p.activeProgress[encodedID]; exists {
		tracker.CurrentProgress = 100
		tracker.LastUpdateTime = time.Now()
		tracker.IsComplete = true
	}
	p.mu.Unlock()

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

	go p.ingestMessages()
	go p.processQueueWorker()
}

func (p *NotificationService) ingestMessages() {
	defer close(p.Channel)
	for {
		select {
		case <-p.stopChan:
			return
		case msg := <-p.Channel:
			if msg.JobId == "" && msg.CorrelationId() == "" && msg.Message == "" {
				continue // Drop completely empty payloads
			}
			p.qMu.Lock()
			encodedID := msg.CorrelationId()
			if p.queue[encodedID] == nil {
				p.queue[encodedID] = make(map[string]NotificationMessage)
			}
			actionKey := msg.CurrentAction
			if actionKey == "" {
				actionKey = "default"
			}
			p.queue[encodedID][actionKey] = msg
			p.qMu.Unlock()
		}
	}
}

func (p *NotificationService) processQueueWorker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.qMu.Lock()
			inflight := p.queue
			p.queue = make(map[string]map[string]NotificationMessage)
			p.qMu.Unlock()

			if len(inflight) > 0 {
				p.processInflightQueue(inflight)
			}
		}
	}
}

func (p *NotificationService) processInflightQueue(inflight map[string]map[string]NotificationMessage) {
	for _, actionMsgs := range inflight {
		for _, msg := range actionMsgs {
			p.processSingleMessage(msg)
		}
	}
}

func (p *NotificationService) processSingleMessage(msg NotificationMessage) {
	p.CurrentMessage = msg
	shouldLog := false

	if p.CurrentMessage.IsProgress {
		p.mu.Lock()
		tracker, exists := p.activeProgress[p.CurrentMessage.CorrelationId()]
		if !exists {
			// New progress notification
			if p.CurrentMessage.CurrentProgress < 100 {
				tracker = &ProgressTracker{
					StartTime:              time.Now(),
					Prefix:                 p.CurrentMessage.Message,
					CurrentProgress:        p.CurrentMessage.CurrentProgress,
					JobPercentage:          p.CurrentMessage.JobPercentage,
					LastJobPercentageSent:  -1.0,
					LastStepPercentageSent: p.CurrentMessage.CurrentProgress,
					LastLogPercentageSent:  -1.0,
					JobId:                  p.CurrentMessage.JobId,
					CurrentAction:          p.CurrentMessage.CurrentAction,
					CurrentActionStep:      p.CurrentMessage.CurrentActionStep,
					Filename:               p.CurrentMessage.Filename,
					LastUpdateTime:         time.Now(),
					CurrentSize:            p.CurrentMessage.currentSize,
					TotalSize:              p.CurrentMessage.totalSize,
					RateSamples:            make([]RateSample, 0, 60),
				}
				p.activeProgress[p.CurrentMessage.CorrelationId()] = tracker
				shouldLog = true

				if tracker.JobId != "" && p.OnUpdateJobSteps != nil {
					p.publishJobSteps(tracker.JobId)
				}
			} else if p.CurrentMessage.JobId != "" {
				// Instant-complete step
				transient := &ProgressTracker{
					StartTime:       time.Now(),
					CurrentProgress: 100,
					JobId:           p.CurrentMessage.JobId,
					CurrentAction:   p.CurrentMessage.CurrentAction,
					LastUpdateTime:  time.Now(),
				}
				if workflow, hasWorkflow := p.activeWorkflows[p.CurrentMessage.JobId]; hasWorkflow {
					autoSkipPreviousSteps(&workflow, transient.CurrentAction)
					workflow.StepProgress[transient.CurrentAction] = 100.0
					if p.CurrentMessage.totalSize > 0 {
						workflow.StepTotal[transient.CurrentAction] = p.CurrentMessage.totalSize
						workflow.StepValue[transient.CurrentAction] = p.CurrentMessage.totalSize
					}
					if p.CurrentMessage.Filename != "" {
						workflow.StepFilename[transient.CurrentAction] = p.CurrentMessage.Filename
					}

					var calculatedJobPercentage float64
					for _, step := range workflow.Steps {
						if prog, exists := workflow.StepProgress[step.Name]; exists {
							calculatedJobPercentage += (prog / 100.0) * step.Weight
						}
					}
					if calculatedJobPercentage > 100 {
						calculatedJobPercentage = 100
					}
					transient.JobPercentage = calculatedJobPercentage
					p.activeWorkflows[p.CurrentMessage.JobId] = workflow
					p.publishJobSteps(p.CurrentMessage.JobId)
				}

				if p.OnUpdateJobProgress != nil && transient.JobPercentage > 0 {
					p.OnUpdateJobProgress(transient.JobId, int(transient.JobPercentage), "running")
				}
				shouldLog = true
			}
		} else {
			// Update existing tracker
			p.updateProgressTracker(tracker, p.CurrentMessage.CurrentProgress, p.CurrentMessage.currentSize)

			tracker.CurrentAction = p.CurrentMessage.CurrentAction
			tracker.CurrentActionStep = p.CurrentMessage.CurrentActionStep
			tracker.Filename = p.CurrentMessage.Filename

			if tracker.JobId != "" {
				if workflow, hasWorkflow := p.activeWorkflows[tracker.JobId]; hasWorkflow {
					autoSkipPreviousSteps(&workflow, tracker.CurrentAction)
					if tracker.CurrentProgress >= 100 {
						workflow.StepProgress[tracker.CurrentAction] = 100.0
					} else {
						workflow.StepProgress[tracker.CurrentAction] = tracker.CurrentProgress
					}

					if tracker.TotalSize > 0 {
						workflow.StepTotal[tracker.CurrentAction] = tracker.TotalSize
						workflow.StepValue[tracker.CurrentAction] = tracker.CurrentSize
					}
					if tracker.Filename != "" {
						workflow.StepFilename[tracker.CurrentAction] = tracker.Filename
					}

					var calculatedJobPercentage float64
					furthestAction := tracker.CurrentAction
					var furthestWeight float64 = -1.0
					var furthestProgress float64 = tracker.CurrentProgress

					for _, step := range workflow.Steps {
						if prog, exists := workflow.StepProgress[step.Name]; exists {
							calculatedJobPercentage += (prog / 100.0) * step.Weight
							if prog > 0 {
								furthestAction = step.Name
								furthestWeight = step.Weight
								furthestProgress = prog
							}
						}
					}
					if furthestWeight != -1.0 {
						tracker.CurrentAction = furthestAction
						tracker.CurrentProgress = furthestProgress
					}
					if calculatedJobPercentage > 100 {
						calculatedJobPercentage = 100
					}
					tracker.JobPercentage = calculatedJobPercentage
					p.activeWorkflows[tracker.JobId] = workflow

					// Single atomic update: progress + steps in one DB write + one event broadcast.
					// This prevents the race where separate UpdateJobSteps/UpdateJobProgress calls
					// could emit a stale no-steps event.
					if p.OnUpdateJobProgressAndSteps != nil && int(calculatedJobPercentage) > 0 {
						dtos := p.buildStepDTOs(tracker.JobId)
						p.OnUpdateJobProgressAndSteps(tracker.JobId, int(calculatedJobPercentage), "running", dtos)
					} else {
						// Fallback: fire separately (old path, kept for compatibility)
						p.publishJobSteps(tracker.JobId)
						if tracker.JobPercentage > 0 && p.OnUpdateJobProgress != nil {
							p.OnUpdateJobProgress(tracker.JobId, int(tracker.JobPercentage), "running")
						}
					}
				} else {
					tracker.JobPercentage = p.CurrentMessage.JobPercentage
				}
			} else {
				tracker.JobPercentage = p.CurrentMessage.JobPercentage
			}

			// Since we're ticking exactly once a second, skip traditional throttling checks and just print what we have
			shouldLog = true

			// Only call OnUpdateJobProgress when we did NOT call OnUpdateJobProgressAndSteps
			if tracker.JobId != "" && p.OnUpdateJobProgressAndSteps == nil {
				if tracker.JobPercentage > 0 && p.OnUpdateJobProgress != nil {
					p.OnUpdateJobProgress(tracker.JobId, int(tracker.JobPercentage), "running")
				}
			}
		}

		shouldDelete := false
		if p.CurrentMessage.Closed() {
			shouldDelete = true
		} else if tracker != nil && tracker.JobId != "" && len(p.activeWorkflows[tracker.JobId].Steps) > 0 {
			if tracker.JobPercentage >= 100 {
				shouldDelete = true
			}
		} else if p.CurrentMessage.CurrentProgress >= 100 {
			shouldDelete = true
		}

		if shouldDelete {
			delete(p.activeProgress, p.CurrentMessage.CorrelationId())
			if tracker != nil && tracker.JobId != "" {
				delete(p.activeWorkflows, tracker.JobId)
			}
		}
		p.mu.Unlock()
	} else {
		if p.CurrentMessage.Message != "" {
			shouldLog = true
		}

		if p.CurrentMessage.JobId != "" {
			p.mu.Lock()
			if workflow, hasWorkflow := p.activeWorkflows[p.CurrentMessage.JobId]; hasWorkflow {
				if p.CurrentMessage.CurrentAction != "" {
					if p.CurrentMessage.Level == NotificationMessageLevelError {
						workflow.StepError[p.CurrentMessage.CurrentAction] = p.CurrentMessage.Message
					} else {
						workflow.StepMessage[p.CurrentMessage.CurrentAction] = p.CurrentMessage.Message
					}
					p.activeWorkflows[p.CurrentMessage.JobId] = workflow
					p.publishJobSteps(p.CurrentMessage.JobId)
				} else {
					// Root-level Job message (CurrentAction is empty)
					workflow.Message = p.CurrentMessage.Message
					p.activeWorkflows[p.CurrentMessage.JobId] = workflow
					if p.OnUpdateJobMessage != nil {
						p.OnUpdateJobMessage(p.CurrentMessage.JobId, p.CurrentMessage.Message)
					}
				}
			} else if p.CurrentMessage.CurrentAction == "" && p.OnUpdateJobMessage != nil {
				// No active workflow, but it's a root-level job message
				p.OnUpdateJobMessage(p.CurrentMessage.JobId, p.CurrentMessage.Message)
			}
			p.mu.Unlock()
		}
	}

	if p.CurrentMessage.Message != p.previousMessage.Message && !p.forceClearLine {
		p.previousMessage = p.CurrentMessage
		p.clearLineOnUpdate = false
	}

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
			p.mu.RLock()
			tracker, exists := p.activeProgress[p.CurrentMessage.CorrelationId()]
			p.mu.RUnlock()
			if exists {
				baseMsg := p.CurrentMessage.Message
				if baseMsg == "" {
					baseMsg = tracker.Prefix
				}
				printMsg += baseMsg + " "
				printMsg += formatProgressMessage(&p.CurrentMessage, tracker, p)
			} else {
				printMsg += fmt.Sprintf("%s (%.1f%%)", p.CurrentMessage.Message, p.CurrentMessage.CurrentProgress)
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

// buildStepDTOs assembles the current step snapshot for a job.
// Caller must hold p.mu (read lock is sufficient).
func (p *NotificationService) buildStepDTOs(jobId string) []data_models.JobStep {
	workflow, ok := p.activeWorkflows[jobId]
	if !ok {
		return nil
	}

	var dtos []data_models.JobStep
	for _, stepInfo := range workflow.Steps {
		prog := workflow.StepProgress[stepInfo.Name]
		state := constants.JobStatePending
		if prog > 0 && prog < 100 {
			state = constants.JobStateRunning
		} else if prog >= 100 {
			state = constants.JobStateCompleted
			prog = 100.0 // Ensure it's exactly 100 in the UI
		}

		if explicitState, ok := workflow.StepState[stepInfo.Name]; ok && explicitState != "" {
			state = explicitState
		}

		dto := data_models.JobStep{
			Name:              stepInfo.Name,
			Weight:            stepInfo.Weight,
			Parallel:          stepInfo.Parallel,
			HasPercentage:     stepInfo.HasPercentage,
			State:             state,
			CurrentPercentage: prog,
		}

		// Inject active tracker metadata (bytes, ETA, filename) from persisted workflow maps
		if total, hasTotal := workflow.StepTotal[stepInfo.Name]; hasTotal && total > 0 {
			dto.Total = total
			if val, hasVal := workflow.StepValue[stepInfo.Name]; hasVal {
				if prog >= 100 {
					dto.Value = total // Ensure Value matches Total when completed
				} else {
					dto.Value = val
				}
			}
			dto.Unit = "bytes"
		}
		if fname, hasFname := workflow.StepFilename[stepInfo.Name]; hasFname {
			dto.Filename = fname
		}

		// Still inject ETA from active trackers if it's currently running
		for _, v := range p.activeProgress {
			if v.JobId == jobId && v.CurrentAction == stepInfo.Name {
				dto.ETA = calculateETA(v.StartTime, v.CurrentSize, v.TotalSize)
			}
		}

		if msg, ok := workflow.StepMessage[stepInfo.Name]; ok {
			dto.Message = msg
		}
		if errStr, ok := workflow.StepError[stepInfo.Name]; ok {
			dto.Error = errStr
		}

		dtos = append(dtos, dto)
	}
	return dtos
}

func (p *NotificationService) publishJobSteps(jobId string) {
	if p.OnUpdateJobSteps == nil || jobId == "" {
		return
	}
	dtos := p.buildStepDTOs(jobId)
	if dtos != nil {
		p.OnUpdateJobSteps(jobId, dtos)
	}
}

func (p *NotificationService) UpdateJobResultRecord(jobId string, recordId string, recordType string) {
	if p.OnUpdateJobResultRecord != nil {
		p.OnUpdateJobResultRecord(jobId, recordId, recordType)
	}
}

func (p *NotificationService) Restart() {
	p.Stop()
	p.Start()
}

func ClearLine() {
	fmt.Printf("\r\033[K")
}

func autoSkipPreviousSteps(workflow *JobWorkflow, currentAction string) {
	currentIdx := -1
	for i, step := range workflow.Steps {
		if step.Name == currentAction {
			currentIdx = i
			break
		}
	}
	if currentIdx > 0 {
		for i := 0; i < currentIdx; i++ {
			prevStep := workflow.Steps[i]
			if !prevStep.Parallel && workflow.StepProgress[prevStep.Name] < 100 {
				workflow.StepProgress[prevStep.Name] = 100.0
				if workflow.StepState == nil {
					workflow.StepState = make(map[string]constants.JobState)
				}
				workflow.StepState[prevStep.Name] = constants.JobStateSkipped
				if workflow.StepMessage == nil {
					workflow.StepMessage = make(map[string]string)
				}
				workflow.StepMessage[prevStep.Name] = "Step implicitly skipped"
			}
		}
	}
}

func (p *NotificationService) SkipStep(jobId string, stepName string, message string) {
	if jobId == "" || stepName == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if workflow, exists := p.activeWorkflows[jobId]; exists {
		if workflow.StepProgress == nil {
			workflow.StepProgress = make(map[string]float64)
		}
		if workflow.StepState == nil {
			workflow.StepState = make(map[string]constants.JobState)
		}
		if workflow.StepMessage == nil {
			workflow.StepMessage = make(map[string]string)
		}
		workflow.StepProgress[stepName] = 100.0
		workflow.StepState[stepName] = constants.JobStateSkipped
		if message != "" {
			workflow.StepMessage[stepName] = message
		} else {
			workflow.StepMessage[stepName] = "Step skipped"
		}

		var calculatedJobPercentage float64
		for _, step := range workflow.Steps {
			if prog, exists := workflow.StepProgress[step.Name]; exists {
				calculatedJobPercentage += (prog / 100.0) * step.Weight
			}
		}
		if calculatedJobPercentage > 100 {
			calculatedJobPercentage = 100
		}

		p.activeWorkflows[jobId] = workflow

		if p.OnUpdateJobProgressAndSteps != nil && int(calculatedJobPercentage) > 0 {
			dtos := p.buildStepDTOs(jobId)
			p.OnUpdateJobProgressAndSteps(jobId, int(calculatedJobPercentage), "running", dtos)
		} else {
			p.publishJobSteps(jobId)
		}
	}
}

// CleanupNotifications removes all progress tracking for a specific correlation ID
func (p *NotificationService) CleanupNotifications(correlationId string) {
	if correlationId == "" {
		return
	}

	encodedID := normalizeCorrelationID(correlationId)
	p.cleanupNotificationsByEncodedID(encodedID)
	p.ctx.LogDebugf("Cleaned up notifications for correlation ID: %s", correlationId)
}

func (p *NotificationService) cleanupNotificationsByEncodedID(encodedID string) {
	if encodedID == "" {
		return
	}

	// Remove from active progress tracking
	p.mu.Lock()
	delete(p.activeProgress, encodedID)
	p.mu.Unlock()

	// Reset previous message if it was for this correlation ID
	if p.previousMessage.correlationId == encodedID {
		p.previousMessage = NotificationMessage{}
	}

	// Reset current message if it was for this correlation ID
	if p.CurrentMessage.correlationId == encodedID {
		p.CurrentMessage = NotificationMessage{}
	}
}

// GetActiveProgressCount returns the number of active progress notifications
func (p *NotificationService) GetActiveProgressCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.activeProgress)
}

// GetActiveProgressIDs returns a slice of correlation IDs for active progress notifications
func (p *NotificationService) GetActiveProgressIDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ids := make([]string, 0, len(p.activeProgress))
	for id := range p.activeProgress {
		ids = append(ids, id)
	}
	return ids
}

// IsProgressActive checks if a progress notification is active for the given correlation ID
func (p *NotificationService) IsProgressActive(correlationId string) bool {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, exists := p.activeProgress[encodedID]
	return exists
}

// GetProgressStatus returns the current progress status for a given correlation ID
// Returns progress percentage and whether the progress exists
func (p *NotificationService) GetProgressStatus(correlationId string) (float64, bool) {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	if tracker, exists := p.activeProgress[encodedID]; exists {
		return tracker.CurrentProgress, true
	}
	return 0, false
}

// CleanupStaleProgress removes progress notifications that haven't been updated
// for longer than the specified duration
func (p *NotificationService) CleanupStaleProgress(staleDuration time.Duration) {
	now := time.Now()
	var idsToCleanup []string

	// First, collect IDs that need cleanup while holding read lock
	p.mu.RLock()
	for id, tracker := range p.activeProgress {
		if now.Sub(tracker.LastUpdateTime) > staleDuration {
			idsToCleanup = append(idsToCleanup, id)
		}
	}
	p.mu.RUnlock()

	// Then cleanup each ID (this will acquire write lock)
	for _, id := range idsToCleanup {
		decodedID, err := decodeCorrelationID(id)
		if err != nil || decodedID == "" {
			decodedID = id
		}
		p.ctx.LogDebugf("Cleaning up stale progress for correlation ID: %s", decodedID)
		p.cleanupNotificationsByEncodedID(id)
	}
}

// GetProgressDuration returns the duration since the progress started
func (p *NotificationService) GetProgressDuration(correlationId string) (time.Duration, bool) {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	if tracker, exists := p.activeProgress[encodedID]; exists {
		return time.Since(tracker.StartTime), true
	}
	return 0, false
}

// GetProgressRate calculates transfer and progress rates for a given correlation ID
func (p *NotificationService) GetProgressRate(correlationId string) (*ProgressRate, bool) {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	tracker, exists := p.activeProgress[encodedID]
	if !exists || tracker.TotalSize <= 0 {
		return nil, false
	}

	totalDuration := time.Since(tracker.StartTime).Seconds()
	if totalDuration <= 0 {
		return nil, false
	}

	rate := &ProgressRate{
		BytesPerSecond:    float64(tracker.CurrentSize) / totalDuration,
		ProgressPerSecond: tracker.CurrentProgress / totalDuration,
	}
	rate.RecentBytesPerSecond = rate.BytesPerSecond

	return rate, true
}

// PredictTimeRemaining estimates the time remaining based on recent progress
func (p *NotificationService) PredictTimeRemaining(correlationId string) (time.Duration, bool) {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	tracker, exists := p.activeProgress[encodedID]
	if !exists || tracker.TotalSize <= 0 {
		return 0, false
	}

	elapsed := time.Since(tracker.StartTime)
	if elapsed <= 0 {
		return 0, false
	}

	bytesPerSecond := float64(tracker.CurrentSize) / elapsed.Seconds()
	if bytesPerSecond <= 0 {
		return 0, false
	}

	remainingBytes := float64(tracker.TotalSize - tracker.CurrentSize)
	remainingSeconds := remainingBytes / bytesPerSecond

	if remainingSeconds < 0 {
		remainingSeconds = 0
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

func (p *NotificationService) shouldLogProgress(tracker *ProgressTracker, currentProgress float64) bool {
	now := time.Now()

	if currentProgress >= 100 && tracker.LastLogPercentageSent < 100 {
		tracker.LastLogPercentageSent = 100
		tracker.LastLogTime = now
		return true
	}

	if currentProgress <= 0 && tracker.LastLogPercentageSent < 0 {
		tracker.LastLogPercentageSent = 0
		tracker.LastLogTime = now
		return true
	}

	timeSinceLastLog := now.Sub(tracker.LastLogTime)

	currentDecile := int(currentProgress / 10)
	lastDecile := int(tracker.LastLogPercentageSent / 10)
	if tracker.LastLogPercentageSent < 0 {
		lastDecile = -1
	}

	// Log if we cross a 10% boundary, or if 1 second has passed
	if currentDecile > lastDecile || timeSinceLastLog >= time.Second {
		tracker.LastLogPercentageSent = currentProgress
		tracker.LastLogTime = now
		return true
	}

	return false
}

// Update the message formatting in Start() method
func formatProgressMessage(msg *NotificationMessage, tracker *ProgressTracker, p *NotificationService) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("%.1f%%", msg.CurrentProgress))

	if msg.TotalSize() > 0 && msg.CurrentSize() > 0 {
		current := formatSize(float64(msg.CurrentSize()))
		total := formatSize(float64(msg.TotalSize()))
		parts = append(parts, fmt.Sprintf("[%s/%s]", current, total))

		elapsed := time.Since(tracker.StartTime)
		if elapsed > 0 {
			bytesPerSecond := float64(tracker.CurrentSize) / elapsed.Seconds()
			parts = append(parts, FormatTransferRate(bytesPerSecond))

			if remainingTime := calculateETA(tracker.StartTime, tracker.CurrentSize, tracker.TotalSize); remainingTime != "calculating..." {
				parts = append(parts, fmt.Sprintf("ETA: %s", remainingTime))
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
