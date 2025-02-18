package urls

import (
	"fmt"
	"net/http"
	"time"
)

type ShortUrl interface {
	GetId() string
	GetLongUrl() string
	GetExpiry() time.Time
	fetchUrlContent(longUrl string) (*http.Response, error)
}

type defaultShortUrl struct {
	id string 
	longUrl string
	expiry time.Time
}

func NewDefaultShortUrl(id string, longUrl string, expiry time.Duration) ShortUrl {
	su := &defaultShortUrl{
		id: id,
		longUrl: longUrl,
	}
	fmt.Println("exiry in shorturl", expiry)

	if expiry == 0 {
		// default behavior
		su.expiry = time.Now().AddDate(1, 0, 0)
	} else if expiry < 0 {
		su.expiry = time.Time{}
	} else {
		su.expiry = time.Now().Add(expiry)
	}
	return su
}

func (su *defaultShortUrl) GetId() string {
	return su.id
}
func (su *defaultShortUrl) GetExpiry() time.Time {
	return su.expiry
}

func (su *defaultShortUrl) GetLongUrl() string {
	return su.longUrl
}

func (su *defaultShortUrl) fetchUrlContent(longUrl string) (*http.Response, error) {return nil, nil}