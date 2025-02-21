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
		io.WriteString(w, "Invalid method: expected DELETE request")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	var deleteData deleteData
	json.Unmarshal(body, &deleteData)

	err = m.deleteKeyFromCacheAndDb(deleteData.Id)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	io.WriteString(w, "Successfully deleted short url!")
}

func (m *defaultUrlManager) GetUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		io.WriteString(w, "Invalid method: expected GET request")
		return
	}

	path := r.URL.Path[1:]
	paths := strings.Split(path, "/")
	if len(paths) == 0 || paths[0] == "" {
		//log
		io.WriteString(w, "Not a valid short url")
		return
	}

	shortUrl, err := m.getShortUrlFromStore(paths[0])

	if err != nil || shortUrl == nil {
		io.WriteString(w, "Short url does not exist")
		return
	}

	// This is a normal short url request and not a summary request
	if len(paths) == 1 {
		m.AddCallToCacheAndDb(shortUrl)
		http.Redirect(w, r, shortUrl.GetLongUrl(), http.StatusFound)
		return
	}

	if len(paths) == 2 && paths[1] == "summary" {
		io.WriteString(w, shortUrl.GetSummary())
		return
	}

	io.WriteString(w, "Invalid GET endpoint")
}

type createData struct {
	Url    string `json:"url"`
	Expiry string `json:"expiry"`
}

func (m *defaultUrlManager) CreateUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		io.WriteString(w, "Invalid method: expected POST request")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	var createData createData
	json.Unmarshal(body, &createData)
	expiry, err := time.ParseDuration(createData.Expiry)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	shortUrl, err := m.createShortUrl(createData.Url, expiry)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, fmt.Sprintf("Successfully created short url: http://localhost:3030/%s", shortUrl.GetId()))
}
