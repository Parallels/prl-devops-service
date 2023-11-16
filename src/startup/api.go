package startup

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/controllers"
	"github.com/Parallels/pd-api-service/restapi"
)

var listener *restapi.HttpListener

func InitApi() *restapi.HttpListener {
	ctx := basecontext.NewRootBaseContext()
	listener = restapi.GetHttpListener()
	cfg := config.NewConfig()
	listener.Options.ApiPrefix = cfg.GetApiPrefix()
	listener.Options.HttpPort = cfg.GetApiPort()
	listener.WithVersion("Version 1", "v1", true)

	if cfg.TLSEnabled() {
		listener.Options.EnableTLS = true
		listener.Options.TLSCertificate = cfg.GetTlsCertificate()
		listener.Options.TLSPrivateKey = cfg.GetTlsPrivateKey()
		listener.Options.TLSPort = cfg.GetTLSPort()
	}

	listener.AddSwagger()
	listener.AddJsonContent().AddLogger().AddHealthCheck()
	listener.WithPublicUserRegistration()
	controllers.RegisterV1Handlers(ctx)

	return listener
}

func ResetApi() {
	listener.WaitAndShutdown()
}
