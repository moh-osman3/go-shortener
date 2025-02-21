package def

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/moh-osman3/shortener/urls"
	"github.com/moh-osman3/shortener/managers"
)

type defaultUrlManager struct {
	cache map[string]urls.ShortUrl
	redisDb *redis.Client
	logger *zap.Logger
}

func NewDefaultUrlManager(logger *zap.Logger, redisDb *redis.Client) managers.UrlManager {
	return &defaultUrlManager{
		cache: make(map[string]urls.ShortUrl),
		logger: logger,
		redisDb: redisDb,
	}
}

func (m *defaultUrlManager) scanAndDeleteDb() {
	var cursor uint64
	ctx := context.Background()
	iter := m.redisDb.Scan(ctx, cursor, "*", 0).Iterator()
	if iter.Err() != nil {
		m.logger.Error("error retrieving redis db keys")
		return
	}
	for iter.Next(ctx) {
		err := m.deleteShortUrlFromDb(iter.Val())
		if err != nil {
			m.logger.Debug("error deleting key", zap.Error(err))
		}
	}
}

func (m *defaultUrlManager) scanAndDeleteCache() {
	for key, val := range m.cache {
		// only delete shorturl if its not zero (no expiration)
		if !val.GetExpiry().IsZero() && time.Now().After(val.GetExpiry()) {
			err := m.deleteShortUrlFromCache(key)
			if err != nil {
				m.logger.Debug("error deleting key", zap.Error(err))
			}
		}
	}
}

func (m *defaultUrlManager) Start(ctx context.Context) error {
	m.logger.Info("manager.go: starting url manager")

	// two separate monitors for cache and db.
	go func() {
		for {
			// todo: make this configurable
			time.Sleep(10*time.Second)
			m.scanAndDeleteCache()
		}
	}()

	go func() {
		for {
			// todo: make this configurable
			// in case db has a lot more keys than db, clean up expired keys less frequently
			time.Sleep(300*time.Second)
			m.scanAndDeleteDb()
		}
	}()
	return nil
}

func (m *defaultUrlManager) End() {
	m.logger.Info("manager.go: shutting down url manager")
}

func (m *defaultUrlManager) deleteShortUrlFromCache(key string) error {
	m.logger.Debug("manager.go: deleting short url from cache")
	if _, ok := m.cache[key]; !ok {
		m.logger.Debug("manager.go: deleting shorturl from cache that does not exist")
		return errors.New("deleting shorturl that does not exist")
	}
	delete(m.cache, key)
	return nil
}

func (m *defaultUrlManager) deleteShortUrlFromDb(key string) error {
	m.logger.Debug("manager.go: deleting short url from db")
	res := m.redisDb.Del(context.Background(), key)

	if res.Err() != nil {
		m.logger.Debug("manager.go: deleting shorturl from db that does not exist")
		return res.Err() 
	}

	return nil
}

func (m *defaultUrlManager) generateShortUrl(longUrl string, expiry time.Duration) urls.ShortUrl {
	hash := md5.Sum([]byte(longUrl))
	hashStr := base64.StdEncoding.EncodeToString(hash[:])
	shortUrl, ok := m.cache[hashStr]

	if ok {
		if longUrl == shortUrl.GetLongUrl() {
			return shortUrl
		}
		// hash collision return nil to retry to get a unique hash.
		return nil
	}

	// didn't find in cache so check db
	val, err := m.redisDb.Get(context.Background(), hashStr).Result()
	shortUrl = nil 
	if (err == nil) {
		shortUrl = urls.NewDefaultShortUrl("", "", time.Second)
		shortUrl.Unmarshal([]byte(val))
	}

	if shortUrl != nil && shortUrl.GetLongUrl() != longUrl {
		// hash collision return nil to retry to get a unique hash.
		fmt.Println("noooo")
		return nil
	} else if shortUrl != nil {
		return shortUrl
	}

	return urls.NewDefaultShortUrl(hashStr, longUrl, expiry)
}

func (m *defaultUrlManager) createShortUrl(longUrl string, expiry time.Duration) (urls.ShortUrl, error) {
	var shortUrl urls.ShortUrl
	// in case of hash collisions retry until you get a unique shortUrl.
	for {
		fmt.Println("loop")
		shortUrl = m.generateShortUrl(longUrl, expiry)
		if shortUrl != nil {
			break
		}
	}

	if m.cache == nil {
		m.logger.Error("manager db cache not initialized")
		return nil, errors.New("manager db cache not initialized")
	}

	m.logger.Debug("manager.go: successfully created short url")
	
	m.cache[shortUrl.GetId()] = shortUrl 
	shortUrlStr, err := shortUrl.Marshal()
	if err != nil {
		return nil, err
	}
	m.redisDb.Set(context.Background(), shortUrl.GetId(), shortUrlStr, 0)

	return shortUrl, nil
}

func (m *defaultUrlManager) getShortUrlFromStore(key string) (urls.ShortUrl, error) {
	shortUrl, ok := m.cache[key]

	if ok {
		return shortUrl, nil
	}

	val, err := m.redisDb.Get(context.Background(), key).Result()
	shortUrl = nil 
	if (err == nil) {
		shortUrl = urls.NewDefaultShortUrl("", "", time.Second)
		shortUrl.Unmarshal([]byte(val))
	}

	return shortUrl, err
}