package api

type Server struct {
	deviceService DeviceService
}

func NewServer(deviceService DeviceService) *Server {
	return &Server{
		deviceService: deviceService,
	}
}
