package restapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"

	_ "github.com/Parallels/prl-devops-service/docs"
	log "github.com/cjlapao/common-go-logger"
	"github.com/cjlapao/common-go/configuration"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/cjlapao/common-go/helper/reflect_helper"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func defaultCORSAllowedHeaders() []string {
	return []string{
		"X-Requested-With",
		"Accept",
		"Authorization",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Origin",
		"Access-Control-Request-Method",
		"Access-Control-Request-Headers",
		"X-Source-Id",
	}
}

func defaultCORSAllowedMethods() []string {
	return []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
}

func splitAndTrimCommaSeparated(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}

	return result
}

func mergeNormalizedValues(defaults []string, configured string, normalize func(string) string) []string {
	result := make([]string, 0, len(defaults))
	seen := make(map[string]bool)

	appendValue := func(value string) {
		normalized := normalize(strings.TrimSpace(value))
		if normalized == "" || seen[normalized] {
			return
		}

		seen[normalized] = true
		result = append(result, normalized)
	}

	for _, value := range defaults {
		appendValue(value)
	}

	for _, value := range splitAndTrimCommaSeparated(configured) {
		appendValue(value)
	}

	return result
}

func buildCORSHandler(cfg *config.Config, handler http.Handler) http.Handler {
	headers := defaultCORSAllowedHeaders()
	if configuredHeaders := cfg.GetKey(constants.CORS_ALLOWED_HEADERS_ENV_VAR); configuredHeaders != "" {
		headers = mergeNormalizedValues(defaultCORSAllowedHeaders(), configuredHeaders, http.CanonicalHeaderKey)
	}

	methods := defaultCORSAllowedMethods()
	if configuredMethods := cfg.GetKey(constants.CORS_ALLOWED_METHODS_ENV_VAR); configuredMethods != "" {
		methods = mergeNormalizedValues(defaultCORSAllowedMethods(), configuredMethods, strings.ToUpper)
	}

	origins := []string{"*"}
	if configuredOrigins := cfg.GetKey(constants.CORS_ALLOWED_ORIGINS_ENV_VAR); configuredOrigins != "" {
		trimmedOrigins := splitAndTrimCommaSeparated(configuredOrigins)
		if len(trimmedOrigins) > 0 {
			origins = trimmedOrigins
		}
	}

	return handlers.CORS(
		handlers.AllowedOrigins(origins),
		handlers.AllowedHeaders(headers),
		handlers.AllowedMethods(methods),
	)(handler)
}

type HttpControllerMethod string

const (
	GET     HttpControllerMethod = "GET"
	POST    HttpControllerMethod = "POST"
	PUT     HttpControllerMethod = "PUT"
	DELETE  HttpControllerMethod = "DELETE"
	PATCH   HttpControllerMethod = "PATCH"
	OPTIONS HttpControllerMethod = "OPTIONS"
	HEAD    HttpControllerMethod = "HEAD"
)

type HttpVersion struct {
	Version string
	Path    string
	Default bool
}

// HttpListener HttpListener structure
type HttpListener struct {
	ServerName          string
	ServerVersion       string
	Router              *mux.Router
	Logger              *log.LoggerService
	Options             *HttpListenerOptions
	Configuration       *configuration.ConfigurationService
	ControllersHandlers []ControllerHandler
	Controllers         []*Controller
	DefaultAdapters     []Adapter
	Servers             []*http.Server
	Versions            []HttpVersion
	shutdownRequest     chan bool
	shutdownRequested   uint32
	needsRestart        bool
}

var (
	globalHttpListener *HttpListener
	shutdown           chan bool
	Initialized        chan bool = make(chan bool, 1)
	needsRestart       chan bool
)

func Get() *HttpListener {
	return globalHttpListener
}

func GetRestartChannel() chan bool {
	return needsRestart
}

