package def

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.uber.org/zap"

	"github.com/moh-osman3/shortener/managers"
	"github.com/moh-osman3/shortener/urls"
)

// this helps with testing with a mock db
type DB interface {
	Get(key []byte, ro *opt.ReadOptions) ([]byte, error)
	Put(key, value []byte, wo *opt.WriteOptions) error
	Delete(key []byte, wo *opt.WriteOptions) error
	NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator
}

type defaultUrlManager struct {
	cache   map[string]urls.ShortUrl
	leveldb DB
	logger  *zap.Logger
	lock    sync.RWMutex
	shutdownCh chan struct{}
}

func NewDefaultUrlManager(logger *zap.Logger, levelDb DB) managers.UrlManager {
	return &defaultUrlManager{
		cache:   make(map[string]urls.ShortUrl),
		logger:  logger,
		leveldb: levelDb,
		shutdownCh: make(chan struct{}, 1),
	}
}

func (m *defaultUrlManager) deleteKeyFromCacheAndDb(key string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	err1 := m.deleteShortUrlFromDb(key)
	err2 := m.deleteShortUrlFromCache(key)

	// only return error if key does not exist in both cache and db
	if err1 != nil && err2 != nil {
		return errors.New("manager.go: deleting shorturl from cache that does not exist")
	}
	return nil
}

func (m *defaultUrlManager) scanAndDeleteDb() {
	m.lock.RLock()
	defer m.lock.RUnlock()
	iter := m.leveldb.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		shortUrl := urls.NewDefaultShortUrl("", "", time.Second, time.Now())
		shortUrl.Unmarshal([]byte(iter.Value()))

		if !shortUrl.GetExpiry().IsZero() && time.Now().After(shortUrl.GetExpiry()) {
			err := m.deleteKeyFromCacheAndDb(shortUrl.GetId())
			if err != nil {
				m.logger.Debug("error deleting key", zap.Error(err))
			}
		}
	}

	if iter.Error() != nil {
		m.logger.Error("error retrieving redis db keys")
		return
	}
}

func (m *defaultUrlManager) scanAndDeleteCache() {
	for key, val := range m.cache {
		if !val.GetExpiry().IsZero() && time.Now().After(val.GetExpiry()) {
			err := m.deleteKeyFromCacheAndDb(key)
			if err != nil {
				m.logger.Debug("error deleting key", zap.Error(err))
			}
		}
	}
}

func (m *defaultUrlManager) Start(ctx context.Context, cacheInterval time.Duration, dbInterval time.Duration) error {
	m.logger.Info("manager.go: starting url manager")

	// todo: make interval configurable
	cacheTicker := time.NewTicker(cacheInterval)

	// in case db has a lot more keys than db, clean up expired keys less frequently
	dbTicker := time.NewTicker(dbInterval)

	// two separate monitors for cache and db.
	go func() {
		defer cacheTicker.Stop()
		for {
			select{
			case <-m.shutdownCh:
				return
			case <-cacheTicker.C:
				m.scanAndDeleteCache()
			}
		}
	}()

	go func() {
		defer dbTicker.Stop()
		for {
			select{
			case <-m.shutdownCh:
				return
			case <-dbTicker.C:
				m.scanAndDeleteDb()
			}
		}
	}()
	return nil
}

func (m *defaultUrlManager) End() {
	m.logger.Info("manager.go: shutting down url manager")
	close(m.shutdownCh)
}

func (m *defaultUrlManager) deleteShortUrlFromCache(key string) error {
	m.logger.Debug("manager.go: deleting short url from cache")
	if _, ok := m.cache[key]; !ok {
		m.logger.Debug("manager.go: deleting shorturl from cache that does not exist")
		return errors.New("manager.go: deleting shorturl from cache that does not exist")
	}
	delete(m.cache, key)
	return nil
}

func (m *defaultUrlManager) deleteShortUrlFromDb(key string) error {
	m.logger.Debug("manager.go: deleting short url from db")
	err := m.leveldb.Delete([]byte(key), nil)

	if err != nil {
		m.logger.Debug("manager.go: deleting shorturl from db that does not exist")
		return err
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
	val, err := m.leveldb.Get([]byte(hashStr), nil)
	shortUrl = nil
	if err == nil {
		shortUrl = urls.NewDefaultShortUrl("", "", time.Second, time.Now())
		shortUrl.Unmarshal([]byte(val))
	}

	if shortUrl != nil && shortUrl.GetLongUrl() != longUrl {
		// hash collision return nil to retry to get a unique hash.
		return nil
	} else if shortUrl != nil {
		return shortUrl
	}

	return urls.NewDefaultShortUrl(hashStr, longUrl, expiry, time.Now())
}

func (m *defaultUrlManager) createShortUrl(longUrl string, expiry time.Duration) (urls.ShortUrl, error) {
	var shortUrl urls.ShortUrl
	// in case of hash collisions retry 10 times until you get a unique shortUrl.
	for i := 0; i < 10; i++ {
		shortUrl = m.generateShortUrl(longUrl, expiry)
		if shortUrl != nil {
			break
		}
	}

	if shortUrl == nil {
		m.logger.Error("unable to generate unique short url")
		return nil, errors.New("manager.go: unable to generate new short url")
	}

	if m.cache == nil {
		m.logger.Error("manager db cache not initialized")
		return nil, errors.New("manager.go: manager db cache not initialized")
	}

	m.logger.Debug("manager.go: successfully created short url")
	shortUrlStr, err := shortUrl.Marshal()
	if err != nil {
		return nil, err
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.cache[shortUrl.GetId()] = shortUrl
	err = m.leveldb.Put([]byte(shortUrl.GetId()), shortUrlStr, nil)

	return shortUrl, err
}

func (m *defaultUrlManager) getShortUrlFromStore(key string) (urls.ShortUrl, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	shortUrl, ok := m.cache[key]

	if ok {
		return shortUrl, nil
	}

	val, err := m.leveldb.Get([]byte(key), nil)
	shortUrl = nil
	if err == nil {
		shortUrl = urls.NewDefaultShortUrl("", "", time.Second, time.Now())
		shortUrl.Unmarshal([]byte(val))
	}

	return shortUrl, err
}

func (m *defaultUrlManager) AddCallToCacheAndDb(shortUrl urls.ShortUrl) {
	m.lock.Lock()
	defer m.lock.Unlock()
	shortUrl.AddCall(time.Now())

	// update cache with new value
	m.cache[shortUrl.GetId()] = shortUrl

	shortUrlStr, err := shortUrl.Marshal()

	if err != nil {
		m.logger.Error("manager.go: failed to save update shortUrl to db", zap.Error(err))
		return
	}

	err = m.leveldb.Put([]byte(shortUrl.GetId()), shortUrlStr, nil)

	if err != nil {
		m.logger.Error("manager.go: failed to save updated shortUrl to db", zap.Error(err))
	}
}
