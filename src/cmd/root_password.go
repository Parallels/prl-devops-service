package cmd

import (
	"net"
	"os"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/cjlapao/common-go/helper"
)

// serviceProbeTimeout keeps the "is the service running" check from delaying
// the command when nothing is listening.
const serviceProbeTimeout = 300 * time.Millisecond

func processRootPassword(ctx basecontext.ApiContext, cmd string) {
	ctx.LogInfof("Updating root password")
	_ = os.Setenv(constants.SOURCE_ENV_VAR, "root_password")
	processTelemetry(cmd)
	rootPassword := helper.GetFlagValue(constants.PASSWORD_FLAG, "")
	if rootPassword != "" {
		serviceprovider.InitServices(ctx)
		db := serviceprovider.Get().JsonDatabase
		if db == nil {
			panic(data.ErrDatabaseNotConnected)
		}
		ctx.LogInfof("Database connection found, updating password")
		_ = db.Connect(ctx)
		if err := db.UpdateRootPassword(ctx, rootPassword); err != nil {
			panic(err)
		}
		// CLI commands never save by default, so the update would live only in this
		// process' memory and be lost on exit. Persist it explicitly.
		if err := db.SaveNowForced(ctx, constants.UPDATE_ROOT_PASSWORD_COMMAND); err != nil {
			ctx.LogErrorf("Root password could not be written to %s: %v", db.Filename(), err)
			_ = db.Disconnect(ctx)
			os.Exit(1)
		}
		_ = db.Disconnect(ctx)
	} else {
		panic("No password provided")
	}

	warnRootPasswordCaveats(ctx)
	ctx.LogInfof("Root password updated")

	os.Exit(0)
}

// warnRootPasswordCaveats reports the two conditions that make a persisted root
// password change ineffective, so the operator is not left with a command that
// reported success while the service still rejects the new password.
func warnRootPasswordCaveats(ctx basecontext.ApiContext) {
	cfg := config.Get()

	if cfg.RootPassword() != "" {
		ctx.LogWarnf("%s is set in this service configuration and takes precedence: it overwrites the "+
			"root password on every start. Remove it to keep this password.", constants.ROOT_PASSWORD_ENV_VAR)
	}

	if address := localServiceAddress(cfg); address != "" && isTCPPortOpen(address) {
		ctx.LogWarnf("Something is serving on %s. If that is the running service, restart it now: it keeps "+
			"the previous database in memory and its periodic save overwrites this change.", address)
	}
}

// localServiceAddress returns the host:port the service listens on locally, or
// an empty string when it cannot be determined.
func localServiceAddress(cfg *config.Config) string {
	port := cfg.ApiPort()
	if cfg.TlsEnabled() {
		port = cfg.TlsPort()
	}

	if port == "" {
		return ""
	}

	return net.JoinHostPort("127.0.0.1", port)
}

func isTCPPortOpen(address string) bool {
	conn, err := net.DialTimeout("tcp", address, serviceProbeTimeout)
	if err != nil {
		return false
	}

	_ = conn.Close()
	return true
}
