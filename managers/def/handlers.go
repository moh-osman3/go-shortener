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
	body, err := io.ReadAll(r.Body)
	fmt.Println("body")
	fmt.Println(body)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	var deleteData deleteData
	json.Unmarshal(body, &deleteData)

	err = m.deleteShortUrl(deleteData.Id)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	io.WriteString(w, "Successfully deleted short url!")
}

func (m *defaultUrlManager) GetUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	paths := strings.Split(path, "/")
	fmt.Println("PATHS")
	fmt.Println(paths[0])
	if len(paths) == 0 || paths[0] == "" {
		//log
		io.WriteString(w, "Not a valid short url")
		return
	}

	fmt.Println(m.db)
	shortUrl, ok := m.db[paths[0]]
	
	if !ok || shortUrl == nil {
		io.WriteString(w, "Short url does not exist")
		return
	}

	// This is a normal short url request and not a summary request
	if len(paths) == 1 {
		shortUrl.AddCall(time.Now())
		io.WriteString(w, shortUrl.GetLongUrl())
		return
	}

	io.WriteString(w, shortUrl.GetSummary())
}

type createData struct {
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
	var createData createData
	json.Unmarshal(body, &createData)
	fmt.Println("req data")
	fmt.Println(createData)
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