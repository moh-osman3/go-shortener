package urls

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ShortUrl interface {
	GetId() string
	GetLongUrl() string
	GetExpiry() time.Time
	AddCall(timestamp time.Time)
	GetSummary() string
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type defaultShortUrl struct {
	// export these fields for json marshaling
	Id string `json:"id"` 
	LongUrl string `json:"long_url"`
	Expiry time.Time `json:"expiry"`
	CreationTime time.Time `json:"creation_time"`
	Counter *Counter `json:"counter"`
}

func (su *defaultShortUrl) Marshal() ([]byte, error) {
	return json.Marshal(su)
}

func (su *defaultShortUrl) Unmarshal(data []byte) (error) {
	return json.Unmarshal(data, su)
}

func (su *defaultShortUrl) AddCall(timestamp time.Time) {
	su.Counter.AddCall(timestamp)
}

func (su *defaultShortUrl) GetSummary() string {
	return su.Counter.GetSummary()
}

func NewDefaultShortUrl(id string, longUrl string, expiry time.Duration, timestamp time.Time) ShortUrl {
	su := &defaultShortUrl{
		Id: id,
		LongUrl: longUrl,
		CreationTime: timestamp,
		Counter: NewCounter(),
	}

	if expiry == 0 {
		// default behavior
		su.Expiry = timestamp.AddDate(1, 0, 0)
	} else if expiry < 0 {
		su.Expiry = time.Time{}
	} else {
		su.Expiry = timestamp.Add(expiry)
	}
	return su
}

func (su *defaultShortUrl) GetId() string {
	return su.Id
}

func (su *defaultShortUrl) GetExpiry() time.Time {
	return su.Expiry
}

func (su *defaultShortUrl) GetLongUrl() string {
	return su.LongUrl
}