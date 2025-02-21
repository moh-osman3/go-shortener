package def

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/moh-osman3/shortener/urls"
)

type errorReader struct {
}

func (er *errorReader) Read(b []byte) (n int, err error) {
	return 0, errors.New("failed reading")
}

func TestDeleteUrlHandleFunc(t *testing.T) {
	m := &defaultUrlManager{
		cache:   make(map[string]urls.ShortUrl),
		logger:  zap.NewNop(),
		leveldb: NewMockDB(),
	}
	// test bad method
	req, err := http.NewRequest(http.MethodGet, "/delete", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(m.DeleteUrlHandleFunc)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	// test io.ReadAll(req.body) return error
	req, err = http.NewRequest(http.MethodDelete, "/delete", &errorReader{})
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	// test short url not found for deletion
	req, err = http.NewRequest(http.MethodDelete, "/delete", bytes.NewBuffer([]byte("test body")))
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	testLongUrl := "www.testlongurl.com"
	expiry := 5 * time.Minute
	createdSurl, err := m.createShortUrl(testLongUrl, expiry)
	assert.NoError(t, err)
	assert.NotNil(t, createdSurl)

	// test happy path
	validBody := fmt.Sprintf("{\"id\":\"%s\"}", createdSurl.GetId())
	req, err = http.NewRequest(http.MethodDelete, "/delete", bytes.NewBuffer([]byte(validBody)))
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUrlHandleFunc(t *testing.T) {
	m := &defaultUrlManager{
		cache:   make(map[string]urls.ShortUrl),
		logger:  zap.NewNop(),
		leveldb: NewMockDB(),
	}
	// test bad method
	req, err := http.NewRequest(http.MethodPost, "/", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(m.GetUrlHandleFunc)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	// test bad URL path too short
	req, err = http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.URL.Path = ""

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// test short url not found
	req, err = http.NewRequest(http.MethodGet, "/testid", nil)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// test bad URL path too long
	req, err = http.NewRequest(http.MethodGet, "/test/path/too/long", nil)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	testLongUrl := "www.testlongurl.com"
	expiry := 5 * time.Minute
	createdSurl, err := m.createShortUrl(testLongUrl, expiry)
	assert.NoError(t, err)
	assert.NotNil(t, createdSurl)

	// test happy path get shorurl
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", createdSurl.GetId()), nil)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)

	// test happy path summary
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/%s/summary", createdSurl.GetId()), nil)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, createdSurl.GetSummary(), w.Body.String())

	// test len(paths) == 2 but path[1] is not "summary"
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/%s/brokensummary", createdSurl.GetId()), nil)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUrlHandleFunc(t *testing.T) {
	m := &defaultUrlManager{
		cache:   make(map[string]urls.ShortUrl),
		logger:  zap.NewNop(),
		leveldb: NewMockDB(),
	}
	// test bad method
	req, err := http.NewRequest(http.MethodGet, "/create", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(m.CreateUrlHandleFunc)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	// test io.ReadAll(req.body) return error
	req, err = http.NewRequest(http.MethodPost, "/create", &errorReader{})
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	// test bad expiry
	badData := "{\"url\":\"www.google.com\",\"expiry\":\"wrongtype\"}"
	req, err = http.NewRequest(http.MethodPost, "/create", bytes.NewBuffer([]byte(badData)))
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// test error during create
	data := "{\"url\":\"www.google.com\",\"expiry\":\"10s\"}"
	// don't initialize cache to trigger internal error
	bm := &defaultUrlManager{
		logger:  zap.NewNop(),
		leveldb: NewMockDB(),
	}
	req, err = http.NewRequest(http.MethodPost, "/create", bytes.NewBuffer([]byte(data)))
	require.NoError(t, err)

	w = httptest.NewRecorder()
	badHandler := http.HandlerFunc(bm.CreateUrlHandleFunc)
	badHandler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// test happy path
	req, err = http.NewRequest(http.MethodPost, "/create", bytes.NewBuffer([]byte(data)))
	require.NoError(t, err)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
