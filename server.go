package shortener

import(
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/moh-osman3/shortener/managers"
)

type server struct {
	manager managers.UrlManager
	logger *zap.Logger
	server http.Server
}

func NewServer(m managers.UrlManager, logger *zap.Logger, port string) *server {
	return &server{
		manager: m,
		logger: logger,
		server: http.Server{Addr: fmt.Sprintf(":%s", port)},
	}
}

func (s *server) AddDefaultRoutes() {
	http.HandleFunc("/create", s.manager.CreateUrlHandleFunc)
	http.HandleFunc("/delete", s.manager.DeleteUrlHandleFunc)
	http.HandleFunc("/", s.manager.GetUrlHandleFunc)
}

func (s *server) Serve() error {
	s.logger.Info("Starting server on port 3030")

	err := s.server.ListenAndServe()
	s.logger.Info("Shutting down server")
	return err
}

func (s *server) Shutdown() error {
	return s.server.Close()
}