// NewHttpListener  Creates a new controller
func NewHttpListener() *HttpListener {
	needsRestart = make(chan bool, 10)
	shutdown = make(chan bool, 1)

	if globalHttpListener != nil {
		globalHttpListener = nil
		if len(globalHttpListener.Servers) > 0 {
			globalHttpListener.shutdownRequest <- true
		}
	}

	listener := HttpListener{
		Router:   mux.NewRouter().StrictSlash(true),
		Servers:  make([]*http.Server, 0),
		Versions: make([]HttpVersion, 0),
	}

	listener.Router.NotFoundHandler = NotFoundController()

	listener.shutdownRequest = make(chan bool)
	listener.Logger = common.Logger
	listener.Configuration = configuration.New().RegisterDefaults()

	listener.ControllersHandlers = make([]ControllerHandler, 0)
	listener.Controllers = make([]*Controller, 0)
	listener.DefaultAdapters = make([]Adapter, 0)

	listener.Options = listener.getDefaultConfiguration()

	// Appending the correlationId renewal
	listener.DefaultAdapters = append(listener.DefaultAdapters, RequestIdMiddlewareAdapter())

	globalHttpListener = &listener
	return globalHttpListener
}

func GetHttpListener() *HttpListener {
	if globalHttpListener != nil {
		return globalHttpListener
	}

	return NewHttpListener()
}

func (l *HttpListener) GetApiPrefix() string {
	if l.Options.ApiPrefix == "" {
		return ""
	}

	return http_helper.JoinUrl(l.Options.ApiPrefix)
}

func (l *HttpListener) AddHealthCheck() *HttpListener {
	l.AddHandler(l.Probe(), http_helper.JoinUrl("health", "probe"), "GET")
	return l
}

func (l *HttpListener) AddLogger() *HttpListener {
	l.DefaultAdapters = append(l.DefaultAdapters, LoggerMiddlewareAdapter(l.Options.LogHealthChecks))
	return l
}

func (l *HttpListener) AddJsonContent() *HttpListener {
	l.DefaultAdapters = append(l.DefaultAdapters, JsonContentMiddlewareAdapter())
	return l
}

func (l *HttpListener) AddDefaultHomepage() *HttpListener {
	return l
}

func (l *HttpListener) WithPublicUserRegistration() *HttpListener {
	l.Options.PublicRegistration = true
	return l
}

func (l *HttpListener) WithVersion(version string, path string, isDefault bool) {
	for i, v := range l.Versions {
		if v.Version == version {
			l.Versions[i].Path = path
			l.Versions[i].Default = isDefault
			return
		}
	}

	l.Versions = append(l.Versions, HttpVersion{
		Version: version,
		Path:    path,
		Default: isDefault,
	})
}

func (l *HttpListener) GetFullPathPrefix() string {
	defaultVersionPath := ""
	for _, v := range l.Versions {
		if v.Default {
			defaultVersionPath = v.Path
			break
		}
	}

	if defaultVersionPath == "" && len(l.Versions) > 0 {
		defaultVersionPath = l.Versions[len(l.Versions)-1].Path
	}

	return http_helper.JoinUrl(l.Options.ApiPrefix, defaultVersionPath)
}

func (l *HttpListener) AddHandler(c ControllerHandler, path string, methods ...string) {
	l.ControllersHandlers = append(l.ControllersHandlers, c)
	var subRouter *mux.Router
	if len(methods) > 0 {
		subRouter = l.Router.Methods(methods...).Subrouter()
	} else {
		subRouter = l.Router.Methods("GET").Subrouter()
	}

	adapters := make([]Adapter, 0)
	adapters = append(adapters, l.DefaultAdapters...)

	if l.GetApiPrefix() != "" && !strings.HasPrefix(path, l.Options.ApiPrefix) {
		path = http_helper.JoinUrl(l.GetApiPrefix(), path)
	}
	subRouter.HandleFunc(path, Adapt(
		http.HandlerFunc(c),
		adapters...).ServeHTTP)
}

