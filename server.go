package shortener

import(
	"net/http"

	"go.uber.org/zap"

	"github.com/moh-osman3/shortener/managers"
)

type server struct {
	manager managers.UrlManager
	logger *zap.Logger
}

func NewServer(m managers.UrlManager, logger *zap.Logger) *server {
	return &server{
		manager: m,
		logger: logger,
	}
}

func (s *server) AddDefaultRoutes() {
	http.HandleFunc("/create", s.manager.CreateUrlHandleFunc)
	http.HandleFunc("/delete", s.manager.DeleteUrlHandleFunc)
	http.HandleFunc("/", s.manager.GetUrlHandleFunc)
}

func (s *server) Serve() error {
	s.logger.Info("Starting server on port 3030")
	err := http.ListenAndServe(":3030", nil)
	s.logger.Info("Shutting down server")
	return err
}