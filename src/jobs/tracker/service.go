package tracker

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
)

var _globalJobProgressService *JobProgressService

type JobProgressService struct {
	ctx                   basecontext.ApiContext
	forceClearLine        bool
	clearLineOnUpdate     bool
	clearProgressOnUpdate bool
	Channel               chan JobMessage
	stopChan              chan bool
	activeProgress        map[string]*ProgressTracker
	activeWorkflows       map[string]JobWorkflow
	progressCounters      map[string]float64
	previousMessage       JobMessage
	CurrentMessage        JobMessage
	mu                    sync.RWMutex
	queue                 map[string]map[string]JobMessage
	qMu                   sync.Mutex

	OnUpdateJobSteps        func(jobId string, steps []data_models.JobStep)
	OnInitJob               func(jobId string)
	OnUpdateJobProgress     func(jobId string, percent int, status string)
	OnUpdateJobMessage      func(jobId string, message string)
	OnUpdateJobResultRecord func(jobId string, recordId string, recordType string)
	// OnUpdateJobProgressAndSteps is called instead of the two separate callbacks when available.
	// It writes progress + step snapshot in a single atomic operation, avoiding the race where
	// a separate UpdateJobProgress read could see stale (empty) steps.
	OnUpdateJobProgressAndSteps func(jobId string, percent int, state string, steps []data_models.JobStep)
}

func NewProgressService(ctx basecontext.ApiContext) *JobProgressService {
	_globalJobProgressService = &JobProgressService{
		ctx:               ctx,
		Channel:           make(chan JobMessage),
		stopChan:          make(chan bool),
		clearLineOnUpdate: false,
		activeProgress:    make(map[string]*ProgressTracker),
		activeWorkflows:   make(map[string]JobWorkflow),
		progressCounters:  make(map[string]float64),
		queue:             make(map[string]map[string]JobMessage),
	}

	_globalJobProgressService.Start()
	return _globalJobProgressService
}

func GetProgressService() *JobProgressService {
	if _globalJobProgressService == nil {
		ctx := basecontext.NewBaseContext()
		_globalJobProgressService = NewProgressService(ctx)
	}
	return _globalJobProgressService
}

func (p *JobProgressService) EnableSingleLineOutput() *JobProgressService {
	p.clearLineOnUpdate = true
	p.forceClearLine = true
	return p
}

func (p *JobProgressService) SetContext(ctx basecontext.ApiContext) *JobProgressService {
	p.ctx = ctx
	return p
}

func (p *JobProgressService) InitJob(jobId string) {
	if jobId == "" {
		return
	}

	if p.OnInitJob != nil {
		p.OnInitJob(jobId)
	}
}

// RegisterJobWorkflow defines the expected steps and their percent-weights for a given JobId.
// This allows the JobProgressService to automatically compute the overall 0-100% JobPercentage
// by combining the current action's completion with the weights.
func (p *JobProgressService) RegisterJobWorkflow(jobId string, steps []JobStep) {
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
		Steps:        steps,
		StepProgress: stepProgress,
		StepMessage:  stepMessage,
		StepError:    stepError,
		StepValue:    stepValue,
		StepTotal:    stepTotal,
		StepFilename: stepFilename,
		StepState:    stepState,
	}

	// Emit the initial step structure immediately so the UI reflects the workflow
	// as soon as it is registered, without waiting for the queue ticker.
	if p.OnUpdateJobProgressAndSteps != nil {
		dtos := p.buildStepDTOs(jobId)
		p.OnUpdateJobProgressAndSteps(jobId, 0, "running", dtos)
	} else {
		p.publishJobSteps(jobId)
	}
}

func (p *JobProgressService) ResetCounters(correlationId string) {
	if correlationId != "" {
		delete(p.progressCounters, normalizeCorrelationID(correlationId))
	}
}

func (p *JobProgressService) Notify(msg *JobMessage) {
	p.Channel <- *msg
}

// NotifyJobMessage sends a root-level job message without assigning it to any specific step.
// It is processed immediately (bypasses the queue) to ensure the UI reflects the update at once.
func (p *JobProgressService) NotifyJobMessage(jobId string, msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	p.ctx.LogInfof("[job:%s] %s", jobId, formatted)
	if jobId != "" && p.OnUpdateJobMessage != nil {
		p.OnUpdateJobMessage(jobId, formatted)
	}
}

type JobLogger struct {
	ns     *JobProgressService
	jobId  string
	action string
}

func (p *JobProgressService) WithJob(jobId, action string) *JobLogger {
	return &JobLogger{
		ns:     p,
		jobId:  jobId,
		action: action,
	}
}

func (l *JobLogger) NotifyInfo(msg string) {
	l.ns.Notify(NewJobMessage(msg, JobMessageLevelInfo).WithJob(l.jobId, l.action))
}

func (l *JobLogger) NotifyInfof(format string, args ...interface{}) {
	l.ns.Notify(NewJobMessage(fmt.Sprintf(format, args...), JobMessageLevelInfo).WithJob(l.jobId, l.action))
}

func (l *JobLogger) NotifyWarning(msg string) {
	l.ns.Notify(NewJobMessage(msg, JobMessageLevelWarning).WithJob(l.jobId, l.action))
}

