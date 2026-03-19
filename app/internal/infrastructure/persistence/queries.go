package persistence

import (
	"fmt"
	"strings"
)

func trasnformIntArrToStringArr(providerIds ...int) []string {
	idStrings := make([]string, len(providerIds))
	for i, id := range providerIds {
		idStrings[i] = fmt.Sprintf("%d", id)
	}
	return idStrings
}

func GetRegimenDicQuery(providerIds ...int) string {
	idStrings := trasnformIntArrToStringArr(providerIds...)

	return `
		SELECT DISTINCT
			to_char(pr.id) AS id,
			to_char(r.id)  AS regimenid,
			r.codigo       AS codigo,
			pr.codext      AS provider_code,
	    r.descripcion  AS descripcion
		FROM
				regimen           r
				LEFT JOIN proveedor_regimen pr ON ( pr.regimen_id = r.id )
		WHERE
				pr.proveedor_id in (` + strings.Join(idStrings, ",") + `)
		`
}

func GetRegimenTxQuery(_ ...int) string {
	return `
		SELECT 
			CONCAT(CONCAT(TO_CHAR(r.id), '|'), 'ES') AS key, 
			TO_CHAR(r.id) AS idregimen, 
			r.codigo AS codregimen, 
			'ES' AS codidioma, 
			r.nombre AS descripcion 
		FROM regimen r 
		
		UNION 
		
		SELECT DISTINCT 
			CONCAT(CONCAT(TO_CHAR(r.id), '|'), idi.codigo) AS key, 
			TO_CHAR(r.id) AS idregimen, 
			r.codigo AS codregimen, 
			idi.codigo AS codidioma, 
			tra.nombre AS descripcion 
		FROM regimen r 
		INNER JOIN traducciones tra 
			ON tra.codigo = r.codigo 
		INNER JOIN idioma idi 
			ON tra.idioma_id = idi.id 
		WHERE tra.tipo_traduccion_id = 3 
		ORDER BY idregimen
	`
}

func GetTipoHabitacionDicPrvQuery(providerIds ...int) string {
	idStrings := trasnformIntArrToStringArr(providerIds...)

	return `
		SELECT 
			TO_CHAR(pm.id) AS identifier, 
			TO_CHAR(m.id) AS giroomid, 
			TO_CHAR(m.codigo) AS giroomcode, 
			TO_CHAR(m.nombre) AS giroomname, 
			pm.codext AS prvroomcode, 
			pm.nombre AS prvroomname, 
			TO_CHAR(i.id) AS integrationid, 
			i.codigo AS integrationcode 
		FROM proveedor_modalidad pm 
		LEFT JOIN modalidad m 
			ON pm.modalidad_id = m.id 
		INNER JOIN integracion i 
			ON pm.proveedor_id = i.proveedor_id 
		WHERE pm.activo = 'S' 
			AND m.activo = 'S' 
			AND pm.proveedor_id IN (` + strings.Join(idStrings, ",") + `) 
		ORDER BY m.id ASC
	`
}

func GetTipoHabitacionTxQuery(_ ...int) string {
	return ` 
		SELECT 
			CONCAT(CONCAT(TO_CHAR(m.ID), '|'), idi.CODIGO) AS key, 
			TO_CHAR(m.ID) AS identifier, 
			m.CODIGO AS codHabitacion, 
			tra.NOMBRE AS descripcion, 
			idi.CODIGO AS codIdioma 
		FROM MODALIDAD m 
		INNER JOIN TRADUCCIONES tra 
			ON (m.CODIGO = tra.CODIGO AND tra.TIPO_TRADUCCION_ID = 1) 
		INNER JOIN IDIOMA idi 
			ON tra.idioma_id = idi.id 
	`
}

func GetAlojamientoQuery(providerIds ...int) string {
	idStrings := trasnformIntArrToStringArr(providerIds...)

	return `
		SELECT DISTINCT 
			TO_CHAR(ser.id) AS hotelID, 
			ser.codigo AS hotelCode, 
			ser.nombre AS hotelName, 
			i.id AS integrationId, 
			i.codigo AS integrationCode, 
			ps.codext AS providerHotelID, 
			zonp.codigo AS areaCode, 
			zonp.nombre AS areaName, 
			zonh.id AS cityID, 
			zonh.nombre AS cityName, 
			cat.codigo AS rating, 
			est.nombre AS category, 
			CASE est.id 
				WHEN 1 THEN '20'
				WHEN 6 THEN '3'
				ELSE ''
			END AS propertyclasscode 
		FROM servicio ser 
		INNER JOIN zona zonh 
			ON ser.zona_id = zonh.id 
		LEFT JOIN zona zonp 
			ON zonh.zona_id = zonp.id 
		INNER JOIN proveedor_servicio ps 
			ON ser.id = ps.servicio_id 
			AND ps.activo = 'S'
		INNER JOIN integracion i 
			ON ps.proveedor_id = i.proveedor_id 
			AND i.proveedor_id IN (` + strings.Join(idStrings, ",") + `) 
		LEFT JOIN categoria cat 
			ON ser.categoria_id = cat.id 
		LEFT JOIN establecimiento est 
			ON ser.establecimiento_id = est.id 
		WHERE ser.tipo_servicio_id = 1 
			AND ser.activo = 'S' 
		ORDER BY zonp.codigo, zonh.id
	`
}

func GetIntegracionDicQuery(providerIds ...int) string {
	idStrings := trasnformIntArrToStringArr(providerIds...)

	return `
		SELECT 
			TO_CHAR(i.ID) AS id, 
			i.codigo AS code, 
			i.observacion AS name 
		FROM INTEGRACION i 
		INNER JOIN PROVEEDOR p 
			ON i.proveedor_id = p.id 
		WHERE p.id IN (` + strings.Join(idStrings, ",") + `)
	`
}
