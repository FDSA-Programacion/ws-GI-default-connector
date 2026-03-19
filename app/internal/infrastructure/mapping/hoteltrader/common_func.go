package hoteltrader

import (
	"math"
	"strconv"
	"time"
	"ws-int-httr/internal/domain"
)

// convertDateFormat convierte fecha de YYYY-MM-DD a DD/MM/YYYY
func convertDateFromGIToOT(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}

	return date.Format("02/01/2006")
}

// roundTo2Decimals redondea un float64 a 2 decimales
func roundTo2Decimals(value float64) float64 {
	return math.Round(value*100) / 100
}

func parseOffsetSeconds(offset string) int {
	sign := 1
	if offset[0] == '-' {
		sign = -1
	}
	hours, _ := strconv.Atoi(offset[1:3])
	mins, _ := strconv.Atoi(offset[4:6])
	return sign * ((hours * 3600) + (mins * 60))
}

// formatAges formatea las edades de adultos de "3030" a "30,30"
func formatAges(ages string) string {
	if ages == "" {
		return ""
	}
	// Si tiene comas, ya está formateado
	if len(ages) > 0 && len(ages)%2 == 0 {
		// Formatear cada par de dígitos con comas
		result := ""
		for i := 0; i < len(ages); i += 2 {
			if i > 0 {
				result += ","
			}
			if i+1 < len(ages) {
				result += ages[i : i+2]
			} else {
				result += ages[i:]
			}
		}
		return result
	}
	return ages
}

func BuildGIErrorContainer(errorMessage string) *domain.ErrorsContainer {
	if errorMessage == "" {
		return nil
	}

	return &domain.ErrorsContainer{
		ErrorsType: []domain.ErrorType{
			{
				Error: errorMessage,
				Type:  400,
			},
		},
	}
}
