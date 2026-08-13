package startup

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/controllers"
	"github.com/Parallels/prl-devops-service/restapi"
)

var listener *restapi.HttpListener

func InitApi() *restapi.HttpListener {
	ctx := basecontext.NewRootBaseContext()
	listener = restapi.GetHttpListener()
	cfg := config.Get()
	listener.Options.ApiPrefix = cfg.ApiPrefix()
	listener.Options.HttpPort = cfg.ApiPort()
	listener.WithVersion("Version 1", "v1", true)

	if cfg.TlsEnabled() {
		listener.Options.EnableTLS = true
		listener.Options.TLSCertificate = cfg.TlsCertificate()
		listener.Options.TLSPrivateKey = cfg.TlsPrivateKey()
		listener.Options.TLSPort = cfg.TlsPort()

		// Check if HTTP should be disabled when TLS is enabled
		listener.Options.DisableHttpWhenTls = cfg.ShouldDisableHttpWhenTls()

		if listener.Options.DisableHttpWhenTls {
			common.Logger.Info("HTTPS-only mode enabled: HTTP traffic will be rejected")
		}
	}

	listener.AddSwagger()
	listener.AddJsonContent().AddLogger().AddHealthCheck()
	listener.WithPublicUserRegistration()
	_ = controllers.RegisterV1Handlers(ctx)

	return listener
}

func ResetApi() {
	listener.WaitAndShutdown()
}
