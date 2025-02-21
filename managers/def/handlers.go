package def

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type deleteData struct {
	Id string `json:"id"`
}

func (m *defaultUrlManager) DeleteUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid method: expected DELETE request", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var deleteData deleteData
	json.Unmarshal(body, &deleteData)

	err = m.deleteKeyFromCacheAndDb(deleteData.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, "Successfully deleted short url!")
}

func (m *defaultUrlManager) GetUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method: expected GET request", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	paths := strings.Split(path, "/")
	if len(paths) == 0 || paths[0] == ""  || len(paths) > 2 {
		http.Error(w, "Invalid request URL", http.StatusBadRequest)
		return
	}

	shortUrl, err := m.getShortUrlFromStore(paths[0])

	if err != nil || shortUrl == nil {
		http.Error(w, "short url does not exist", http.StatusInternalServerError)
		return
	}

	// This is a normal short url request and not a summary request
	if len(paths) == 1 {
		m.AddCallToCacheAndDb(shortUrl)
		io.WriteString(w, shortUrl.GetLongUrl())
		return
	}

	if len(paths) == 2 && paths[1] == "summary" {
		io.WriteString(w, shortUrl.GetSummary())
		return
	}

	http.Error(w, "Invalid request URL", http.StatusBadRequest)
}

type createData struct {
	Url    string `json:"url"`
	Expiry string `json:"expiry"`
}

func (m *defaultUrlManager) CreateUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method: expected POST request", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var createData createData
	json.Unmarshal(body, &createData)
	expiry, err := time.ParseDuration(createData.Expiry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortUrl, err := m.createShortUrl(createData.Url, expiry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, fmt.Sprintf("Successfully created short url: http://localhost:3030/%s", shortUrl.GetId()))
}
