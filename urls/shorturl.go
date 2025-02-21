package urls

import (
	"net/http"
	"time"
)

type ShortUrl interface {
	GetId() string
	GetLongUrl() string
	fetchUrlContent(longUrl string) (*http.Response, error)
}

type defaultShortUrl struct {
	id string 
	longUrl string
	expiry time.Time
}

func NewDefaultShortUrl(id string, longUrl string) ShortUrl {
	return &defaultShortUrl{
		id: id,
		longUrl: longUrl,
	}
}

func (su *defaultShortUrl) GetId() string {
	return su.id
}

func (su *defaultShortUrl) GetLongUrl() string {
	return su.longUrl
}

func (su *defaultShortUrl) fetchUrlContent(longUrl string) (*http.Response, error) {return nil, nil}