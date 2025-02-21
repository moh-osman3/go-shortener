package shortener

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockUrlManager struct{}

func (m *mockUrlManager) CreateUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	return
}
func (m *mockUrlManager) DeleteUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	return
}
func (m *mockUrlManager) GetUrlHandleFunc(w http.ResponseWriter, r *http.Request) {
	return
}
func (m *mockUrlManager) Start(ctx context.Context, cacheInterval time.Duration, dbInterval time.Duration) error {
	return nil
}
func (m *mockUrlManager) End() {
	return
}

func TestBasicServer(t *testing.T) {
	server := NewServer(&mockUrlManager{}, zap.NewNop(), "3131")
	assert.NotNil(t, server)

	errs := make(chan error, 1)
	go func() {
		errs <- server.Serve()
	}()

	time.Sleep(5 * time.Second)
	err := server.Shutdown()
	assert.NoError(t, err)
	err = <-errs
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Server closed")
}
