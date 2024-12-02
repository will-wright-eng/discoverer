package service

type Service struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Path     string `json:"path"`
	Protocol string `json:"protocol"`
}

type ServiceConfig struct {
	Services []Service `json:"services"`
}
