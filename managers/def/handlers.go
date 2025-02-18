package def

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (m *defaultUrlManager) GetUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if path == "" {
		//log
		io.WriteString(w, "Not a valid short url")
		return
	}

	shortUrl, ok := m.db[path]
	fmt.Println(m.db)
	fmt.Println(path)
	
	if !ok || shortUrl == nil {
		io.WriteString(w, "Short url does not exist")
		return
	}

	io.WriteString(w, shortUrl.GetLongUrl())
}

type reqData struct {
	Url string `json:"url"`
	Expiry string `json:"expiry"`
}

func (m *defaultUrlManager) CreateUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	fmt.Println("body")
	fmt.Println(body)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	var reqData reqData
	json.Unmarshal(body, &reqData)
	fmt.Println("req data")
	fmt.Println(reqData)
	expiry, err := time.ParseDuration(reqData.Expiry)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	shortUrl, err := m.createShortUrl(reqData.Url, expiry)
	if err != nil {
		io.WriteString(w, err.Error())
	}
	io.WriteString(w, fmt.Sprintf("Successfully created short url: http://localhost:3030/%s", shortUrl.GetId()))
}