func (l *HttpListener) AddAuthorizedHandler(c ControllerHandler, path string, methods ...string) {
	l.AddAuthorizedHandlerWithRolesAndClaims(
		c,
		path,
		[]string{},
		[]string{},
		ComparisonOperationAnd,
		ComparisonOperationAnd,
		methods...)
}

func (l *HttpListener) AddAuthorizedHandlerWithRoles(c ControllerHandler, path string, roles []string, methods ...string) {
	l.AddAuthorizedHandlerWithRolesAndClaims(
		c,
		path,
		roles,
		[]string{},
		ComparisonOperationAnd,
		ComparisonOperationAnd,
		methods...)
}

func (l *HttpListener) AddAuthorizedHandlerWithClaims(c ControllerHandler, path string, claims []string, methods ...string) {
	l.AddAuthorizedHandlerWithRolesAndClaims(
		c,
		path,
		[]string{},
		claims,
		ComparisonOperationAnd,
		ComparisonOperationAnd,
		methods...)
}

// AddAuthorizedHandlerWithExtraAdapters is like AddAuthorizedHandlerWithRolesAndClaims but
// inserts extraAdapters into the chain just before EndAuthorizationMiddlewareAdapter.
// Use this to add per-endpoint auth extensions (e.g. enrollment-token support).
func (l *HttpListener) AddAuthorizedHandlerWithExtraAdapters(
	c ControllerHandler,
	path string,
	roles []string,
	claims []string,
	roleComparisonOperation ComparisonOperation,
	claimComparisonOperation ComparisonOperation,
	extraAdapters []Adapter,
	methods ...string) {
	l.ControllersHandlers = append(l.ControllersHandlers, c)
	var subRouter *mux.Router
	if len(methods) > 0 {
		subRouter = l.Router.Methods(methods...).Subrouter()
	} else {
		subRouter = l.Router.Methods("GET").Subrouter()
	}
	adapters := make([]Adapter, 0)
	adapters = append(adapters, l.DefaultAdapters...)
	adapters = append(adapters,
		AddAuthorizationContextMiddlewareAdapter(),
		TokenAuthorizationMiddlewareAdapter(roles, claims, roleComparisonOperation, claimComparisonOperation),
		ApiKeyAuthorizationMiddlewareAdapter(roles, claims, roleComparisonOperation, claimComparisonOperation))
	adapters = append(adapters, extraAdapters...)
	adapters = append(adapters, EndAuthorizationMiddlewareAdapter())

	if l.GetApiPrefix() != "" && !strings.HasPrefix(path, l.Options.ApiPrefix) {
		path = http_helper.JoinUrl(l.GetApiPrefix(), path)
	}

	subRouter.HandleFunc(path,
		Adapt(
			http.HandlerFunc(c),
			adapters...).ServeHTTP)
}

func (l *HttpListener) AddAuthorizedHandlerWithRolesAndClaims(
	c ControllerHandler,
	path string,
	roles []string,
	claims []string,
	roleComparisonOperation ComparisonOperation,
	claimComparisonOperation ComparisonOperation,
	methods ...string) {
	l.ControllersHandlers = append(l.ControllersHandlers, c)
	var subRouter *mux.Router
	if len(methods) > 0 {
		subRouter = l.Router.Methods(methods...).Subrouter()
	} else {
		subRouter = l.Router.Methods("GET").Subrouter()
	}
	adapters := make([]Adapter, 0)
	adapters = append(adapters, l.DefaultAdapters...)
	adapters = append(adapters,
		AddAuthorizationContextMiddlewareAdapter(),
		TokenAuthorizationMiddlewareAdapter(roles, claims, roleComparisonOperation, claimComparisonOperation),
		ApiKeyAuthorizationMiddlewareAdapter(roles, claims, roleComparisonOperation, claimComparisonOperation),
		EndAuthorizationMiddlewareAdapter())

	if l.GetApiPrefix() != "" && !strings.HasPrefix(path, l.Options.ApiPrefix) {
		path = http_helper.JoinUrl(l.Options.ApiPrefix, path)
	}

	subRouter.HandleFunc(path,
		Adapt(
			http.HandlerFunc(c),
			adapters...).ServeHTTP)
}

