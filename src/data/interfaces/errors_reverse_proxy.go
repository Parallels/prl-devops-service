package interfaces

// Sentinels relocated from data/reverse_proxy.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrReverseProxyHostEmptyNameOrId           = errors.NewWithCode("no reverse proxy host specified", 500)
	ErrorReverseProxyHostNotFound              = errors.NewWithCode("reverse proxy host not found", 404)
	ErrorReverseProxyHttpRouteNotFound         = errors.NewWithCode("reverse proxy host http route not found", 404)
	ErrorReverseProxyHttpRouteInvalidOrder     = errors.NewWithCode("invalid reverse proxy host http route order", 400)
	ErrorReverseProxyHttpRouteOrderUpdate      = errors.NewWithCode("cannot update reverse proxy host http route and change order at the same time", 400)
	ErrorReverseProxyTcpRouteNotFound          = errors.NewWithCode("reverse proxy host tcp route not found", 404)
	ErrorReverseProxyTcpRouteWithHttpRouteHost = errors.NewWithCode("cannot update reverse proxy TCP route when HTTP routes are present", 400)
)
