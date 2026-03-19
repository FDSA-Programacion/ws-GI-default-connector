package cache

import (
	"fmt"
	"reflect"
	"time"

	"ws-int-httr/internal/domain"

	gocache "github.com/patrickmn/go-cache"
)

const (
	CacheNameRegimenDic                   = "REGIMENDIC"
	CacheNameRegimenDicSupplier           = "REGIMENDIC_SUPPLIER"
	CacheNameRegimenTx                    = "REGIMENTX"
	CacheNameTipoHabitacionDicPrv         = "TIPOHABITACIONDICPRV"
	CacheNameTipoHabitacionDicPrvSupplier = "TIPOHABITACIONDICPRV_SUPPLIER"
	CacheNameTipoHabitacionTx             = "TIPOHABITACIONTX"
	CacheNameTipoHabitacionTxPrv          = "TIPOHABITACIONTXPRV"
	CacheNameAlojamiento                  = "ALOJAMIENTO_QUERY"
	CacheNameAlojamientoDic               = "ALOJAMIENTO_DIC"
	CacheNameAlojamientoPrvDic            = "ALOJAMIENTO_PRV_DIC"
	CacheNameAlojamientoDicByCity         = "ALOJAMIENTO_DIC_CITY"
	CacheNameAlojamientoDicByArea         = "ALOJAMIENTO_DIC_AREA"
	CacheNameEmpresaFactDic               = "EMPRESA_FAC_DIC"
	CacheNameEmpresaFactDicSupplier       = "EMPRESA_FAC_PRV_DIC"
)

type cacheServiceImpl struct {
	masterCaches map[string]*gocache.Cache
}

func NewCacheService() domain.CacheService {
	return &cacheServiceImpl{
		masterCaches: make(map[string]*gocache.Cache),
	}
}

func (c *cacheServiceImpl) LoadDataToCache(cacheName string, data map[string]interface{}) {
	tmpCache := gocache.New(gocache.NoExpiration, gocache.NoExpiration)

	for k, v := range data {
		val := reflect.ValueOf(v)

		if val.Kind() == reflect.Ptr {
			tmpCache.Set(k, v, gocache.NoExpiration)
		} else if val.CanAddr() {
			tmpCache.Set(k, val.Addr().Interface(), gocache.NoExpiration)
		} else {
			tmpCache.Set(k, v, gocache.NoExpiration)
		}
		// tmpCache.Set(k, v, gocache.NoExpiration)
	}
	c.masterCaches[cacheName] = tmpCache
}

func (c *cacheServiceImpl) IsCacheOK() bool {
	if c.masterCaches[CacheNameRegimenDic] != nil &&
		c.masterCaches[CacheNameRegimenDicSupplier] != nil &&
		c.masterCaches[CacheNameRegimenTx] != nil &&
		c.masterCaches[CacheNameTipoHabitacionDicPrv] != nil &&
		c.masterCaches[CacheNameTipoHabitacionDicPrvSupplier] != nil &&
		c.masterCaches[CacheNameTipoHabitacionTx] != nil &&
		c.masterCaches[CacheNameTipoHabitacionTxPrv] != nil &&
		c.masterCaches[CacheNameAlojamiento] != nil {
		return true
	}
	return false
}

func (c *cacheServiceImpl) From(cacheName string) interface{} {
	cacheInstance, ok := c.masterCaches[cacheName]
	if !ok {
		return nil
	}

	return cacheInstance
}

func (c *cacheServiceImpl) Get(cacheName string, key string) (interface{}, bool) {
	cacheInstance, ok := c.masterCaches[cacheName]
	if !ok {
		return nil, false
	}
	return cacheInstance.Get(key)
}

func (c *cacheServiceImpl) Set(cacheName string, key string, value interface{}, duration time.Duration) {
	cacheInstance, ok := c.masterCaches[cacheName]
	if !ok {
		return
	}
	cacheInstance.Set(key, value, duration)
}

func (c *cacheServiceImpl) GetCacheItems(cacheName string) (map[string]interface{}, error) {
	cacheInstance, ok := c.masterCaches[cacheName]
	if !ok {
		return nil, fmt.Errorf("cache no encontrada: %s", cacheName)
	}
	items := cacheInstance.Items()

	dataOutput := make(map[string]interface{}, len(items))
	for k, v := range items {
		dataOutput[k] = v.Object
	}

	return dataOutput, nil
}