func (l *HttpListener) Start(serviceName string, serviceVersion string) {
	l.ServerName = serviceName
	l.ServerVersion = serviceVersion
	config := config.Get()

	l.Logger.Notice("Starting %v Go Rest API %v", serviceName, serviceVersion)

	done := make(chan bool)

	l.Router.HandleFunc(l.GetApiPrefix()+"/", defaultHomepageController)
	l.Router.HandleFunc(l.GetApiPrefix()+"/shutdown", globalHttpListener.ShutdownHandler)

	// Creating and starting the http server
	var srv *http.Server

	if config.IsCorsEnabled() {
		l.Logger.Info("Enabling CORS for HTTP")
		srv = &http.Server{
			Addr:              ":" + l.Options.HttpPort,
			Handler:           buildCORSHandler(config, l.Router),
			ReadHeaderTimeout: time.Duration(30) * time.Second,
			ReadTimeout:       time.Duration(5) * time.Hour,
			WriteTimeout:      time.Duration(5) * time.Hour,
			IdleTimeout:       time.Duration(60) * time.Second,
		}
	} else {
		srv = &http.Server{
			Addr:              ":" + l.Options.HttpPort,
			Handler:           l.Router,
			ReadHeaderTimeout: time.Duration(30) * time.Second,
			ReadTimeout:       time.Duration(5) * time.Hour,
			WriteTimeout:      time.Duration(5) * time.Hour,
			IdleTimeout:       time.Duration(60) * time.Second,
		}
	}

	l.Servers = append(l.Servers, srv)

	for _, controller := range l.Controllers {
		_ = controller.Serve()
	}

	go func() {
		l.Logger.Info("Api listening on http://::" + l.Options.HttpPort + l.GetApiPrefix())
		l.Logger.Success("Finished Initiating http server")
		if err := srv.ListenAndServe(); err != nil {
			if !strings.Contains(err.Error(), "http: Server closed") {
				l.Logger.Error("There was an error shutting down the http server: %v", err.Error())
			}
		}
		done <- true
	}()

	if l.Options.EnableTLS {
		cert, err := tls.X509KeyPair([]byte(l.Options.TLSCertificate), []byte(l.Options.TLSPrivateKey))
		if err == nil {
			tlsConfig := &tls.Config{
				Certificates:       []tls.Certificate{cert},
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true,
			}

			var sslSrv *http.Server

			if config.IsCorsEnabled() {
				l.Logger.Info("Enabling CORS for HTTPS")
				sslSrv = &http.Server{
					Addr:              ":" + l.Options.TLSPort,
					TLSConfig:         tlsConfig,
					Handler:           buildCORSHandler(config, l.Router),
					ReadHeaderTimeout: time.Duration(30) * time.Second,
					ReadTimeout:       time.Duration(5) * time.Hour,
					WriteTimeout:      time.Duration(5) * time.Hour,
					IdleTimeout:       time.Duration(60) * time.Second,
				}
			} else {
				sslSrv = &http.Server{
					Addr:              ":" + l.Options.TLSPort,
					TLSConfig:         tlsConfig,
					Handler:           l.Router,
					ReadHeaderTimeout: time.Duration(30) * time.Second,
					ReadTimeout:       time.Duration(5) * time.Hour,
					WriteTimeout:      time.Duration(5) * time.Hour,
					IdleTimeout:       time.Duration(60) * time.Second,
				}
			}

			l.Servers = append(l.Servers, sslSrv)

			go func() {
				l.Logger.Info("Api listening on https://::" + l.Options.TLSPort + l.GetApiPrefix())
				l.Logger.Success("Finished Initiating https server")
				if err := sslSrv.ListenAndServeTLS("", ""); err != nil {
					if !strings.Contains(err.Error(), "http: Server closed") {
						l.Logger.Error("There was an error shutting down the https server: %v", err.Error())
					}
				}
				done <- true
			}()
		} else {
			l.Logger.Error("There was an error reading the certificates to enable HTTPS")
		}
	}

	Initialized <- true
	l.WaitAndShutdown()
	<-done

	l.Logger.Info("Server shut down successfully...")
	shutdown <- true
	if !l.needsRestart {
		needsRestart <- false
	}
}

