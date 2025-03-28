package reverse_proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/reverse_proxy/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"golang.org/x/sync/errgroup"
)

var globalReverseProxyService *ReverseProxyService

type reverseProxyOperationRequest struct {
	operation func() error
	result    chan error
}

type reverseProxyServiceState int

const (
	ReverseProxyServiceStateStopped reverseProxyServiceState = iota
	ReverseProxyServiceStateStarting
	ReverseProxyServiceStateStarted
	ReverseProxyServiceStateStopping

	defaultOperationTimeout = 10 * time.Second
	defaultShutdownTimeout  = 5 * time.Second
	defaultReadTimeout      = 5 * time.Second
	defaultWriteTimeout     = 60 * time.Second
	defaultIdleTimeout      = 5 * time.Minute
)

var (
	// Use environment variables or defaults
	operationTimeout = getEnvDuration("REVERSE_PROXY_OPERATION_TIMEOUT", defaultOperationTimeout)
	shutdownTimeout  = getEnvDuration("REVERSE_PROXY_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	readTimeout      = getEnvDuration("REVERSE_PROXY_READ_TIMEOUT", defaultReadTimeout)
	writeTimeout     = getEnvDuration("REVERSE_PROXY_WRITE_TIMEOUT", defaultWriteTimeout)
	idleTimeout      = getEnvDuration("REVERSE_PROXY_IDLE_TIMEOUT", defaultIdleTimeout)
)

// Helper function to get duration from environment variable
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

type ReverseProxyService struct {
	enabled           bool
	host              string
	port              string
	State             reverseProxyServiceState
	forwarding_hosts  []*data_models.ReverseProxyHost
	db                *data.JsonDatabase
	api_ctx           basecontext.ApiContext
	ctx               context.Context
	cancelFunc        context.CancelFunc
	tcpListeners      []net.Listener
	httpListeners     []*http.Server
	wg                *sync.WaitGroup
	activeConnections sync.WaitGroup

	opQueue   chan reverseProxyOperationRequest
	queueOnce sync.Once
}

func Get(ctx basecontext.ApiContext) *ReverseProxyService {
	if globalReverseProxyService == nil {
		globalReverseProxyService = New(ctx)
	}

	return globalReverseProxyService
}

func GetConfig() models.ReverseProxyConfig {
	cfg := config.Get()
	result := models.ReverseProxyConfig{
		Enabled: false,
	}

	if cfg == nil || globalReverseProxyService == nil {
		result.Enabled = false
		return result
	}

	result.Host = cfg.ReverseProxyHost()
	result.Port = cfg.ReverseProxyPort()
	result.Enabled = cfg.IsReverseProxyEnabled()

	// Returning db data if available and not different from the config
	dtoRp, err := globalReverseProxyService.db.GetReverseProxyConfig(globalReverseProxyService.api_ctx)
	if dtoRp != nil && err == nil {
		if !dtoRp.Diff(mappers.ConfigReverseProxyToDto(result)) {
			result.Host = dtoRp.Host
			result.Port = dtoRp.Port
			result.Enabled = dtoRp.Enabled
			result.ID = dtoRp.ID
		}

		return result
	}

	return result
}

func New(ctx basecontext.ApiContext) *ReverseProxyService {
	db, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("Error getting database service: %s", err)
		return nil
	}

	globalReverseProxyService = &ReverseProxyService{
		api_ctx:       ctx,
		db:            db,
		tcpListeners:  []net.Listener{},
		httpListeners: []*http.Server{},
		wg:            &sync.WaitGroup{},
		State:         ReverseProxyServiceStateStopped,
	}

	return globalReverseProxyService
}

