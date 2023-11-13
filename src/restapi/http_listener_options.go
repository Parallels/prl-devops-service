package restapi

type HttpListenerOptions struct {
	ApiPrefix               string
	HttpPort                string
	EnableTLS               bool
	TLSPort                 string
	TLSCertificate          string
	TLSPrivateKey           string
	UseAuthBackend          bool
	MongoDbConnectionString string
	DatabaseName            string
	EnableAuthentication    bool
	LogHealthChecks         bool
	PublicRegistration      bool
	DefaultApiVersion       string
}
