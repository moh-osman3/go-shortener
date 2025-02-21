package urls

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShortUrl(t *testing.T) {
	id := "hashid"
	longUrl := "www.longurl.com"
	expiry := 5 * time.Minute
	timestamp := time.Now()
	surl := NewDefaultShortUrl(id, longUrl, expiry, timestamp)

	surl.AddCall(time.Now())
	expectedSummary := fmt.Sprintf("Summary of shorturl:\n calls in the last day: %d calls\n calls in the last week: %d calls\n total calls since creation: %d calls\n", 1, 1, 1)
	assert.Equal(t, expectedSummary, surl.GetSummary())

	assert.Equal(t, id, surl.GetId())
	assert.Equal(t, longUrl, surl.GetLongUrl())
	assert.Equal(t, timestamp.Add(expiry), surl.GetExpiry())
}

func TestExpiry(t *testing.T) {
	timestamp := time.Now()
	id := "hashid"
	longUrl := "www.longurl.com"
	expiry := 5 * time.Minute
	surl := NewDefaultShortUrl(id, longUrl, expiry, timestamp)
	assert.Equal(t, timestamp.Add(expiry), surl.GetExpiry())

	timestamp = time.Now()
	expiry = -1 * time.Minute
	surl = NewDefaultShortUrl(id, longUrl, expiry, timestamp) 
	assert.Equal(t, time.Time{}, surl.GetExpiry())

	timestamp = time.Now()
	expiry = 0 * time.Minute
	surl = NewDefaultShortUrl(id, longUrl, expiry, timestamp) 
	assert.Equal(t, timestamp.AddDate(1, 0, 0), surl.GetExpiry())
}

func TestMarshal(t *testing.T) {
	timestamp := time.Now()
	id := "hashid"
	longUrl := "www.longurl.com"
	expiry := 5 * time.Minute
	surl := NewDefaultShortUrl(id, longUrl, expiry, timestamp)
	surl.AddCall(time.Now())

	out, err := surl.Marshal()
	assert.NoError(t, err)

	unmarshaledSurl := NewDefaultShortUrl("", "", time.Second, time.Now())
	err = unmarshaledSurl.Unmarshal(out)
	assert.NoError(t, err)
	assert.Equal(t, surl.GetId(), unmarshaledSurl.GetId())
	assert.Equal(t, surl.GetLongUrl(), unmarshaledSurl.GetLongUrl())
	assert.Equal(t, surl.GetExpiry().Unix(), unmarshaledSurl.GetExpiry().Unix())
	assert.Equal(t, surl.GetSummary(), unmarshaledSurl.GetSummary())
}
