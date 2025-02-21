package shortener

import(
	"net/http"

	"github.com/moh-osman3/shortener/managers"
)

type server struct {
	manager managers.UrlManager
	// TODO put a real db here
}

func NewServer(m managers.UrlManager) *server {
	return &server{
		manager: m,
	}
}

func (s *server) AddDefaultRoutes() {
	http.HandleFunc("/create", s.manager.CreateUrlHandleFunc)
	http.HandleFunc("/delete", s.manager.DeleteUrlHandleFunc)
	http.HandleFunc("/", s.manager.GetUrlHandleFunc)
}

func (s *server) Serve() error {
	err := http.ListenAndServe(":3030", nil)
	return err
}