func (rps *ReverseProxyService) ImportFromConfig() error {
	// Check if the config is nil, if it is no action is going to be taken
	cfg := config.Get()
	rpsCfg := cfg.GetReverseProxyConfig()
	if rpsCfg == nil {
		return nil
	}
	dtoRp, _ := rps.db.GetReverseProxyConfig(rps.api_ctx)
	if dtoRp == nil {
		dto := data_models.ReverseProxy{
			ID:   helpers.GenerateId(),
			Host: rpsCfg.Host,
			Port: rpsCfg.Port,
		}
		if _, err := rps.db.UpdateReverseProxy(rps.api_ctx, dto); err != nil {
			return err
		}

		dtoRp = &dto
		rps.api_ctx.LogDebugf("Reverse proxy configuration imported: %v", dtoRp)
	}

	for _, host := range rpsCfg.Hosts {
		dtoHost, _ := rps.db.GetReverseProxyHost(rps.api_ctx, host.GetHost())
		if dtoHost != nil {
			configDto := mappers.ConfigReverseProxyHostToDto(*host)
			if dtoHost.Diff(configDto) {
				r, err := rps.db.UpdateReverseProxyHost(rps.api_ctx, &configDto)
				if err != nil {
					return err
				}
				rps.api_ctx.LogDebugf("Updated reverse proxy host %s", r.GetHost())
			}
			continue
		}

		dto := mappers.ConfigReverseProxyHostToDto(*host)
		r, err := rps.db.CreateReverseProxyHost(rps.api_ctx, dto)
		if err != nil {
			return err
		}
		rps.api_ctx.LogDebugf("Created reverse proxy host %s", r.GetHost())
	}

	if err := cfg.CleanReverseProxyDataFromConfig(); err != nil {
		return err
	}

	cfg.Save()
	return nil
}

func (rps *ReverseProxyService) CheckChangesInConfiguration() error {
	cfg := config.Get()
	dtoRp, _ := rps.db.GetReverseProxyConfig(rps.api_ctx)
	if dtoRp == nil && cfg.IsReverseProxyEnabled() {
		dto := data_models.ReverseProxy{
			ID:      helpers.GenerateId(),
			Enabled: cfg.IsReverseProxyEnabled(),
			Host:    cfg.ReverseProxyHost(),
			Port:    cfg.ReverseProxyPort(),
		}
		if _, err := rps.db.UpdateReverseProxy(rps.api_ctx, dto); err != nil {
			return err
		}

		dtoRp = &dto
		rps.host = dto.Host
		rps.port = dto.Port
		rps.api_ctx.LogDebugf("Reverse proxy configuration imported: %v", dtoRp)
	} else {
		configRp := GetConfig()
		if dtoRp.Diff(mappers.ConfigReverseProxyToDto(configRp)) {
			dto := mappers.ConfigReverseProxyToDto(GetConfig())
			if _, err := rps.db.UpdateReverseProxy(rps.api_ctx, dto); err != nil {
				return err
			}
		}
		rps.host = configRp.Host
		rps.port = configRp.Port
	}

	return nil
}

func (rps *ReverseProxyService) LoadFromDb() error {
	dtoRp, err := rps.db.GetReverseProxyConfig(rps.api_ctx)
	if err != nil {
		return err
	}
	if dtoRp == nil {
		return nil
	}

	rps.host = dtoRp.Host
	rps.port = dtoRp.Port
	hosts, err := rps.db.GetReverseProxyHosts(rps.api_ctx, "")
	if err != nil {
		return err
	}

	prl_svc := serviceprovider.Get().ParallelsDesktopService
	for _, host := range hosts {
		hostCopy := host
		if len(hostCopy.HttpRoutes) > 0 {
			for i, route := range hostCopy.HttpRoutes {
				if route.TargetVmId != "" {
					vm, err := prl_svc.GetVm(rps.api_ctx, route.TargetVmId)
					if err != nil || vm.InternalIpAddress == "" || vm.InternalIpAddress == "-" || vm.State != "running" {
						e := ""
						if err != nil {
							e = err.Error()
						}
						if vm == nil {
							e = "vm could not be found"
						}
						if vm != nil && vm.InternalIpAddress == "" {
							e = "vm internal ip address is empty"
						}
						if vm != nil && vm.InternalIpAddress == "-" {
							e = "vm internal ip address is not assigned"
						}
						if vm != nil && vm.State != "running" {
							e = "vm is not running"
						}
						rps.api_ctx.LogErrorf("Error getting vm %s for reverse proxy route: %s", route.TargetVmId, e)
						hostCopy.HttpRoutes[i].TargetHost = "---"
					} else {
						hostCopy.HttpRoutes[i].TargetHost = vm.InternalIpAddress
					}
				}
			}
		}
		if hostCopy.TcpRoute != nil && hostCopy.TcpRoute.TargetVmId != "" {
			vm, err := prl_svc.GetVm(rps.api_ctx, hostCopy.TcpRoute.TargetVmId)
			if err != nil || vm.InternalIpAddress == "" || vm.InternalIpAddress == "-" || vm.State != "running" {
				if err != nil {
					rps.api_ctx.LogErrorf("Error getting vm %s for reverse proxy tcp route: %s", hostCopy.TcpRoute.TargetVmId, err)
				} else if vm == nil {
					rps.api_ctx.LogErrorf("Error getting vm %s for reverse proxy tcp route: vm could not be found", hostCopy.TcpRoute.TargetVmId)
				} else if vm.InternalIpAddress == "" || vm.InternalIpAddress == "-" {
					rps.api_ctx.LogErrorf("Error getting vm %s for reverse proxy tcp route: vm internal ip address is empty", hostCopy.TcpRoute.TargetVmId)
				}
				hostCopy.TcpRoute.TargetHost = "---"
			} else {
				hostCopy.TcpRoute.TargetHost = vm.InternalIpAddress
			}
		}

		rps.forwarding_hosts = append(rps.forwarding_hosts, &hostCopy)
	}

	return nil
}

