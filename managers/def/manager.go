package def

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/moh-osman3/shortener/urls"
	"github.com/moh-osman3/shortener/managers"
)

type defaultUrlManager struct {
	db map[string]urls.ShortUrl
	logger *zap.Logger
}

func NewDefaultUrlManager(logger *zap.Logger) managers.UrlManager {
	return &defaultUrlManager{
		db: make(map[string]urls.ShortUrl),
		logger: logger,
	}
}

func (m *defaultUrlManager) scanAndDelete() {
	for key, val := range m.db {
		// only delete shorturl if its not zero (no expiration)
		if !val.GetExpiry().IsZero() && time.Now().After(val.GetExpiry()) {
			m.deleteShortUrl(key)
		}
	}
}

func (m *defaultUrlManager) Start(ctx context.Context) error {
	m.logger.Info("manager.go: starting url manager")
	go func() {
		for {
			m.scanAndDelete()
			// todo: make this configurable
			time.Sleep(10*time.Second)
		}
	}()
	return nil
}

func (m *defaultUrlManager) End() {
	m.logger.Info("manager.go: shutting down url manager")
}

func (m *defaultUrlManager) deleteShortUrl(key string) error {
	m.logger.Debug("manager.go: deleting short url from db")
	if _, ok := m.db[key]; !ok {
		m.logger.Debug("manager.go: deleting shorturl that does not exist")
		return errors.New("deleting shorturl that does not exist")
	}
	delete(m.db, key)
	return nil
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
		m.logger.Error("manager db cache not initialized")
		return nil, errors.New("manager db cache not initialized")
	}

	m.logger.Debug("manager.go: successfully created short url")
	
	m.db[shortUrl.GetId()] = shortUrl 

	return shortUrl, nil
}