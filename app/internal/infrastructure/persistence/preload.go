package persistence

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/infrastructure/persistence/cache"
	"ws-int-httr/internal/infrastructure/persistence/orm"

	"github.com/mitchellh/mapstructure"
)

type CacheLoaderConfig struct {
	Name    string
	QueryFn func(providerIDs ...int) string
	SaveFn  func(data map[string]interface{}, cacheService domain.CacheService)
	TypeORM reflect.Type
}

func saveRegimenDic(data map[string]interface{}, cacheService domain.CacheService) {
	dataOutput := make(map[string]interface{})

	for _, v := range data {
		obj := orm.DBRegimen{}

		if err := mapstructure.Decode(v, &obj); err != nil {
			continue
		}

		key := obj.ProviderCode

		dataOutput[key] = &obj
	}

	cacheService.LoadDataToCache(cache.CacheNameRegimenDicSupplier, dataOutput)
}

func saveTipoHabitacionDicPrv(data map[string]interface{}, cacheService domain.CacheService) {
	dataOutput := make(map[string]interface{})

	for _, v := range data {
		obj := orm.DBRoomMapping{}

		if err := mapstructure.Decode(v, &obj); err != nil {
			continue
		}

		key := obj.PrvRoomCode + "|" + obj.IntegrationID

		dataOutput[key] = obj
	}

	cacheService.LoadDataToCache(cache.CacheNameTipoHabitacionDicPrvSupplier, dataOutput)
}

func saveTipoHabitacionTx(data map[string]interface{}, cacheService domain.CacheService) {
	dataOutput := make(map[string]interface{})

	for _, v := range data {
		obj := orm.DBRoomDescription{}

		if err := mapstructure.Decode(v, &obj); err != nil {
			continue
		}

		key := obj.Key

		dataOutput[key] = obj
	}

	cacheService.LoadDataToCache(cache.CacheNameTipoHabitacionTx, dataOutput)
}

func saveAlojamiento(data map[string]interface{}, cacheService domain.CacheService) {
	dataDicOutput := make(map[string]interface{})
	dataPrvOutput := make(map[string]interface{})
	dataGiAccomCity := make(map[string]interface{})
	dataGiAccomArea := make(map[string]interface{})

	for _, v := range data {
		obj := orm.DBAlojamiento{}

		if err := mapstructure.Decode(v, &obj); err != nil {
			continue
		}

		keyDic := obj.HotelID + "|" + strconv.Itoa(obj.IntegrationID)
		keyPrv := obj.ProviderHotelID + "|" + strconv.Itoa(obj.IntegrationID)
		keyDicByCity := strconv.Itoa(obj.CityID) + "|" + strconv.Itoa(obj.IntegrationID)
		keyDicByArea := obj.AreaCode + "|" + strconv.Itoa(obj.IntegrationID)

		dataDicOutput[keyDic] = obj
		dataPrvOutput[keyPrv] = obj

		if dataGiAccomCity[keyDicByCity] != nil {
			als := dataGiAccomCity[keyDicByCity].([]*orm.DBAlojamiento)
			als = append(als, &obj)
			dataGiAccomCity[keyDicByCity] = als
		} else {
			als := []*orm.DBAlojamiento{&obj}
			dataGiAccomCity[keyDicByCity] = als
		}
		if dataGiAccomArea[keyDicByArea] != nil {
			als := dataGiAccomArea[keyDicByArea].([]*orm.DBAlojamiento)
			als = append(als, &obj)
			dataGiAccomArea[keyDicByArea] = als
		} else {
			als := []*orm.DBAlojamiento{&obj}
			dataGiAccomArea[keyDicByArea] = als
		}
	}

	cacheService.LoadDataToCache(cache.CacheNameAlojamientoDic, dataDicOutput)
	cacheService.LoadDataToCache(cache.CacheNameAlojamientoPrvDic, dataPrvOutput)
	cacheService.LoadDataToCache(cache.CacheNameAlojamientoDicByCity, dataGiAccomCity)
	cacheService.LoadDataToCache(cache.CacheNameAlojamientoDicByArea, dataGiAccomArea)
}

func saveIntegracionDic(data map[string]interface{}, cacheService domain.CacheService) {
	dataOutput := make(map[string]interface{})

	for _, v := range data {
		obj := orm.DBRoomDescription{}

		if err := mapstructure.Decode(v, &obj); err != nil {
			continue
		}

		key := obj.Key

		dataOutput[key] = obj
	}

	cacheService.LoadDataToCache(cache.CacheNameTipoHabitacionTx, dataOutput)
}

func RefreshCache(cacheService domain.CacheService, providerIds ...int) {
	fmt.Println("Refreshing cache......")
	LoadAllCache(cacheService, providerIds...)
}

func LoadAllCache(cacheService domain.CacheService, providerIds ...int) {

	cachesToLoad := []CacheLoaderConfig{
		{
			Name:    "RegimenDic",
			QueryFn: GetRegimenDicQuery,
			SaveFn:  saveRegimenDic,
		},
		{
			Name:    "TipoHabitacionDicPrv",
			QueryFn: GetTipoHabitacionDicPrvQuery,
			SaveFn:  saveTipoHabitacionDicPrv,
		},
		{
			Name:    "TipoHabitacionTx",
			QueryFn: GetTipoHabitacionTxQuery,
			SaveFn:  saveTipoHabitacionTx,
		},
		{
			Name:    "Alojamiento",
			QueryFn: GetAlojamientoQuery,
			SaveFn:  saveAlojamiento,
		},
		{
			Name:    "IntegracionDic",
			QueryFn: GetIntegracionDicQuery,
			SaveFn:  saveIntegracionDic,
		},
	}

	for _, c := range cachesToLoad {
		log.Println("Inicia la carga de datos de " + c.Name)
		query := c.QueryFn(providerIds...)

		data, err := GenericQuery(query)
		if err != nil {
			log.Println("error on loadCache().", err)
			continue
		}
		log.Println("Se han registrado " + strconv.Itoa(len(data)) + " registros de " + c.Name)

		c.SaveFn(data, cacheService)
	}

	log.Println("Datos de la Base de datos cargados correctamente.")
}