func (rps *ReverseProxyService) Start() error {
	rps.initQueue()
	resultChan := make(chan error)
	rps.opQueue <- reverseProxyOperationRequest{
		operation: rps.startInternal,
		result:    resultChan,
	}
	return <-resultChan
}

func (rps *ReverseProxyService) startInternal() error {
	if rps.State == ReverseProxyServiceStateStarted || rps.State == ReverseProxyServiceStateStarting {
		rps.api_ctx.LogInfof("[Reverse Proxy] Reverse proxy service already started")
		return nil
	}

	rps.State = ReverseProxyServiceStateStarting
	cfg := config.Get()
	rps.ctx, rps.cancelFunc = context.WithCancel(context.Background())
	errorChan := make(chan error, 1)

	// Re-initialize variables
	rps.forwarding_hosts = make([]*data_models.ReverseProxyHost, 0)
	rps.tcpListeners = make([]net.Listener, 0)
	rps.httpListeners = make([]*http.Server, 0)
	rps.wg = &sync.WaitGroup{}

	if err := rps.ImportFromConfig(); err != nil {
		rps.api_ctx.LogErrorf("Error importing reverse proxy config: %s", err)
	}
	if err := rps.LoadFromDb(); err != nil {
		rps.api_ctx.LogErrorf("Error loading reverse proxy config from db: %s", err)
		return err
	}
	if err := rps.CheckChangesInConfiguration(); err != nil {
		rps.api_ctx.LogErrorf("Error checking changes in reverse proxy config: %s", err)
		return err
	}

	if rps.port == "" {
		rps.port = cfg.ReverseProxyPort()
		rps.api_ctx.LogWarnf("[Reverse Proxy] Port not set for reverse proxy, using default port", rps.port)
	}
	if rps.host == "" {
		rps.host = cfg.ReverseProxyHost()
		rps.api_ctx.LogWarnf("[Reverse Proxy] Host not set for reverse proxy, using default host", rps.host)
	}

	rps.api_ctx.LogInfof("[Reverse Proxy] Starting reverse proxy on %s:%s", rps.host, rps.port)
	go rps.startServer(errorChan)

	select {
	case err := <-errorChan:
		if err != nil {
			rps.api_ctx.LogErrorf("[Reverse Proxy] Error starting reverse proxy: %s", err)
		}
		return err
	case <-rps.ctx.Done():
		rps.api_ctx.LogInfof("[Reverse Proxy] Stopping reverse proxy due to context cancellation")
		return nil
	default:
	}

	rps.api_ctx.LogInfof("[Reverse Proxy] Reverse proxy started")
	rps.State = ReverseProxyServiceStateStarted
	return nil
}

