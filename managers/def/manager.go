package def

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/moh-osman3/shortener/urls"
	"github.com/moh-osman3/shortener/managers"
)

type defaultUrlManager struct {
	db map[string]urls.ShortUrl
}

func NewUrlManager() managers.UrlManager {
	return &defaultUrlManager{db:make(map[string]urls.ShortUrl)}
}

func (m *defaultUrlManager) scanAndDelete() {
	for key, val := range m.db {
		fmt.Println(time.Now().After(val.GetExpiry()))
		fmt.Println(val.GetExpiry())
		if time.Now().After(val.GetExpiry()) {
			m.deleteShortUrl(key)
		}
	}
}

func (m *defaultUrlManager) Start(ctx context.Context) error {

	go func() {
		for {
			m.scanAndDelete()
			// todo: make this configurable
			time.Sleep(10*time.Second)
		}

	}()
	return nil
}

func (m *defaultUrlManager) End() {}
func (m *defaultUrlManager) deleteShortUrl(key string) {
	fmt.Println("DELETING FROM DB")
	fmt.Println(m.db[key])
	delete(m.db, key)
}

func (m *defaultUrlManager) generateShortUrl(longUrl string, expiry time.Duration) urls.ShortUrl {
	hash := md5.Sum([]byte(longUrl))
	hashStr := base64.StdEncoding.EncodeToString(hash[:])
	shortUrl, ok := m.db[hashStr]
	if ok {
		if longUrl == shortUrl.GetLongUrl() {
			return shortUrl
		}
		// hash collision return nil to retry to get a unique hash.
		return nil
	}

	return urls.NewDefaultShortUrl(hashStr, longUrl, expiry)
}

func (m *defaultUrlManager) createShortUrl(longUrl string, expiry time.Duration) (urls.ShortUrl, error) {
	var shortUrl urls.ShortUrl
	// in case of hash collisions retry until you get a unique shortUrl.
	for {
		shortUrl = m.generateShortUrl(longUrl, expiry)
		if shortUrl != nil {
			break
		}
	}

	if m.db == nil {
		return nil, errors.New("manager db cache not initialized")
	}
	
	m.db[shortUrl.GetId()] = shortUrl 

	return shortUrl, nil
}