func (l *JobLogger) NotifyWarningf(format string, args ...interface{}) {
	l.ns.Notify(NewJobMessage(fmt.Sprintf(format, args...), JobMessageLevelWarning).WithJob(l.jobId, l.action))
}

func (l *JobLogger) NotifyError(msg string) {
	l.ns.Notify(NewJobMessage(msg, JobMessageLevelError).WithJob(l.jobId, l.action))
}

func (l *JobLogger) NotifyErrorf(format string, args ...interface{}) {
	l.ns.Notify(NewJobMessage(fmt.Sprintf(format, args...), JobMessageLevelError).WithJob(l.jobId, l.action))
}

func (l *JobLogger) NotifyDebug(msg string) {
	l.ns.Notify(NewJobMessage(msg, JobMessageLevelDebug).WithJob(l.jobId, l.action))
}

func (l *JobLogger) NotifyDebugf(format string, args ...interface{}) {
	l.ns.Notify(NewJobMessage(fmt.Sprintf(format, args...), JobMessageLevelDebug).WithJob(l.jobId, l.action))
}

func (l *JobLogger) StartStep(message string) {
	l.ns.StartStep(l.jobId, l.action, message)
}

func (l *JobLogger) StartStepf(format string, args ...any) {
	l.ns.StartStepf(l.jobId, l.action, format, args...)
}

func (l *JobLogger) UpdateStepProgress(progress float64) {
	l.ns.UpdateStepProgress(l.jobId, l.action, progress)
}

func (l *JobLogger) UpdateStepMessage(message string) {
	l.ns.UpdateStepMessage(l.jobId, l.action, message)
}

func (l *JobLogger) UpdateStepMessagef(format string, args ...any) {
	l.ns.UpdateStepMessagef(l.jobId, l.action, format, args...)
}

func (l *JobLogger) SkipStep(message string) {
	l.ns.SkipStep(l.jobId, l.action, message)
}

func (l *JobLogger) SkipStepf(format string, args ...any) {
	l.ns.SkipStepf(l.jobId, l.action, format, args...)
}

func (l *JobLogger) CompleteStep(message string) {
	l.ns.CompleteStep(l.jobId, l.action, message)
}

func (l *JobLogger) CompleteStepf(format string, args ...any) {
	l.ns.CompleteStepf(l.jobId, l.action, format, args...)
}

func (l *JobLogger) CompleteStepWithFile(filename string, message string) {
	l.ns.CompleteStepWithFile(l.jobId, l.action, filename, message)
}

func (l *JobLogger) FailStep(message string) {
	l.ns.FailStep(l.jobId, l.action, message)
}

func (l *JobLogger) FailStepf(format string, args ...any) {
	l.ns.FailStepf(l.jobId, l.action, format, args...)
}

func (l *JobLogger) FailJob(message string) {
	l.ns.FailJob(l.jobId, message)
}

func (l *JobLogger) FailJobf(format string, args ...any) {
	l.ns.FailJobf(l.jobId, format, args...)
}

// LogInfof implements a standard interface matching ApiContext and forwards to NotifyInfof.
func (l *JobLogger) LogInfof(format string, args ...interface{}) {
	l.NotifyInfof(format, args...)
}

// LogErrorf implements a standard interface matching ApiContext and forwards to NotifyErrorf.
func (l *JobLogger) LogErrorf(format string, args ...interface{}) {
	l.NotifyErrorf(format, args...)
}

// LogDebugf implements a standard interface matching ApiContext and forwards to NotifyDebugf.
func (l *JobLogger) LogDebugf(format string, args ...interface{}) {
	l.NotifyDebugf(format, args...)
}

func (p *JobProgressService) NotifyInfo(msg string) {
	p.Notify(NewJobMessage(msg, JobMessageLevelInfo))
}

func (p *JobProgressService) NotifyInfof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Notify(NewJobMessage(msg, JobMessageLevelInfo))
}

func (p *JobProgressService) NotifyWarning(msg string) {
	p.Notify(NewJobMessage(msg, JobMessageLevelWarning))
}

func (p *JobProgressService) NotifyWarningf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Notify(NewJobMessage(msg, JobMessageLevelWarning))
}

func (p *JobProgressService) NotifyError(msg string) {
	p.Notify(NewJobMessage(msg, JobMessageLevelError))
}

func (p *JobProgressService) NotifyErrorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Notify(NewJobMessage(msg, JobMessageLevelError))
}

func (p *JobProgressService) NotifyDebug(msg string) {
	p.Notify(NewJobMessage(msg, JobMessageLevelDebug))
}

func (p *JobProgressService) NotifyDebugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Notify(NewJobMessage(msg, JobMessageLevelDebug))
}