func (rps *ReverseProxyService) Stop() error {
	rps.api_ctx.LogInfof("[Reverse Proxy] Stopping reverse proxy service...")
	rps.State = ReverseProxyServiceStateStopping

	// Cancel the main context to stop new connections
	if rps.cancelFunc != nil {
		rps.cancelFunc()
	}

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Create a channel to signal when connections are done
	done := make(chan struct{})
	go func() {
		rps.activeConnections.Wait()
		close(done)
	}()

	// Wait for either connections to finish or timeout
	select {
	case <-done:
		rps.api_ctx.LogInfof("[Reverse Proxy] All connections closed gracefully")
	case <-shutdownCtx.Done():
		rps.api_ctx.LogWarnf("[Reverse Proxy] Timeout waiting for connections to close")
	}

	// Create error group for parallel shutdown
	eg, ctx := errgroup.WithContext(shutdownCtx)

	// Shutdown HTTP servers in parallel
	for _, server := range rps.httpListeners {
		srv := server // Create local variable for closure
		eg.Go(func() error {
			if err := srv.Shutdown(ctx); err != nil {
				rps.api_ctx.LogWarnf("[Reverse Proxy] Error during graceful shutdown: %v", err)
				if err := srv.Close(); err != nil {
					rps.api_ctx.LogErrorf("[Reverse Proxy] Error force closing server: %v", err)
					return err
				}
			}
			return nil
		})
	}

	// Close TCP listeners in parallel
	for _, listener := range rps.tcpListeners {
		l := listener // Create local variable for closure
		eg.Go(func() error {
			if err := l.Close(); err != nil {
				rps.api_ctx.LogErrorf("[Reverse Proxy] Error closing TCP listener: %v", err)
				return err
			}
			return nil
		})
	}

	// Wait for all shutdowns to complete
	if err := eg.Wait(); err != nil {
		rps.api_ctx.LogErrorf("[Reverse Proxy] Error during shutdown: %v", err)
	}

	// Clear the listeners
	rps.httpListeners = nil
	rps.tcpListeners = nil

	rps.State = ReverseProxyServiceStateStopped
	rps.api_ctx.LogInfof("[Reverse Proxy] Service stopped")
	return nil
}

func (rps *ReverseProxyService) Restart() error {
	rps.initQueue()
	resultChan := make(chan error)
	rps.opQueue <- reverseProxyOperationRequest{
		operation: rps.restartInternal,
		result:    resultChan,
	}
	return <-resultChan
}

func (rps *ReverseProxyService) restartInternal() error {
	// Save current state in case we need to rollback
	previousState := rps.State
	previousHosts := rps.forwarding_hosts

	rps.api_ctx.LogInfof("[Reverse Proxy] Restarting reverse proxy service...")

	// First stop the service
	if err := rps.Stop(); err != nil {
		rps.api_ctx.LogErrorf("[Reverse Proxy] Failed to stop service during restart: %v", err)
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Small delay to ensure all connections are properly closed
	time.Sleep(100 * time.Millisecond)

	// Reload configuration from DB
	if err := rps.LoadFromDb(); err != nil {
		rps.api_ctx.LogErrorf("[Reverse Proxy] Failed to reload configuration: %v", err)
		// Try to restore previous state
		rps.State = previousState
		rps.forwarding_hosts = previousHosts
		if err := rps.startInternal(); err != nil {
			rps.api_ctx.LogErrorf("[Reverse Proxy] Failed to restore previous state: %v", err)
		}
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	// Start the service again
	if err := rps.startInternal(); err != nil {
		rps.api_ctx.LogErrorf("[Reverse Proxy] Failed to start service during restart: %v", err)
		// Try to restore previous state
		rps.State = previousState
		rps.forwarding_hosts = previousHosts
		if startErr := rps.startInternal(); startErr != nil {
			rps.api_ctx.LogErrorf("[Reverse Proxy] Failed to restore previous state: %v", startErr)
		}
		return fmt.Errorf("failed to start service: %w", err)
	}

	rps.api_ctx.LogInfof("[Reverse Proxy] Service successfully restarted")
	return nil
}

func (rps *ReverseProxyService) startServer(errorChan chan error) {
	// Checking if we have a tcp route to deal with
	for _, host := range rps.forwarding_hosts {
		h := host
		if h.TcpRoute != nil {
			if h.TcpRoute.TargetHost == "" || h.TcpRoute.TargetPort == "---" {
				rps.api_ctx.LogErrorf("[TCP Route] target host is required for starting a tcp route, skipping host %s", h.GetHost())
				continue
			}

			rps.wg.Add(1)
			go func(h *data_models.ReverseProxyHost) {
				defer rps.wg.Done()
				if err := rps.listenTcpRoute(h, errorChan); err != nil {
					errorChan <- err
				}
			}(h)
		} else {
			rps.wg.Add(1)
			go func(h *data_models.ReverseProxyHost) {
				defer rps.wg.Done()
				if err := rps.listenHttpRoute(h, errorChan); err != nil {
					errorChan <- err
				}
			}(h)
		}
	}
}

func (rps *ReverseProxyService) listenTcpRoute(host *data_models.ReverseProxyHost, errorChan chan error) error {
	if host.TcpRoute.TargetPort == "" {
		return fmt.Errorf("[TCP Route] port is required for starting a tcp route")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host.Host, host.Port))
	if err != nil {
		errorChan <- err
		return err
	}

	rps.tcpListeners = append(rps.tcpListeners, listener)
	defer listener.Close()

	rps.api_ctx.LogInfof("[Reverse Proxy] [TCP Route] Listening on %s:%s", host.Host, host.Port)
	for {
		select {
		case <-rps.ctx.Done():
			rps.api_ctx.LogDebugf("[Reverse Proxy] [TCP Route] Shutting down listener for %s:%s", host.Host, host.Port)
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-rps.ctx.Done():
				return nil
			default:
				if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
					rps.api_ctx.LogDebugf("[Reverse Proxy] [TCP Route] Listener closed for %s:%s", host.Host, host.Port)
					return nil
				}
				rps.api_ctx.LogErrorf("[Reverse Proxy] [TCP Route] Error accepting connection: %s", err)
				return err
			}
		}

		rps.activeConnections.Add(1)
		go func() {
			defer rps.activeConnections.Done()
			rps.handleTcpTraffic(conn, host.Host, fmt.Sprintf("%s:%s", host.TcpRoute.TargetHost, host.TcpRoute.TargetPort))
		}()
	}
}

