package shortener

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockUrlManager struct {}
func (m *mockUrlManager) CreateUrlHandleFunc(w http.ResponseWriter, r *http.Request) {}
func (m *mockUrlManager) DeleteUrlHandleFunc(w http.ResponseWriter, r *http.Request) {}
func (m *mockUrlManager) GetUrlHandleFunc(w http.ResponseWriter, r *http.Request) {}
func (m *mockUrlManager) Start(ctx context.Context) error {return nil}
func (m *mockUrlManager) End() {}

func TestBasicServer(t *testing.T) {
	server := NewServer(&mockUrlManager{}, zap.NewNop(), "3131") 
	assert.NotNil(t, server)

	go func() {
		err := server.Serve()
		assert.NoError(t, err)
	}()

	time.Sleep(5*time.Second)
	err := server.Shutdown()
	assert.NoError(t, err)
}