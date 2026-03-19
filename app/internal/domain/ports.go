package domain

import (
	"time"
)

type BookingServicer interface {
	Availability(req *AvailRequest) (*BaseJsonRS[*AvailResponse], error)
	PreBook(req *PreBookRequest) (*BaseJsonRS[*PreBookResponse], error)
	Book(req *BookRequest) (*BaseJsonRS[*BookResponse], error)
	Cancel(req *CancelRequest) (*BaseJsonRS[*CancelResponse], error)
}

type BookingProvider interface {
	SendAvail(req *AvailRequest) (*BaseJsonRS[*AvailResponse], error)
	SendPreBook(req *PreBookRequest) (*BaseJsonRS[*PreBookResponse], error)
	SendBook(req *BookRequest) (*BaseJsonRS[*BookResponse], error)
	SendCancel(req *CancelRequest) (*BaseJsonRS[*CancelResponse], error)
}

type MasterDataRepository interface {
	GenericQuery(query string) (map[string]interface{}, error)
	GetQuery(queryID string, providerCode string) string
}

type CacheService interface {
	LoadDataToCache(cacheName string, data map[string]interface{})
	IsCacheOK() bool
	From(cacheName string) interface{}
	Get(cacheName string, key string) (interface{}, bool)
	Set(cacheName string, key string, value interface{}, duration time.Duration)

	GetCacheItems(cacheName string) (map[string]interface{}, error)
}