func (rps *ReverseProxyService) listenHttpRoute(host *data_models.ReverseProxyHost, errorChan chan error) error {
	if host.Port == "" {
		host.Port = "80"
	}
	target := host.Host

	if host.Port != "" {
		target = fmt.Sprintf("%s:%s", host.Host, host.Port)
	}

	for _, route := range host.HttpRoutes {
		if route.Pattern == "" {
			pattern, err := regexp.Compile(route.Path)
			if err != nil {
				rps.api_ctx.LogErrorf("[HTTP Route] Error compiling route pattern: %s", err)
				return err
			}

			route.RegexpPattern = pattern
		}
		if route.Pattern != "" {
			pattern, err := regexp.Compile(route.Pattern)
			if err != nil {
				rps.api_ctx.LogErrorf("[HTTP Route] Error compiling route pattern: %s", err)
				return err
			}

			route.RegexpPattern = pattern
		}
	}

	mux := http.NewServeMux()
	proxy := newReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if err != nil {
			rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Proxy error for %s: %v", r.URL.Path, err)
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					w.WriteHeader(http.StatusGatewayTimeout)
					return
				}
			}
			w.WriteHeader(http.StatusBadGateway)
		}
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		if startTime, ok := resp.Request.Context().Value("request_start_time").(time.Time); ok {
			duration := time.Since(startTime)
			rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Request completed in %v: %s",
				duration, resp.Request.URL.Path)
		}

		if resp.Request.Context().Err() == context.Canceled {
			rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Client closed connection gracefully for %s", resp.Request.URL.Path)
		}

		// Track response status codes
		rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Response status %d for %s",
			resp.StatusCode, resp.Request.URL.Path)

		if host.Cors != nil && host.Cors.Enabled {
			rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Modifying response headers for CORS")
			if len(host.Cors.AllowedOrigins) > 0 {
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Setting Access-Control-Allow-Origin to %s", strings.Join(host.Cors.AllowedOrigins, ","))
				resp.Header.Set("Access-Control-Allow-Origin", strings.Join(host.Cors.AllowedOrigins, ","))
			} else {
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Setting Access-Control-Allow-Origin to *")
				resp.Header.Set("Access-Control-Allow-Origin", "*")
			}
			if len(host.Cors.AllowedMethods) > 0 {
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Setting Access-Control-Allow-Methods to %s", strings.Join(host.Cors.AllowedMethods, ","))
				resp.Header.Set("Access-Control-Allow-Methods", strings.Join(host.Cors.AllowedMethods, ","))
			} else {
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Setting Access-Control-Allow-Methods to GET, POST, PUT, DELETE, OPTIONS")
				resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			}
			if len(host.Cors.AllowedHeaders) > 0 {
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Setting Access-Control-Allow-Headers to %s", strings.Join(host.Cors.AllowedHeaders, ","))
				resp.Header.Set("Access-Control-Allow-Headers", strings.Join(host.Cors.AllowedHeaders, ","))
			} else {
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Setting Access-Control-Allow-Headers to Content-Type, Authorization")
				resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
		}

		for _, route := range host.HttpRoutes {
			if route.RegexpPattern.MatchString(resp.Request.URL.Path) {
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Modifying response headers for route %s", route.Path)
				for key, value := range route.ResponseHeaders {
					rps.api_ctx.LogDebugf("[HTTP Route] Setting response header %s to %s", key, value)
					resp.Header.Set(key, value)
				}
			}
		}

		return nil
	}

	proxy.Director = func(req *http.Request) {
		target := host.Host
		requestID := req.Header.Get("X-Request-ID")
		rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] [%s] Request received for %s", requestID, req.URL.Path)

		if host.Port != "" {
			target = fmt.Sprintf("%s:%s", host.Host, host.Port)
		}

		if strings.EqualFold(target, req.Host) {
			matched := false
			for _, route := range host.HttpRoutes {
				if route.TargetHost == "" || route.TargetHost == "---" {
					rps.api_ctx.LogErrorf("[HTTP Route] [%s] Target host is required for route %s, skipping",
						requestID, route.Path)
					continue
				}

				if route.RegexpPattern.MatchString(req.URL.Path) {
					matched = true
					rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] [%s] Matched with proxy route %s",
						requestID, route.Path)

					forwardTo := route.TargetHost
					if route.TargetPort != "" {
						forwardTo = fmt.Sprintf("%s:%s", route.TargetHost, route.TargetPort)
					}

					if strings.HasPrefix(forwardTo, "http") {
						forwardTo = strings.TrimPrefix(forwardTo, "http://")
						forwardTo = strings.TrimPrefix(forwardTo, "https://")
					}

					scheme := "http"
					if route.Schema != "" {
						scheme = route.Schema
					}

					rps.api_ctx.LogInfof("[Reverse Proxy] [HTTP Route] [%s] Forwarding traffic from %s%s to %s",
						requestID, target, req.URL.Path, forwardTo)

					req.Host = forwardTo
					req.URL.Scheme = scheme
					req.URL.Host = forwardTo

					req.Header.Set("X-Forwarded-By", constants.ExecutableName)
					req.Header.Set("X-Forwarded-Host", forwardTo)
					req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)

					if req.URL.Path == "" {
						req.URL.Path = "/"
					}
					break
				}
			}

			if !matched {
				rps.api_ctx.LogWarnf("[Reverse Proxy] [HTTP Route] [%s] No matching route found for %s",
					requestID, req.URL.Path)
			}
		}
	}

	mux.Handle("/", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.WriteHeader(200)
				return
			}

			// Add request start time to context
			ctx := context.WithValue(r.Context(), "request_start_time", time.Now())
			r = r.WithContext(ctx)

			// Add request ID for tracing
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = helpers.GenerateId()
				r.Header.Set("X-Request-ID", requestID)
			}
			rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Starting request %s: %s",
				requestID, r.URL.Path)

			next.ServeHTTP(w, r)
		})
	}(proxy))

	rps.api_ctx.LogInfof("[Reverse Proxy] [HTTP Route] Listening to %s on port %s...", host.Host, host.Port)
	hostTarget := fmt.Sprintf("%s:%s", host.Host, host.Port)
	server := &http.Server{
		Addr:              hostTarget,
		Handler:           mux,
		ReadHeaderTimeout: readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    1 << 20,
		ConnState: func(conn net.Conn, state http.ConnState) {
			connID := fmt.Sprintf("%s-%s", conn.RemoteAddr(), helpers.GenerateId()[:8])
			switch state {
			case http.StateNew:
				rps.activeConnections.Add(1)
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] [%s] New connection from %s", connID, conn.RemoteAddr())
			case http.StateClosed, http.StateHijacked:
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] [%s] Connection %s from %s", connID, state, conn.RemoteAddr())
				rps.activeConnections.Done()
			case http.StateActive:
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] [%s] Connection active from %s", connID, conn.RemoteAddr())
			case http.StateIdle:
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] [%s] Connection idle from %s", connID, conn.RemoteAddr())
			}
		},
		ErrorLog: log.New(io.Discard, "", 0),
	}

	rps.httpListeners = append(rps.httpListeners, server)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				rps.api_ctx.LogErrorf("[Reverse Proxy] [HTTP Route] Server error for %s:%s - %v",
					host.Host, host.Port, err)
				errorChan <- fmt.Errorf("server error on %s:%s - %w", host.Host, host.Port, err)
			} else {
				rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Server closed normally for %s:%s",
					host.Host, host.Port)
			}
		}
	}()

	<-rps.ctx.Done()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxShutdown); err != nil {
		rps.api_ctx.LogErrorf("[Reverse Proxy] [HTTP Route] Error shutting down server: %s", err)
	}
	return nil
}

