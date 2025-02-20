package def

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.uber.org/zap"

	"github.com/moh-osman3/shortener/urls"
)

type mockDB struct {
	db map[string][]byte
}

func NewMockDB() DB {
	return &mockDB{
		db: make(map[string][]byte),
	}
}

func (mdb *mockDB) Get(key []byte, ro *opt.ReadOptions) ([]byte, error) {
	val, ok := mdb.db[string(key)]
	if !ok {
		return []byte{}, errors.New("value not found in mockdb")
	}

	return val, nil
}

func (mdb *mockDB) Put(key, value []byte, wo *opt.WriteOptions) error {
	mdb.db[string(key)] = value
	return nil
}
func (mdb *mockDB) Delete(key []byte, wo *opt.WriteOptions) error {
	delete(mdb.db, string(key))
	return nil
}
func (mdb *mockDB) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return iterator.NewEmptyIterator(nil)
}

func TestCreateAndGetUrl(t *testing.T) {
	defManager := &defaultUrlManager{
		cache:   make(map[string]urls.ShortUrl),
		logger:  zap.NewNop(),
		leveldb: NewMockDB(),
	}

	testLongUrl := "www.testlongurl.com"
	expiry := 5 * time.Minute
	createdSurl, err := defManager.createShortUrl(testLongUrl, expiry)
	assert.NoError(t, err)
	assert.NotNil(t, createdSurl)
	assert.Equal(t, testLongUrl, createdSurl.GetLongUrl())

	expectedId := createdSurl.GetId()

	fetchedSurl, err := defManager.getShortUrlFromStore(expectedId)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedSurl)
	assert.Equal(t, expectedId, fetchedSurl.GetId())
	assert.Equal(t, createdSurl.GetLongUrl(), fetchedSurl.GetLongUrl())
	assert.Equal(t, createdSurl.GetExpiry(), fetchedSurl.GetExpiry())
	assert.Equal(t, createdSurl.GetSummary(), fetchedSurl.GetSummary())

	// check that both cache and db have the value
	val, ok := defManager.cache[expectedId]
	assert.True(t, ok)
	assert.Equal(t, val, fetchedSurl)

	valStr, err := defManager.leveldb.Get([]byte(expectedId), nil)
	assert.NoError(t, err)
	surl := urls.NewDefaultShortUrl("", "", time.Second, time.Now())
	surl.Unmarshal([]byte(valStr))
	assert.Equal(t, createdSurl.GetLongUrl(), surl.GetLongUrl())
	assert.Equal(t, createdSurl.GetExpiry().Unix(), surl.GetExpiry().Unix())
	assert.Equal(t, createdSurl.GetSummary(), surl.GetSummary())
}

func TestDeleteUrl(t *testing.T) {
	defManager := &defaultUrlManager{
		cache:   make(map[string]urls.ShortUrl),
		logger:  zap.NewNop(),
		leveldb: NewMockDB(),
	}

	testLongUrl := "www.testlongurl.com"
	expiry := 5 * time.Minute
	createdSurl, err := defManager.createShortUrl(testLongUrl, expiry)
	assert.NoError(t, err)
	assert.NotNil(t, createdSurl)
	assert.Equal(t, testLongUrl, createdSurl.GetLongUrl())

	expectedId := createdSurl.GetId()

	err = defManager.deleteKeyFromCacheAndDb(expectedId)
	assert.NoError(t, err)

	fetchedSurl, err := defManager.getShortUrlFromStore(expectedId)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "value not found")
	assert.Nil(t, fetchedSurl)

	// confirm deleted from cache and db
	val, ok := defManager.cache[expectedId]
	assert.False(t, ok)
	assert.Nil(t, val)

	_, err = defManager.leveldb.Get([]byte(expectedId), nil)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "value not found")
}

func TestAddCallToShortUrl(t *testing.T) {
	defManager := &defaultUrlManager{
		cache:   make(map[string]urls.ShortUrl),
		logger:  zap.NewNop(),
		leveldb: NewMockDB(),
	}

	testLongUrl := "www.testlongurl.com"
	expiry := 5 * time.Minute
	createdSurl, err := defManager.createShortUrl(testLongUrl, expiry)
	assert.NoError(t, err)
	assert.NotNil(t, createdSurl)
	assert.Equal(t, testLongUrl, createdSurl.GetLongUrl())

	expectedId := createdSurl.GetId()

	startSummary := createdSurl.GetSummary()

	defManager.AddCallToCacheAndDb(createdSurl)

	fetchedSurl, err := defManager.getShortUrlFromStore(expectedId)

	assert.NoError(t, err)
	assert.NotNil(t, fetchedSurl)
	assert.Equal(t, expectedId, fetchedSurl.GetId())
	assert.Equal(t, createdSurl.GetSummary(), fetchedSurl.GetSummary())

	// confirm instrumentation persisted to cache and db
	cacheSurl, ok := defManager.cache[expectedId]
	assert.True(t, ok)
	assert.NotEqual(t, startSummary, cacheSurl.GetSummary())

	valStr, err := defManager.leveldb.Get([]byte(expectedId), nil)
	assert.NoError(t, err)
	surl := urls.NewDefaultShortUrl("", "", time.Second, time.Now())
	surl.Unmarshal([]byte(valStr))
	assert.NotEqual(t, startSummary, surl.GetSummary())

	assert.Equal(t, cacheSurl.GetSummary(), surl.GetSummary())
}
