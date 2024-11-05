package reverse_proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

var globalReverseProxyService *ReverseProxyService

type ReverseProxyService struct {
	enabled          bool
	host             string
	port             string
	forwarding_hosts []*data_models.ReverseProxyHost
	db               *data.JsonDatabase
	api_ctx          basecontext.ApiContext
	ctx              context.Context
	cancelFunc       context.CancelFunc
	tcpListeners     []net.Listener
	httpListeners    []*http.Server
	wg               *sync.WaitGroup
}

func Get(ctx basecontext.ApiContext) *ReverseProxyService {
	if globalReverseProxyService == nil {
		globalReverseProxyService = New(ctx)
	}

	return globalReverseProxyService
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

func (rps *ReverseProxyService) LoadFromDb() error {
	dtoRp, err := rps.db.GetReverseProxyConfig(rps.api_ctx)
	if err != nil {
		return err
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
						rps.api_ctx.LogErrorf("Error getting vm %s for reverse proxy route: %s", route.TargetVmId, err)
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
				rps.api_ctx.LogErrorf("Error getting vm %s for reverse proxy tcp route: %s", hostCopy.TcpRoute.TargetVmId, err)
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

	if rps.port == "" {
		rps.port = "8080"
		rps.api_ctx.LogWarnf("[Reverse Proxy] Port not set for reverse proxy, using default port 8080")
	}
	if rps.host == "" {
		rps.host = "localhost"
		rps.api_ctx.LogWarnf("[Reverse Proxy] Host not set for reverse proxy, using default host localhost")
	}

	rps.api_ctx.LogInfof("[Reverse Proxy] Starting reverse proxy on %s:%s", rps.host, rps.port)
	go rps.startServer(errorChan)

	select {
	case err := <-errorChan:
		if err != nil {
			rps.api_ctx.LogErrorf("[Reverse Proxy] Error starting reverse proxy: %s", err)
		}
		return err
		// case <-rps.ctx.Done():
		// 	rps.api_ctx.LogInfof("[Reverse Proxy] Stopping reverse proxy due to context cancellation")
		// 	return nil
	default:
	}

	return nil
}

func (rps *ReverseProxyService) Stop() error {
	rps.api_ctx.LogInfof("[Reverse Proxy] Stopping reverse proxy service")
	rps.cancelFunc()

	for _, listener := range rps.tcpListeners {
		rps.api_ctx.LogInfof("[Reverse Proxy] [TCP Route] Closing listener")
		if err := listener.Close(); err != nil {
			rps.api_ctx.LogErrorf("[Reverse Proxy] [TCP Route] Error closing listener: %s", err)
		}
		rps.api_ctx.LogInfof("[Reverse Proxy] [TCP Route] Listener closed")
	}

	for _, server := range rps.httpListeners {
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rps.api_ctx.LogInfof("[Reverse Proxy] [HTTP Route] Closing server")
		if err := server.Shutdown(ctxShutdown); err != nil {
			rps.api_ctx.LogErrorf("[Reverse Proxy] [HTTP Route] Error shutting down server: %s", err)
		}

		rps.api_ctx.LogInfof("[Reverse Proxy] [HTTP Route] Server closed")
	}

	rps.wg.Wait()

	rps.api_ctx.LogInfof("[Reverse Proxy] Reverse proxy service stopped")
	return nil
}

func (rps *ReverseProxyService) Restart() error {
	if err := rps.Stop(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- rps.Start()
	}()

	return <-done
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
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-rps.ctx.Done():
				return nil
			default:
			}

			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				return nil // Listener closed
			}

			rps.api_ctx.LogErrorf("[Reverse Proxy] [TCP Route] Error accepting connection: %s", err)
			return err
		}

		// if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		// 	rps.api_ctx.LogErrorf("[Reverse Proxy] [TCP Route] Error setting deadline for connection: %s", err)
		// 	if err := conn.Close(); err != nil {
		// 		rps.api_ctx.LogErrorf("[Reverse Proxy] [TCP Route] Error closing connection: %s", err)
		// 		return err
		// 	}
		// }

		go rps.handleTcpTraffic(conn, host.Host, fmt.Sprintf("%s:%s", host.TcpRoute.TargetHost, host.TcpRoute.TargetPort))
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
	proxy.ModifyResponse = func(resp *http.Response) error {
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
		rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Request received for %s", req.URL.Path)
		if host.Port != "" {
			target = fmt.Sprintf("%s:%s", host.Host, host.Port)
		}

		if strings.EqualFold(target, req.Host) {
			for _, route := range host.HttpRoutes {
				if route.TargetHost == "" || route.TargetHost == "---" {
					rps.api_ctx.LogErrorf("[HTTP Route] target host is required for starting a http route, skipping route %s", route.Path)
					continue
				}
				if route.RegexpPattern.MatchString(req.URL.Path) {
					rps.api_ctx.LogDebugf("[Reverse Proxy] [HTTP Route] Matched with proxy route %s", route.Path)
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

					rps.api_ctx.LogInfof("[Reverse Proxy] [HTTP Route] Forwarding http traffic from host %s%s to proxy on %s", target, req.URL.Path, forwardTo)
					req.Host = forwardTo
					req.URL.Scheme = scheme
					req.URL.Host = forwardTo

					req.Header.Add("X-Forwarded-By", constants.ExecutableName)
					req.Header.Add("X-Forwarded-Host", forwardTo)
					req.Header.Add("X-Forwarded-Proto", req.URL.Scheme)

					// req.URL.Path = route.Pattern.ReplaceAllString(req.URL.Path, "")
					if req.URL.Path == "" {
						req.URL.Path = "/"
					}

					break
				}
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
			next.ServeHTTP(w, r)
		})
	}(proxy))

	rps.api_ctx.LogInfof("[Reverse Proxy] [HTTP Route] Listening to %s on port %s...", host.Host, host.Port)
	hostTarget := fmt.Sprintf("%s:%s", host.Host, host.Port)
	server := &http.Server{
		Addr:              hostTarget,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	rps.httpListeners = append(rps.httpListeners, server)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			rps.api_ctx.LogErrorf("There was an error starting the HTTP server: %v", err.Error())
			errorChan <- err
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
	rps.api_ctx.LogInfof("[Reverse Proxy] [TCP Route] Forwarding tcp traffic from host %s to proxy on %s", host, target)

	defer src.Close()

	dst, err := net.Dial("tcp", target)
	if err != nil {
		rps.api_ctx.LogErrorf("[Reverse Proxy] [TCP Route] Unable to connect to target: %s", err)
		return
	}

	defer dst.Close()

	// go func() {
	// 	// forward traffic from source to destination
	// 	if _, err := io.Copy(dst, src); err != nil {
	// 		rps.api_ctx.LogErrorf("[Reverse Proxy] [TCP Route] error forwarding package to host %s, err: %v", target, err.Error())
	// 	}
	// }()

	// forward traffic from destination to source
	// if _, err := io.Copy(src, dst); err != nil {
	// 	rps.api_ctx.LogErrorf("[Reverse Proxy] [TCP Route] error forwarding package to host %s, err: %v", target, err.Error())
	// }

	// Use a context-aware copy function
	go rps.copyWithContext(rps.ctx, dst, src)
	rps.copyWithContext(rps.ctx, src, dst)
}

func (rps *ReverseProxyService) copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) {
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := src.Read(buf)
			if n > 0 {
				if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}
}

func newReverseProxy(target string) *httputil.ReverseProxy {
	url, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(url)
}