func (l *HttpListener) WaitAndShutdown() {
	irqSign := make(chan os.Signal, 1)
	signal.Notify(irqSign, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-irqSign:
		l.Logger.Info("Server shutdown requested (signal: %v)", sig.String())
	case sig := <-l.shutdownRequest:
		l.Logger.Info("Server shutdown requested (/shutdown: %v)", fmt.Sprintf("%v", sig))
	}

	l.Logger.Info("Stopping the server...")

	// Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create shutdown context with 10 second timeout
	for _, s := range l.Servers {
		err := s.Shutdown(ctx)
		if err != nil {
			l.Logger.Error("Shutdown request error: %v", err.Error())
		}
	}
}

func (l *HttpListener) Restart() {
	l.needsRestart = true
	l.Logger.Info("Restarting the server...")
	pid := os.Getpid()
	p, err := os.FindProcess(pid)
	if err != nil {
		l.Logger.Error("Error finding process: %v", err.Error())
	}
	err = p.Signal(syscall.SIGTERM)
	if err != nil {
		l.Logger.Error("Error sending signal: %v", err.Error())
	}

	<-shutdown
	globalHttpListener = nil
	needsRestart <- true
	l.Logger.Info("Server restarted successfully...")
}

// region Private Methods
func (l *HttpListener) getDefaultConfiguration() *HttpListenerOptions {
	options := HttpListenerOptions{
		HttpPort:        l.Configuration.GetString(constants.API_PORT_ENV_VAR),
		EnableTLS:       l.Configuration.GetBool(constants.TLS_ENABLED_ENV_VAR),
		TLSPort:         l.Configuration.GetString(constants.TLS_PORT_ENV_VAR),
		TLSCertificate:  l.Configuration.GetBase64(constants.TLS_CERTIFICATE_ENV_VAR),
		TLSPrivateKey:   l.Configuration.GetBase64(constants.TLS_PRIVATE_KEY_ENV_VAR),
		LogHealthChecks: false,
	}

	if reflect_helper.IsNilOrEmpty(options.HttpPort) {
		options.HttpPort = "5000"
	}

	if reflect_helper.IsNilOrEmpty(options.TLSPort) {
		options.TLSPort = "5001"
	}

	if reflect_helper.IsNilOrEmpty(options.DatabaseName) {
		options.DatabaseName = "users"
	}

	apiPrefix := http_helper.JoinUrl(l.Configuration.GetString("API_PREFIX"))

	options.ApiPrefix = apiPrefix

	l.Options = &options

	return l.Options
}

func defaultHomepageController(w http.ResponseWriter, r *http.Request) {
	response := DefaultHomepage{
		Timestamp: fmt.Sprint(time.Now().Format(time.RFC850)),
	}

	_ = json.NewEncoder(w).Encode(response)
}

func (l *HttpListener) AddSwagger() *HttpListener {
	l.Logger.Info("Adding Swagger support")

	l.Router.HandleFunc("/swagger", SwaggerHandler(
		SwaggerDeepLinking(true),
		SwaggerDocExpansion("none"),
		SwaggerPersistAuthorization(true),
		SwaggerDomID("swagger-ui"),
	))

	l.Router.HandleFunc("/swagger/{.*}", SwaggerHandler(
		SwaggerDeepLinking(true),
		SwaggerDocExpansion("none"),
		SwaggerPersistAuthorization(true),
		SwaggerDomID("swagger-ui"),
	))

	return l
}

// endregion
