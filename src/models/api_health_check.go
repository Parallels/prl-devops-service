package models

type ApiHealthCheck struct {
	Healthy      bool                 `json:"healthy"`
	Message      string               `json:"message,omitempty"`
	ErrorMessage string               `json:"error_message,omitempty"`
	Services     []ServiceHealthCheck `json:"services,omitempty"`
}

type ServiceHealthCheck struct {
	Name         string `json:"name"`
	Healthy      bool   `json:"healthy"`
	Message      string `json:"message,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func (s *ApiHealthCheck) GetHealthStatus() (bool, string) {
	if s == nil {
		return false, "ApiHealthCheck is nil"
	}

	unHealthServices := make([]string, 0)
	for _, service := range s.Services {
		if !service.Healthy {
			unHealthServices = append(unHealthServices, service.Name)
			break
		}
	}

	if len(unHealthServices) == 0 {
		return true, "All Services Running"
	} else if len(unHealthServices) > 0 && len(unHealthServices) < len(s.Services) {
		return false, "Service Degraded"
	} else {
		return false, "Service Unhealthy"
	}
}