func (p *JobProgressService) updateProgressTracker(pt *ProgressTracker, progress float64, currentSize int64) {
	now := time.Now()

	if (len(pt.RateSamples) == 0 || now.Sub(pt.RateSamples[len(pt.RateSamples)-1].Timestamp) >= minSampleInterval) &&
		currentSize != pt.CurrentSize {

		newSample := RateSample{
			Timestamp: now,
			Size:      currentSize,
			Progress:  progress,
		}

		if len(pt.RateSamples) >= 5 {
			pt.RateSamples = pt.RateSamples[1:]
		}
		pt.RateSamples = append(pt.RateSamples, newSample)
	}

	if progress > pt.CurrentProgress {
		pt.CurrentProgress = progress
	}
	pt.LastUpdateTime = now
	if currentSize > pt.CurrentSize {
		pt.CurrentSize = currentSize
	}
}

func (p *JobProgressService) NotifyProgress(correlationId string, prefix string, progress float64) {
	if correlationId == "" {
		return
	}
	msg := NewJobProgressMessage(correlationId, prefix, progress)
	encodedID := msg.CorrelationId()

	p.mu.Lock()
	pt, exists := p.activeProgress[encodedID]
	if !exists {
		pt = &ProgressTracker{
			StartTime:   time.Now(),
			Prefix:      prefix,
			IsComplete:  false,
			RateSamples: make([]RateSample, 0, 60),
			TotalSize:   msg.totalSize,
		}
		p.activeProgress[encodedID] = pt
	}

	currentSize := msg.currentSize
	if currentSize == 0 {
		currentSize = pt.CurrentSize
	}
	p.updateProgressTracker(pt, progress, currentSize)

	if progress >= 100 {
		msg.Close()
		pt.IsComplete = true
	}
	p.mu.Unlock()

	p.Notify(msg)
}

func (p *JobProgressService) FinishProgress(correlationId string, prefix string) {
	if correlationId == "" {
		return
	}
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.Lock()
	if pt, exists := p.activeProgress[encodedID]; exists {
		pt.CurrentProgress = 100
		pt.LastUpdateTime = time.Now()
		pt.IsComplete = true
	}
	p.mu.Unlock()

	msg := NewJobProgressMessage(correlationId, prefix, 100)
	msg.Close()
	p.Notify(msg)
}

func (p *JobProgressService) Stop() {
	p.stopChan <- true
}

