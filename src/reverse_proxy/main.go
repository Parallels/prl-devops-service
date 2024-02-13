package reverse_proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
)

type ReverseProxyService struct {
	cfg          *config.ReverseProxyConfig
	ctx          basecontext.ApiContext
	error        chan error
	tcpListeners []net.Listener
}

func New(ctx basecontext.ApiContext) *ReverseProxyService {
	cfg := config.Get()
	if cfg.GetReverseProxyConfig() == nil {
		return nil
	}

	return &ReverseProxyService{
		ctx:          ctx,
		cfg:          cfg.GetReverseProxyConfig(),
		tcpListeners: []net.Listener{},
	}
}

func (rps *ReverseProxyService) Start() error {
	if rps.cfg.Port != "" {
		rps.cfg.Port = "8080"
		rps.ctx.LogWarnf("Port not set for reverse proxy, using default port 8080")
	}
	if rps.cfg.Host == "" {
		rps.cfg.Host = "localhost"
		rps.ctx.LogWarnf("Host not set for reverse proxy, using default host localhost")
	}

	rps.ctx.LogInfof("Starting reverse proxy on %s:%s", rps.cfg.Host, rps.cfg.Port)
	for _, host := range rps.cfg.Hosts {
		go rps.startServer(host)
	}

	err := <-rps.error
	if err != nil {
		rps.ctx.LogErrorf("Error starting reverse proxy: %s", err)
	}

	select {}
}

func (rps *ReverseProxyService) Stop() error {
	for _, listener := range rps.tcpListeners {
		if err := listener.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (rps *ReverseProxyService) Restart() error {
	if err := rps.Stop(); err != nil {
		return err
	}

	if err := rps.Start(); err != nil {
		return err
	}

	return nil
}

func (rps *ReverseProxyService) startServer(host *config.ReverseProxyConfigHost) {
	// Checking if we have a tcp route to deal with
	if host.TcpRoute != nil {
		if err := rps.listenTcpRoute(host); err != nil {
			rps.error <- err
		}
	} else {
		if err := rps.listenHttpRoute(host); err != nil {
			rps.error <- err
		}
	}
}

func (rps *ReverseProxyService) listenTcpRoute(host *config.ReverseProxyConfigHost) error {
	if host.TcpRoute.TargetPort == "" {
		return fmt.Errorf("[TCP Route] port is required for starting a tcp route")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host.Host, host.Port))
	rps.tcpListeners = append(rps.tcpListeners, listener)
	if err != nil {
		return err
	}

	defer listener.Close()
	rps.ctx.LogInfof("[TCP Route] Listening on %s:%s", host.Host, host.Port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			rps.ctx.LogErrorf("[TCP Route] Error accepting connection: %s", err)
			return err
		}
		if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
			rps.ctx.LogErrorf("[TCP Route] Error setting deadline for connection: %s", err)
			if err := conn.Close(); err != nil {
				rps.ctx.LogErrorf("[TCP Route] Error closing connection: %s", err)
				return err
			}
		}

		go rps.handleTcpTraffic(conn, host.Host, fmt.Sprintf("%s:%s", host.TcpRoute.TargetHost, host.TcpRoute.TargetPort))
	}
}