func (rps *ReverseProxyService) handleTcpTraffic(src net.Conn, host string, target string) {
	rps.activeConnections.Add(1)
	defer rps.activeConnections.Done()

	// Generate connection ID for tracking
	connID := helpers.GenerateId()
	startTime := time.Now()

	rps.api_ctx.LogInfof("[Reverse Proxy] [TCP Route] [%s] New connection from %s to %s",
		connID, src.RemoteAddr(), target)

	defer func() {
		duration := time.Since(startTime)
		rps.api_ctx.LogInfof("[Reverse Proxy] [TCP Route] [%s] Connection closed after %v",
			connID, duration)
	}()

	defer src.Close()

	dst, err := net.Dial("tcp", target)
	if err != nil {
		rps.api_ctx.LogErrorf("[Reverse Proxy] [TCP Route] [%s] Unable to connect to target: %s",
			connID, err)
		return
	}

	defer dst.Close()

	// Use a context-aware copy function with connection tracking
	ctx, cancel := context.WithCancel(rps.ctx)
	defer cancel()

	go rps.copyWithContextAndTracking(ctx, dst, src, connID, "client->target")
	rps.copyWithContextAndTracking(ctx, src, dst, connID, "target->client")
}

func (rps *ReverseProxyService) copyWithContextAndTracking(ctx context.Context, dst io.Writer, src io.Reader, connID string, direction string) {
	var bytesCopied int64
	startTime := time.Now()

	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			rps.api_ctx.LogDebugf("[Reverse Proxy] [%s] [%s] Context cancelled after copying %d bytes in %v",
				connID, direction, bytesCopied, time.Since(startTime))
			return
		default:
			n, err := src.Read(buf)
			if n > 0 {
				bytesCopied += int64(n)
				if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
					if writeErr != io.EOF {
						rps.api_ctx.LogDebugf("[Reverse Proxy] [%s] [%s] Write error: %v",
							connID, direction, writeErr)
					}
					return
				}
			}
			if err != nil {
				if err == io.EOF {
					rps.api_ctx.LogDebugf("[Reverse Proxy] [%s] [%s] Connection closed gracefully after copying %d bytes in %v",
						connID, direction, bytesCopied, time.Since(startTime))
				} else if netErr, ok := err.(net.Error); ok {
					rps.api_ctx.LogDebugf("[Reverse Proxy] [%s] [%s] Network error after copying %d bytes in %v: %v (timeout: %v)",
						connID, direction, bytesCopied, time.Since(startTime), netErr, netErr.Timeout())
				} else {
					rps.api_ctx.LogDebugf("[Reverse Proxy] [%s] [%s] Read error after copying %d bytes in %v: %v",
						connID, direction, bytesCopied, time.Since(startTime), err)
				}
				return
			}
		}
	}
}

func (rps *ReverseProxyService) initQueue() {
	rps.queueOnce.Do(func() {
		rps.opQueue = make(chan reverseProxyOperationRequest, 100)
		go rps.processQueue()
	})
}

func (rps *ReverseProxyService) processQueue() {
	for req := range rps.opQueue {
		// Create a timeout context for the operation
		ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)

		// Create error channel for the operation
		done := make(chan error, 1)

		// Run the operation in a goroutine
		go func() {
			done <- req.operation()
		}()

		// Wait for either timeout or completion
		select {
		case err := <-done:
			req.result <- err
		case <-ctx.Done():
			req.result <- fmt.Errorf("operation timed out after %v", operationTimeout)
		}

		cancel() // Clean up the context
	}
}

func newReverseProxy(target string) *httputil.ReverseProxy {
	url, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(url)
}
