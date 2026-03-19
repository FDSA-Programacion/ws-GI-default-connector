package persistence

import (
	"database/sql"
	"log"
	"strconv"
	"ws-int-httr/internal/infrastructure/config"

	"github.com/godror/godror"
)

var db *sql.DB
var conf config.DBConfig

func InitDB(config config.DBConfig) error {
	var err error

	conf = config

	dbChain := config.DBHost() + ":" + strconv.Itoa(config.DBPort()) + "/" + config.DBSID()
	db, err = sql.Open(config.DBDriver(), `user="`+config.DBUser()+`" password="`+config.DBPass()+`" connectString="`+dbChain+`"`)

	if err != nil {
		return err
	}

	return db.Ping()
}

func GenericQuery(query string) (map[string]interface{}, error) {
	rows, err := db.Query(query, godror.FetchArraySize(conf.DBFechRowCount()))

	if err != nil {
		log.Println("Error al obtener las filas:", err)
		return make(map[string]interface{}), err
	}

	cols, err := rows.Columns()
	if err != nil {
		log.Println("Error al obtener las columnas:", err)
		return make(map[string]interface{}), err
	}

	ccc := make(map[string]interface{})

	rowCount := 0
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return make(map[string]interface{}), err
		}

		newMap := make(map[string]interface{})
		for i, v := range columns {
			newMap[cols[i]] = v
		}

		ccc[strconv.Itoa(rowCount)] = newMap
		rowCount++
	}

	log.Printf("El nº de rows es: %v", rowCount)

	return ccc, nil
}