func (rps *ReverseProxyService) listenHttpRoute(host *config.ReverseProxyConfigHost) error {
	if host.Port == "" {
		host.Port = "80"
	}
	target := host.Host

	if host.Port != "" {
		target = fmt.Sprintf("%s:%s", host.Host, host.Port)
	}
	for _, route := range host.HttpRoutes {
		if route.Pattern == nil {
			pattern, err := regexp.Compile(route.Path)
			if err != nil {
				rps.ctx.LogErrorf("[HTTP Route] Error compiling route pattern: %s", err)
				return err
			}

			route.Pattern = pattern
		}
	}

	mux := http.NewServeMux()
	proxy := newReverseProxy(target)
	proxy.ModifyResponse = func(resp *http.Response) error {
		if rps.cfg.Cors != nil && rps.cfg.Cors.Enabled {
			rps.ctx.LogDebugf("[HTTP Route] Modifying response headers for CORS")
			if len(rps.cfg.Cors.AllowedOrigins) > 0 {
				rps.ctx.LogDebugf("[HTTP Route] Setting Access-Control-Allow-Origin to %s", strings.Join(rps.cfg.Cors.AllowedOrigins, ","))
				resp.Header.Set("Access-Control-Allow-Origin", strings.Join(rps.cfg.Cors.AllowedOrigins, ","))
			} else {
				rps.ctx.LogDebugf("[HTTP Route] Setting Access-Control-Allow-Origin to *")
				resp.Header.Set("Access-Control-Allow-Origin", "*")
			}
			if len(rps.cfg.Cors.AllowedMethods) > 0 {
				rps.ctx.LogDebugf("[HTTP Route] Setting Access-Control-Allow-Methods to %s", strings.Join(rps.cfg.Cors.AllowedMethods, ","))
				resp.Header.Set("Access-Control-Allow-Methods", strings.Join(rps.cfg.Cors.AllowedMethods, ","))
			} else {
				rps.ctx.LogDebugf("[HTTP Route] Setting Access-Control-Allow-Methods to GET, POST, PUT, DELETE, OPTIONS")
				resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			}
			if len(rps.cfg.Cors.AllowedHeaders) > 0 {
				rps.ctx.LogDebugf("[HTTP Route] Setting Access-Control-Allow-Headers to %s", strings.Join(rps.cfg.Cors.AllowedHeaders, ","))
				resp.Header.Set("Access-Control-Allow-Headers", strings.Join(rps.cfg.Cors.AllowedHeaders, ","))
			} else {
				rps.ctx.LogDebugf("[HTTP Route] Setting Access-Control-Allow-Headers to Content-Type, Authorization")
				resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
		}

		for _, route := range host.HttpRoutes {
			if route.Pattern.MatchString(resp.Request.URL.Path) {
				rps.ctx.LogDebugf("[HTTP Route] Modifying response headers for route %s", route.Path)
				for key, value := range route.ResponseHeaders {
					rps.ctx.LogDebugf("[HTTP Route] Setting response header %s to %s", key, value)
					resp.Header.Set(key, value)
				}
			}
		}

		return nil
	}

	proxy.Director = func(req *http.Request) {
		target := host.Host
		rps.ctx.LogDebugf("[HTTP Route] Request received for %s", req.URL.Path)
		if host.Port != "" {
			target = fmt.Sprintf("%s:%s", host.Host, host.Port)
		}

		if strings.EqualFold(target, req.Host) {
			for _, route := range host.HttpRoutes {
				if route.Pattern.MatchString(req.URL.Path) {
					rps.ctx.LogDebugf("[HTTP Route] Matched with proxy route %s", route.Path)
					forwardTo := route.TargetHost
					if route.TargetPort != "" {
						forwardTo = fmt.Sprintf("%s:%s", route.TargetHost, route.TargetPort)
					}

					if strings.HasPrefix(forwardTo, "http") {
						forwardTo = strings.TrimPrefix(forwardTo, "http://")
						forwardTo = strings.TrimPrefix(forwardTo, "https://")
					}
					scheme := "http"
					if route.Scheme != "" {
						scheme = route.Scheme
					}

					rps.ctx.LogInfof("[HTTP Route] Forwarding http traffic from host %s%s to proxy on %s", target, req.URL.Path, forwardTo)
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

	rps.ctx.LogInfof("[HTTP Route] Listening to %s on port %s...", host.Host, host.Port)
	hostTarget := fmt.Sprintf("%s:%s", host.Host, host.Port)
	server := &http.Server{
		Addr:              hostTarget,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		if !strings.Contains(err.Error(), "http: Server closed") {
			rps.ctx.LogErrorf("There was an error shutting down the https server: %v", err.Error())
		}
	}

	return nil
}

func (rps *ReverseProxyService) handleTcpTraffic(src net.Conn, host string, target string) {
	rps.ctx.LogInfof("[TCP Route] Forwarding tcp traffic from host %s to proxy on %s", host, target)

	dst, err := net.Dial("tcp", target)
	if err != nil {
		rps.ctx.LogErrorf("[TCP Route] Unable to connect to target: %s", err)
		if err := src.Close(); err != nil {
			rps.ctx.LogErrorf("[TCP Route] Unable to close source connection: %s", err)
			rps.error <- err
		}

		rps.error <- err
	}
	defer dst.Close()

	go func() {
		// forward traffic from source to destination
		if _, err := io.Copy(dst, src); err != nil {
			rps.ctx.LogErrorf("[TCP Route] error forwarding package to host %s, err: %v", target, err.Error())
		}
	}()

	// forward traffic from destination to source
	if _, err := io.Copy(src, dst); err != nil {
		rps.ctx.LogErrorf("[TCP Route] error forwarding package to host %s, err: %v", target, err.Error())
	}
}

func newReverseProxy(target string) *httputil.ReverseProxy {
	url, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(url)
}