func (p *JobProgressService) Start() {
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

func (p *JobProgressService) ingestMessages() {
	defer close(p.Channel)
	for {
		select {
		case <-p.stopChan:
			return
		case msg := <-p.Channel:
			if msg.JobId == "" && msg.CorrelationId() == "" && msg.Message == "" {
				continue
			}
			p.qMu.Lock()
			encodedID := msg.CorrelationId()
			if p.queue[encodedID] == nil {
				p.queue[encodedID] = make(map[string]JobMessage)
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

func (p *JobProgressService) processQueueWorker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.qMu.Lock()
			inflight := p.queue
			p.queue = make(map[string]map[string]JobMessage)
			p.qMu.Unlock()

			if len(inflight) > 0 {
				p.processInflightQueue(inflight)
			}
		}
	}
}

func (p *JobProgressService) processInflightQueue(inflight map[string]map[string]JobMessage) {
	for _, actionMsgs := range inflight {
		for _, msg := range actionMsgs {
			p.processSingleMessage(msg)
		}
	}
}

func (p *JobProgressService) processSingleMessage(msg JobMessage) {
	p.CurrentMessage = msg
	shouldLog := false

	if p.CurrentMessage.IsProgress {
		p.mu.Lock()
		pt, exists := p.activeProgress[p.CurrentMessage.CorrelationId()]
		if !exists {
			if p.CurrentMessage.CurrentProgress < 100 {
				pt = &ProgressTracker{
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
				p.activeProgress[p.CurrentMessage.CorrelationId()] = pt
				shouldLog = true

				// First message for this step: seed the workflow step progress so
				// job percentage immediately reflects this step starting.
				if pt.JobId != "" {
					if workflow, hasWorkflow := p.activeWorkflows[pt.JobId]; hasWorkflow {
						autoSkipPreviousSteps(&workflow, pt.CurrentAction)
						workflow.StepProgress[pt.CurrentAction] = pt.CurrentProgress
						if pt.TotalSize > 0 {
							workflow.StepTotal[pt.CurrentAction] = pt.TotalSize
							workflow.StepValue[pt.CurrentAction] = pt.CurrentSize
						}
						if pt.Filename != "" {
							workflow.StepFilename[pt.CurrentAction] = pt.Filename
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
						pt.JobPercentage = calculatedJobPercentage
						p.activeWorkflows[pt.JobId] = workflow

						if p.OnUpdateJobProgressAndSteps != nil && int(calculatedJobPercentage) > 0 {
							dtos := p.buildStepDTOs(pt.JobId)
							p.OnUpdateJobProgressAndSteps(pt.JobId, int(calculatedJobPercentage), "running", dtos)
						} else {
							p.publishJobSteps(pt.JobId)
							if calculatedJobPercentage > 0 && p.OnUpdateJobProgress != nil {
								p.OnUpdateJobProgress(pt.JobId, int(calculatedJobPercentage), "running")
							}
						}
					} else {
						p.publishJobSteps(pt.JobId)
					}
				}
			} else if p.CurrentMessage.JobId != "" {
				// Progress == 100 with no existing tracker: instant-complete step.
				transient := &ProgressTracker{
					JobId:         p.CurrentMessage.JobId,
					CurrentAction: p.CurrentMessage.CurrentAction,
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

					if p.OnUpdateJobProgressAndSteps != nil && int(calculatedJobPercentage) > 0 {
						dtos := p.buildStepDTOs(p.CurrentMessage.JobId)
						p.OnUpdateJobProgressAndSteps(p.CurrentMessage.JobId, int(calculatedJobPercentage), "running", dtos)
					} else {
						p.publishJobSteps(p.CurrentMessage.JobId)
						if transient.JobPercentage > 0 && p.OnUpdateJobProgress != nil {
							p.OnUpdateJobProgress(transient.JobId, int(transient.JobPercentage), "running")
						}
					}
				}
				shouldLog = true
			}
		} else {
			// If the action changed (e.g. compression tracker reused for upload), reset
			// progress so the new step starts from 0 instead of inheriting the old 100%.
			if pt.CurrentAction != p.CurrentMessage.CurrentAction && p.CurrentMessage.CurrentAction != "" {
				pt.CurrentProgress = 0
				pt.CurrentSize = 0
				pt.StartTime = time.Now()
				pt.RateSamples = make([]RateSample, 0, 60)
			}
			p.updateProgressTracker(pt, p.CurrentMessage.CurrentProgress, p.CurrentMessage.currentSize)

			pt.CurrentAction = p.CurrentMessage.CurrentAction
			pt.CurrentActionStep = p.CurrentMessage.CurrentActionStep
			pt.Filename = p.CurrentMessage.Filename

			if pt.JobId != "" {
				if workflow, hasWorkflow := p.activeWorkflows[pt.JobId]; hasWorkflow {
					autoSkipPreviousSteps(&workflow, pt.CurrentAction)
					if pt.CurrentProgress >= 100 {
						workflow.StepProgress[pt.CurrentAction] = 100.0
					} else {
						workflow.StepProgress[pt.CurrentAction] = pt.CurrentProgress
					}

					if pt.TotalSize > 0 {
						workflow.StepTotal[pt.CurrentAction] = pt.TotalSize
						workflow.StepValue[pt.CurrentAction] = pt.CurrentSize
					}
					if pt.Filename != "" {
						workflow.StepFilename[pt.CurrentAction] = pt.Filename
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
					pt.JobPercentage = calculatedJobPercentage
					p.activeWorkflows[pt.JobId] = workflow

					// Single atomic update: progress + steps in one DB write + one event broadcast.
					if p.OnUpdateJobProgressAndSteps != nil && int(calculatedJobPercentage) > 0 {
						dtos := p.buildStepDTOs(pt.JobId)
						p.OnUpdateJobProgressAndSteps(pt.JobId, int(calculatedJobPercentage), "running", dtos)
					} else {
						p.publishJobSteps(pt.JobId)
						if pt.JobPercentage > 0 && p.OnUpdateJobProgress != nil {
							p.OnUpdateJobProgress(pt.JobId, int(pt.JobPercentage), "running")
						}
					}
				} else {
					pt.JobPercentage = p.CurrentMessage.JobPercentage
				}
			} else {
				pt.JobPercentage = p.CurrentMessage.JobPercentage
			}

			shouldLog = true
		}

		shouldDelete := false
		if p.CurrentMessage.Closed() {
			shouldDelete = true
		} else if pt != nil && pt.JobId != "" && len(p.activeWorkflows[pt.JobId].Steps) > 0 {
			if pt.JobPercentage >= 100 {
				shouldDelete = true
			}
		} else if p.CurrentMessage.CurrentProgress >= 100 {
			shouldDelete = true
		}

		if shouldDelete {
			delete(p.activeProgress, p.CurrentMessage.CorrelationId())
			if pt != nil && pt.JobId != "" {
				delete(p.activeWorkflows, pt.JobId)
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
					if p.CurrentMessage.Level == JobMessageLevelError {
						workflow.StepError[p.CurrentMessage.CurrentAction] = p.CurrentMessage.Message
					} else {
						workflow.StepMessage[p.CurrentMessage.CurrentAction] = p.CurrentMessage.Message
					}
					p.activeWorkflows[p.CurrentMessage.JobId] = workflow
					p.publishJobSteps(p.CurrentMessage.JobId)
				} else {
					workflow.Message = p.CurrentMessage.Message
					p.activeWorkflows[p.CurrentMessage.JobId] = workflow
					if p.OnUpdateJobMessage != nil {
						p.OnUpdateJobMessage(p.CurrentMessage.JobId, p.CurrentMessage.Message)
					}
				}
			} else if p.CurrentMessage.CurrentAction == "" && p.OnUpdateJobMessage != nil {
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
			pt, exists := p.activeProgress[p.CurrentMessage.CorrelationId()]
			p.mu.RUnlock()
			if exists {
				baseMsg := p.CurrentMessage.Message
				if baseMsg == "" {
					baseMsg = pt.Prefix
				}
				printMsg += baseMsg + " "
				printMsg += formatProgressMessage(&p.CurrentMessage, pt)
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
			case JobMessageLevelError:
				p.ctx.LogErrorf("%s", printMsg)
			case JobMessageLevelWarning:
				p.ctx.LogWarnf("%s", printMsg)
			case JobMessageLevelDebug:
				p.ctx.LogDebugf("%s", printMsg)
			default:
				p.ctx.LogInfof("%s", printMsg)
			}
		}
	}
}

// buildStepDTOs assembles the current step snapshot for a job.
// Caller must hold p.mu (read lock is sufficient).
func (p *JobProgressService) buildStepDTOs(jobId string) []data_models.JobStep {
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
			prog = 100.0
		}

		if explicitState, ok := workflow.StepState[stepInfo.Name]; ok && explicitState != "" {
			state = explicitState
		}

		// DisplayName is the UI label; always populated, falling back to Name.
		displayName := stepInfo.DisplayName
		if displayName == "" {
			displayName = stepInfo.Name
		}

		dto := data_models.JobStep{
			Name:              stepInfo.Name,
			DisplayName:       displayName,
			Weight:            stepInfo.Weight,
			Parallel:          stepInfo.Parallel,
			HasPercentage:     stepInfo.HasPercentage,
			State:             state,
			CurrentPercentage: prog,
		}

		if total, hasTotal := workflow.StepTotal[stepInfo.Name]; hasTotal && total > 0 {
			dto.Total = total
			if val, hasVal := workflow.StepValue[stepInfo.Name]; hasVal {
				if prog >= 100 {
					dto.Value = total
				} else {
					dto.Value = val
				}
			}
			dto.Unit = "bytes"
		}
		if fname, hasFname := workflow.StepFilename[stepInfo.Name]; hasFname {
			dto.Filename = fname
		}

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

func (p *JobProgressService) publishJobSteps(jobId string) {
	if p.OnUpdateJobSteps == nil || jobId == "" {
		return
	}
	dtos := p.buildStepDTOs(jobId)
	if dtos != nil {
		p.OnUpdateJobSteps(jobId, dtos)
	}
}

func (p *JobProgressService) UpdateJobResultRecord(jobId string, recordId string, recordType string) {
	if p.OnUpdateJobResultRecord != nil {
		p.OnUpdateJobResultRecord(jobId, recordId, recordType)
	}
}

func (p *JobProgressService) Restart() {
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

func (p *JobProgressService) SkipStep(jobId string, stepName string, message string) {
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

// FailStep marks a workflow step as failed, stores the error message, recalculates
// job percentage, publishes the updated step snapshot, and logs the error to the console.
func (p *JobProgressService) FailStep(jobId string, stepName string, message string) {
	if jobId == "" || stepName == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	workflow, exists := p.activeWorkflows[jobId]
	if !exists {
		return
	}

	if workflow.StepProgress == nil {
		workflow.StepProgress = make(map[string]float64)
	}
	if workflow.StepState == nil {
		workflow.StepState = make(map[string]constants.JobState)
	}
	if workflow.StepError == nil {
		workflow.StepError = make(map[string]string)
	}

	workflow.StepProgress[stepName] = 100.0
	workflow.StepState[stepName] = constants.JobStateFailed
	if message != "" {
		workflow.StepError[stepName] = message
	} else {
		workflow.StepError[stepName] = "Step failed"
	}

	var calculatedJobPercentage float64
	for _, step := range workflow.Steps {
		if prog, ok := workflow.StepProgress[step.Name]; ok {
			calculatedJobPercentage += (prog / 100.0) * step.Weight
		}
	}
	if calculatedJobPercentage > 100 {
		calculatedJobPercentage = 100
	}

	p.activeWorkflows[jobId] = workflow

	p.ctx.LogErrorf("[job:%s] step %q failed: %s", jobId, stepName, workflow.StepError[stepName])

	if p.OnUpdateJobProgressAndSteps != nil && int(calculatedJobPercentage) > 0 {
		dtos := p.buildStepDTOs(jobId)
		p.OnUpdateJobProgressAndSteps(jobId, int(calculatedJobPercentage), "running", dtos)
	} else {
		p.publishJobSteps(jobId)
		if calculatedJobPercentage > 0 && p.OnUpdateJobProgress != nil {
			p.OnUpdateJobProgress(jobId, int(calculatedJobPercentage), "running")
		}
	}
}

// FailJob marks all pending and running steps as failed, leaving completed and skipped
// steps untouched. It then publishes a final JOB_UPDATE with the failed state.
func (p *JobProgressService) FailJob(jobId string, message string) {
	if jobId == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	workflow, exists := p.activeWorkflows[jobId]
	if !exists {
		return
	}

	if workflow.StepState == nil {
		workflow.StepState = make(map[string]constants.JobState)
	}
	if workflow.StepError == nil {
		workflow.StepError = make(map[string]string)
	}
	if workflow.StepProgress == nil {
		workflow.StepProgress = make(map[string]float64)
	}

	errMsg := message
	if errMsg == "" {
		errMsg = "Job failed"
	}

	for _, step := range workflow.Steps {
		state := workflow.StepState[step.Name]
		if state == constants.JobStateCompleted || state == constants.JobStateSkipped {
			continue
		}
		workflow.StepProgress[step.Name] = 100.0
		workflow.StepState[step.Name] = constants.JobStateFailed
		workflow.StepError[step.Name] = errMsg
	}

	p.activeWorkflows[jobId] = workflow
	p.ctx.LogErrorf("[job:%s] failed: %s", jobId, errMsg)

	if p.OnUpdateJobProgressAndSteps != nil {
		dtos := p.buildStepDTOs(jobId)
		p.OnUpdateJobProgressAndSteps(jobId, 100, "failed", dtos)
	} else {
		p.publishJobSteps(jobId)
		if p.OnUpdateJobProgress != nil {
			p.OnUpdateJobProgress(jobId, 100, "failed")
		}
	}
}

// FailJobf is a formatting helper for FailJob.
func (p *JobProgressService) FailJobf(jobId string, format string, args ...any) {
	p.FailJob(jobId, fmt.Sprintf(format, args...))
}

// StartStep immediately marks a step as running (JobStateRunning) without changing its
// percentage, logs the message to the console, and emits a full JOB_UPDATE bypassing the queue.
func (p *JobProgressService) StartStep(jobId string, stepName string, message string) {
	if jobId == "" || stepName == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	workflow, exists := p.activeWorkflows[jobId]
	if !exists {
		return
	}

	if workflow.StepState == nil {
		workflow.StepState = make(map[string]constants.JobState)
	}
	if workflow.StepMessage == nil {
		workflow.StepMessage = make(map[string]string)
	}
	workflow.StepState[stepName] = constants.JobStateRunning
	if message != "" {
		workflow.StepMessage[stepName] = message
		p.ctx.LogInfof("[job:%s] step %q started: %s", jobId, stepName, message)
	} else {
		p.ctx.LogInfof("[job:%s] step %q started", jobId, stepName)
	}
	p.activeWorkflows[jobId] = workflow

	var calculatedJobPercentage float64
	for _, step := range workflow.Steps {
		if prog, ok := workflow.StepProgress[step.Name]; ok {
			calculatedJobPercentage += (prog / 100.0) * step.Weight
		}
	}
	if calculatedJobPercentage > 100 {
		calculatedJobPercentage = 100
	}

	if p.OnUpdateJobProgressAndSteps != nil {
		dtos := p.buildStepDTOs(jobId)
		p.OnUpdateJobProgressAndSteps(jobId, int(calculatedJobPercentage), "running", dtos)
	} else {
		p.publishJobSteps(jobId)
	}
}

// StartStepf is a formatting helper for StartStep.
func (p *JobProgressService) StartStepf(jobId string, stepName string, format string, args ...any) {
	p.StartStep(jobId, stepName, fmt.Sprintf(format, args...))
}

// UpdateStepProgress updates a step's percentage and emits a full JOB_UPDATE without
// changing the step's message or state. progress should be 0-100.
func (p *JobProgressService) UpdateStepProgress(jobId string, stepName string, progress float64) {
	if jobId == "" || stepName == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	workflow, exists := p.activeWorkflows[jobId]
	if !exists {
		return
	}

	if workflow.StepProgress == nil {
		workflow.StepProgress = make(map[string]float64)
	}
	if progress > 100 {
		progress = 100
	}
	workflow.StepProgress[stepName] = progress
	p.activeWorkflows[jobId] = workflow

	var calculatedJobPercentage float64
	for _, step := range workflow.Steps {
		if prog, ok := workflow.StepProgress[step.Name]; ok {
			calculatedJobPercentage += (prog / 100.0) * step.Weight
		}
	}
	if calculatedJobPercentage > 100 {
		calculatedJobPercentage = 100
	}

	if p.OnUpdateJobProgressAndSteps != nil && int(calculatedJobPercentage) > 0 {
		dtos := p.buildStepDTOs(jobId)
		p.OnUpdateJobProgressAndSteps(jobId, int(calculatedJobPercentage), "running", dtos)
	} else {
		p.publishJobSteps(jobId)
		if calculatedJobPercentage > 0 && p.OnUpdateJobProgress != nil {
			p.OnUpdateJobProgress(jobId, int(calculatedJobPercentage), "running")
		}
	}
}

// UpdateStepMessage updates a step's message, logs it to the console, and emits a full
// JOB_UPDATE (via OnUpdateJobProgressAndSteps) without changing the step's percentage.
func (p *JobProgressService) UpdateStepMessage(jobId string, stepName string, message string) {
	if jobId == "" || stepName == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	workflow, exists := p.activeWorkflows[jobId]
	if !exists {
		return
	}

	if workflow.StepMessage == nil {
		workflow.StepMessage = make(map[string]string)
	}
	workflow.StepMessage[stepName] = message
	p.activeWorkflows[jobId] = workflow

	p.ctx.LogInfof("[job:%s] step %q: %s", jobId, stepName, message)

	var calculatedJobPercentage float64
	for _, step := range workflow.Steps {
		if prog, ok := workflow.StepProgress[step.Name]; ok {
			calculatedJobPercentage += (prog / 100.0) * step.Weight
		}
	}
	if calculatedJobPercentage > 100 {
		calculatedJobPercentage = 100
	}

	if p.OnUpdateJobProgressAndSteps != nil {
		dtos := p.buildStepDTOs(jobId)
		p.OnUpdateJobProgressAndSteps(jobId, int(calculatedJobPercentage), "running", dtos)
	} else {
		p.publishJobSteps(jobId)
	}
}

// UpdateStepMessagef is a formatting helper for UpdateStepMessage.
func (p *JobProgressService) UpdateStepMessagef(jobId string, stepName string, format string, args ...any) {
	p.UpdateStepMessage(jobId, stepName, fmt.Sprintf(format, args...))
}

// SkipStepf is a formatting helper for SkipStep.
func (p *JobProgressService) SkipStepf(jobId string, stepName string, format string, args ...any) {
	p.SkipStep(jobId, stepName, fmt.Sprintf(format, args...))
}

// CompleteStepf is a formatting helper for CompleteStep.
func (p *JobProgressService) CompleteStepf(jobId string, stepName string, format string, args ...any) {
	p.CompleteStep(jobId, stepName, fmt.Sprintf(format, args...))
}

// CompleteStepWithFile marks a step as completed and also records the filename
// associated with it (e.g. a downloaded file). Use this instead of sending a
// raw 100% progress message when you need to set the filename in the step DTO.
func (p *JobProgressService) CompleteStepWithFile(jobId string, stepName string, filename string, message string) {
	if jobId == "" || stepName == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	workflow, exists := p.activeWorkflows[jobId]
	if !exists {
		return
	}

	if workflow.StepProgress == nil {
		workflow.StepProgress = make(map[string]float64)
	}
	if workflow.StepState == nil {
		workflow.StepState = make(map[string]constants.JobState)
	}
	if workflow.StepMessage == nil {
		workflow.StepMessage = make(map[string]string)
	}
	if workflow.StepFilename == nil {
		workflow.StepFilename = make(map[string]string)
	}

	workflow.StepProgress[stepName] = 100.0
	workflow.StepState[stepName] = constants.JobStateCompleted
	if filename != "" {
		workflow.StepFilename[stepName] = filename
	}
	if message != "" {
		workflow.StepMessage[stepName] = message
	}

	var calculatedJobPercentage float64
	for _, step := range workflow.Steps {
		if prog, ok := workflow.StepProgress[step.Name]; ok {
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
		if calculatedJobPercentage > 0 && p.OnUpdateJobProgress != nil {
			p.OnUpdateJobProgress(jobId, int(calculatedJobPercentage), "running")
		}
	}
}

// FailStepf is a formatting helper for FailStep.
func (p *JobProgressService) FailStepf(jobId string, stepName string, format string, args ...any) {
	p.FailStep(jobId, stepName, fmt.Sprintf(format, args...))
}

// CompleteStep manually marks a workflow step as completed (100%, JobStateCompleted)
// and immediately publishes the updated step snapshot.
// Use this when automatic progress tracking hasn't fired for a step
// that is done (e.g. near-instant steps, or steps driven outside the tracker).
func (p *JobProgressService) CompleteStep(jobId string, stepName string, message string) {
	if jobId == "" || stepName == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	workflow, exists := p.activeWorkflows[jobId]
	if !exists {
		return
	}

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
	workflow.StepState[stepName] = constants.JobStateCompleted
	if message != "" {
		workflow.StepMessage[stepName] = message
	}

	var calculatedJobPercentage float64
	for _, step := range workflow.Steps {
		if prog, ok := workflow.StepProgress[step.Name]; ok {
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
		if calculatedJobPercentage > 0 && p.OnUpdateJobProgress != nil {
			p.OnUpdateJobProgress(jobId, int(calculatedJobPercentage), "running")
		}
	}
}

// PublishJobUpdate manually recalculates and broadcasts the current job progress
// and step snapshot. Use this to force a UI refresh without changing any step state.
func (p *JobProgressService) PublishJobUpdate(jobId string) {
	if jobId == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	workflow, exists := p.activeWorkflows[jobId]
	if !exists {
		return
	}

	var calculatedJobPercentage float64
	for _, step := range workflow.Steps {
		if prog, ok := workflow.StepProgress[step.Name]; ok {
			calculatedJobPercentage += (prog / 100.0) * step.Weight
		}
	}
	if calculatedJobPercentage > 100 {
		calculatedJobPercentage = 100
	}

	if p.OnUpdateJobProgressAndSteps != nil {
		dtos := p.buildStepDTOs(jobId)
		p.OnUpdateJobProgressAndSteps(jobId, int(calculatedJobPercentage), "running", dtos)
	} else {
		p.publishJobSteps(jobId)
		if calculatedJobPercentage > 0 && p.OnUpdateJobProgress != nil {
			p.OnUpdateJobProgress(jobId, int(calculatedJobPercentage), "running")
		}
	}
}

// CleanupProgress removes all progress tracking for a specific correlation ID.
func (p *JobProgressService) CleanupProgress(correlationId string) {
	if correlationId == "" {
		return
	}

	encodedID := normalizeCorrelationID(correlationId)
	p.cleanupProgressByEncodedID(encodedID)
	p.ctx.LogDebugf("Cleaned up job progress for correlation ID: %s", correlationId)
}

// CleanupNotifications is an alias for CleanupProgress for backward compatibility.
func (p *JobProgressService) CleanupNotifications(correlationId string) {
	p.CleanupProgress(correlationId)
}

func (p *JobProgressService) cleanupProgressByEncodedID(encodedID string) {
	if encodedID == "" {
		return
	}

	p.mu.Lock()
	delete(p.activeProgress, encodedID)
	p.mu.Unlock()

	if p.previousMessage.correlationId == encodedID {
		p.previousMessage = JobMessage{}
	}

	if p.CurrentMessage.correlationId == encodedID {
		p.CurrentMessage = JobMessage{}
	}
}

func (p *JobProgressService) GetActiveProgressCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.activeProgress)
}

func (p *JobProgressService) GetActiveProgressIDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ids := make([]string, 0, len(p.activeProgress))
	for id := range p.activeProgress {
		ids = append(ids, id)
	}
	return ids
}

func (p *JobProgressService) IsProgressActive(correlationId string) bool {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, exists := p.activeProgress[encodedID]
	return exists
}

func (p *JobProgressService) GetProgressStatus(correlationId string) (float64, bool) {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	if pt, exists := p.activeProgress[encodedID]; exists {
		return pt.CurrentProgress, true
	}
	return 0, false
}

// CleanupStaleProgress removes progress entries that haven't been updated
// for longer than the specified duration.
func (p *JobProgressService) CleanupStaleProgress(staleDuration time.Duration) {
	now := time.Now()
	var idsToCleanup []string

	p.mu.RLock()
	for id, pt := range p.activeProgress {
		if now.Sub(pt.LastUpdateTime) > staleDuration {
			idsToCleanup = append(idsToCleanup, id)
		}
	}
	p.mu.RUnlock()

	for _, id := range idsToCleanup {
		decodedID, err := decodeCorrelationID(id)
		if err != nil || decodedID == "" {
			decodedID = id
		}
		p.ctx.LogDebugf("Cleaning up stale progress for correlation ID: %s", decodedID)
		p.cleanupProgressByEncodedID(id)
	}
}

func (p *JobProgressService) GetProgressDuration(correlationId string) (time.Duration, bool) {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	if pt, exists := p.activeProgress[encodedID]; exists {
		return time.Since(pt.StartTime), true
	}
	return 0, false
}

func (p *JobProgressService) GetProgressRate(correlationId string) (*ProgressRate, bool) {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	pt, exists := p.activeProgress[encodedID]
	if !exists || pt.TotalSize <= 0 {
		return nil, false
	}

	totalDuration := time.Since(pt.StartTime).Seconds()
	if totalDuration <= 0 {
		return nil, false
	}

	rate := &ProgressRate{
		BytesPerSecond:    float64(pt.CurrentSize) / totalDuration,
		ProgressPerSecond: pt.CurrentProgress / totalDuration,
	}
	rate.RecentBytesPerSecond = rate.BytesPerSecond

	return rate, true
}

func (p *JobProgressService) PredictTimeRemaining(correlationId string) (time.Duration, bool) {
	encodedID := normalizeCorrelationID(correlationId)
	p.mu.RLock()
	defer p.mu.RUnlock()
	pt, exists := p.activeProgress[encodedID]
	if !exists || pt.TotalSize <= 0 {
		return 0, false
	}

	elapsed := time.Since(pt.StartTime)
	if elapsed <= 0 {
		return 0, false
	}

	bytesPerSecond := float64(pt.CurrentSize) / elapsed.Seconds()
	if bytesPerSecond <= 0 {
		return 0, false
	}

	remainingBytes := float64(pt.TotalSize - pt.CurrentSize)
	remainingSeconds := remainingBytes / bytesPerSecond

	if remainingSeconds < 0 {
		remainingSeconds = 0
	}

	return time.Duration(remainingSeconds * float64(time.Second)), true
}

func (p *JobProgressService) GetFormattedTimeRemaining(correlationId string) string {
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

// FormatTransferRate converts bytes per second to a human-readable string.
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

func (p *JobProgressService) GetFormattedProgressRate(correlationId string) string {
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

func formatProgressMessage(msg *JobMessage, pt *ProgressTracker) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("%.1f%%", msg.CurrentProgress))

	if msg.TotalSize() > 0 && msg.CurrentSize() > 0 {
		current := formatSize(float64(msg.CurrentSize()))
		total := formatSize(float64(msg.TotalSize()))
		parts = append(parts, fmt.Sprintf("[%s/%s]", current, total))

		elapsed := time.Since(pt.StartTime)
		if elapsed > 0 {
			bytesPerSecond := float64(pt.CurrentSize) / elapsed.Seconds()
			parts = append(parts, FormatTransferRate(bytesPerSecond))

			if remainingTime := calculateETA(pt.StartTime, pt.CurrentSize, pt.TotalSize); remainingTime != "calculating..." {
				parts = append(parts, fmt.Sprintf("ETA: %s", remainingTime))
			}
		}
	}

	return strings.Join(parts, " ")
}

func formatSize(bytes float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	unitIndex := 0

	for bytes >= 1024 && unitIndex < len(units)-1 {
		bytes /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.2f %s", bytes, units[unitIndex])
}

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
