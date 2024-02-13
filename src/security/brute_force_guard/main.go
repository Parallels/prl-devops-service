package bruteforceguard

import (
	"strconv"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/serviceprovider"
)

var globalBruteForceGuard *BruteForceGuard

type BruteForceGuard struct {
	ctx     basecontext.ApiContext
	options *BruteForceGuardOptions
}

func New(ctx basecontext.ApiContext) *BruteForceGuard {
	globalBruteForceGuard = &BruteForceGuard{
		ctx:     ctx,
		options: NewDefaultOptions(),
	}

	globalBruteForceGuard.processEnvironmentVariables()
	return globalBruteForceGuard
}

func Get() *BruteForceGuard {
	if globalBruteForceGuard == nil {
		ctx := basecontext.NewRootBaseContext()
		return New(ctx)
	}

	return globalBruteForceGuard
}

func (s *BruteForceGuard) WithMaxLoginAttempts(maxAttempts int) *BruteForceGuard {
	s.options.WithMaxLoginAttempts(maxAttempts)
	return s
}

func (s *BruteForceGuard) WithBlockDuration(duration string) *BruteForceGuard {
	s.options.WithBlockDuration(duration)
	return s
}

func (s *BruteForceGuard) WithIncrementalWait(incremental bool) *BruteForceGuard {
	s.options.WithIncrementalWait(incremental)
	return s
}

func (s *BruteForceGuard) Options() *BruteForceGuardOptions {
	return s.options
}

func (s *BruteForceGuard) Process(userId string, loginState bool, reason string) *errors.Diagnostics {
	diag := errors.NewDiagnostics()
	dbService, err := serviceprovider.GetDatabaseService(s.ctx)
	if err != nil {
		diag.AddError(err)
		return diag
	}

	user, err := dbService.GetUser(s.ctx, userId)
	if err != nil {
		diag.AddError(err)
		return diag
	}

	if user == nil {
		diag.AddError(errors.ErrNotFound())
		return diag
	}

	if loginState {
		user.FailedLoginAttempts = 0
		user.BlockedSince = ""
		user.Blocked = false
		user.BlockedReason = ""
		err := dbService.UpdateUserBlockStatus(s.ctx, *user)
		if err != nil {
			diag.AddError(err)
		}
		return diag
	} else {
		user.FailedLoginAttempts++

		if user.FailedLoginAttempts >= s.options.MaxLoginAttempts() {
			user.Blocked = true
			user.BlockedSince = time.Now().Format(time.RFC3339)
			user.BlockedReason = reason
			err := dbService.UpdateUserBlockStatus(s.ctx, *user)
			if err != nil {
				diag.AddError(err)
				return diag
			}
			if s.options.IncrementalWait() {
				countExtraAttempts := user.FailedLoginAttempts - (s.options.MaxLoginAttempts() - 1)
				sleepFor := time.Duration(s.options.BlockDuration().Seconds()*float64(countExtraAttempts)) * time.Second
				time.Sleep(sleepFor)
			} else {
				sleepFor := s.options.BlockDuration()
				time.Sleep(sleepFor)
			}
		} else {
			err := dbService.UpdateUserBlockStatus(s.ctx, *user)
			if err != nil {
				diag.AddError(err)
				return diag
			}
		}
	}

	return diag
}

func (s *BruteForceGuard) IsBlocked(userId string) bool {
	dbService, err := serviceprovider.GetDatabaseService(s.ctx)
	if err != nil {
		return false
	}

	user, err := dbService.GetUser(s.ctx, userId)
	if err != nil {
		return false
	}

	if user == nil {
		return false
	}

	return user.Blocked
}

func (s *BruteForceGuard) processEnvironmentVariables() {
	cfg := config.Get()
	if cfg.GetKey(constants.BRUTE_FORCE_MAX_LOGIN_ATTEMPTS_ENV_VAR) != "" {
		maxLoginAttempts, err := strconv.Atoi(cfg.GetKey(constants.BRUTE_FORCE_MAX_LOGIN_ATTEMPTS_ENV_VAR))
		if err != nil {
			s.ctx.LogWarnf("[BruteForceGuard] Invalid value for %s: %s", constants.BRUTE_FORCE_MAX_LOGIN_ATTEMPTS_ENV_VAR, err.Error())
		} else {
			s.ctx.LogDebugf("[BruteForceGuard] Setting %s to %d", constants.BRUTE_FORCE_MAX_LOGIN_ATTEMPTS_ENV_VAR, maxLoginAttempts)
			s.options.WithMaxLoginAttempts(maxLoginAttempts)
		}
	}

	if cfg.GetKey(constants.BRUTE_FORCE_LOCKOUT_DURATION_ENV_VAR) != "" {
		s.ctx.LogDebugf("[BruteForceGuard] Setting %s to %s", constants.BRUTE_FORCE_LOCKOUT_DURATION_ENV_VAR, cfg.GetKey(constants.BRUTE_FORCE_LOCKOUT_DURATION_ENV_VAR))
		s.options.WithBlockDuration(cfg.GetKey(constants.BRUTE_FORCE_LOCKOUT_DURATION_ENV_VAR))
	}

	if cfg.GetKey(constants.BRUTE_FORCE_INCREMENTAL_WAIT_ENV_VAR) != "" {
		incrementalWait, err := strconv.ParseBool(cfg.GetKey(constants.BRUTE_FORCE_INCREMENTAL_WAIT_ENV_VAR))
		if err != nil {
			s.ctx.LogWarnf("[BruteForceGuard] Invalid value for %s: %s", constants.BRUTE_FORCE_INCREMENTAL_WAIT_ENV_VAR, err.Error())
		} else {
			s.ctx.LogInfof("[BruteForceGuard] Setting %s to %v", constants.BRUTE_FORCE_INCREMENTAL_WAIT_ENV_VAR, incrementalWait)
			s.options.WithIncrementalWait(incrementalWait)
		}
	}
}
