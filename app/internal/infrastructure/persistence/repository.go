package persistence

import (
	"strconv"
	"strings"

	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/infrastructure/persistence/cache"
	"ws-int-httr/internal/infrastructure/persistence/orm"

	gocache "github.com/patrickmn/go-cache"
)

type RepositoryService interface {
	GetBoardFromExternalCode(boardCode string) *orm.DBRegimen
	GIHotelCodeToExternalCode(hotelCode string, supplierId int) string
	GiHotelFromExternalCode(hotelCode string, supplierId int) *orm.DBAlojamiento
	GICityCodeToExternalCode(cityCode string, supplierId int) []string
	GIAreaIdToExternalCode(areaCode string, supplierId int) []string
	GetRoomFromExternalCode(prvRoomCode, prvRoomName, primaryLangID string, supplierID int) *orm.DBRoomMapping
	GetRoomFromCodHabitacion(roomCode string, PrimaryLangID string) *orm.DBRoomDescription
	GetCacheService() domain.CacheService
}

type repositoryServiceImpl struct {
	cache domain.CacheService
}

func NewRepositoryService(cacheService domain.CacheService) RepositoryService {
	return &repositoryServiceImpl{
		cache: cacheService,
	}
}

func (r *repositoryServiceImpl) GetBoardFromExternalCode(boardCode string) *orm.DBRegimen {
	key := boardCode
	val, ok := r.cache.Get(cache.CacheNameRegimenDicSupplier, key)
	if !ok {
		return nil
	}

	return val.(*orm.DBRegimen)
}

func (r *repositoryServiceImpl) GIHotelCodeToExternalCode(hotelCode string, supplierId int) string {
	key := hotelCode + "|" + strconv.Itoa(supplierId)
	val, ok := r.cache.Get(cache.CacheNameAlojamientoDic, key)
	if !ok {
		return ""
	}
	return val.(orm.DBAlojamiento).ProviderHotelID
}

func (r *repositoryServiceImpl) GiHotelFromExternalCode(hotelCode string, supplierId int) *orm.DBAlojamiento {
	key := hotelCode + "|" + strconv.Itoa(supplierId)
	val, ok := r.cache.Get(cache.CacheNameAlojamientoPrvDic, key)
	if !ok {
		return nil
	}

	if aloj, ok := val.(*orm.DBAlojamiento); ok {
		return aloj
	}
	if aloj, ok := val.(orm.DBAlojamiento); ok {
		copy := aloj
		return &copy
	}
	return nil
}

func (r *repositoryServiceImpl) GICityCodeToExternalCode(cityCode string, supplierId int) []string {
	key := cityCode + "|" + strconv.Itoa(supplierId)
	val, ok := r.cache.Get(cache.CacheNameAlojamientoDicByCity, key)
	if !ok {
		return []string{}
	}

	ormObj := val.([]*orm.DBAlojamiento)

	rs := make([]string, len(ormObj))
	for i := 0; i < len(ormObj); i++ {
		rs[i] = ormObj[i].ProviderHotelID
	}
	return rs
}

func (r *repositoryServiceImpl) GIAreaIdToExternalCode(areaCode string, supplierId int) []string {
	key := areaCode + "|" + strconv.Itoa(supplierId)
	val, ok := r.cache.Get(cache.CacheNameAlojamientoDicByArea, key)
	if !ok {
		return []string{}
	}

	ormObj := val.([]*orm.DBAlojamiento)

	rs := make([]string, len(ormObj))
	for i := 0; i < len(ormObj); i++ {
		rs[i] = ormObj[i].ProviderHotelID
	}
	return rs
}

func (r *repositoryServiceImpl) GetRoomFromExternalCode(prvRoomCode, prvRoomName, primaryLangID string, supplierID int) *orm.DBRoomMapping {
	var room = &orm.DBRoomMapping{PrvRoomName: prvRoomName}

	if len(prvRoomCode) <= 0 {
		room.PrvRoomCode = strings.Replace(strings.ToUpper(prvRoomName), " ", "_", -1)
	} else {
		room.PrvRoomCode = prvRoomCode
	}
	key := prvRoomCode + "|" + strconv.Itoa(supplierID)

	obj, ok := r.cache.Get(cache.CacheNameTipoHabitacionDicPrvSupplier, key)

	if ok {
		th := obj.(orm.DBRoomMapping)

		room.GIRoomID = th.GIRoomID
		room.GIRoomCode = th.GIRoomCode
	}
	return room

}

func (r *repositoryServiceImpl) GetRoomFromCodHabitacion(roomCode string, PrimaryLangID string) *orm.DBRoomDescription {
	cacheInstance, ok := r.cache.From(cache.CacheNameTipoHabitacionTx).(*gocache.Cache)
	if !ok {
		return nil
	}

	items := cacheInstance.Items()
	for _, v := range items {
		obj, ok := v.Object.(orm.DBRoomDescription)
		if !ok {
			continue
		}

		if obj.CodHabitacion == roomCode && obj.CodIdioma == PrimaryLangID {
			return &obj
		}
	}

	return nil
}

func (r *repositoryServiceImpl) GetCacheService() domain.CacheService {
	return r.cache
}
