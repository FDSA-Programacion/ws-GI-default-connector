package provider

import (
	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/infrastructure/config"
)

// ===================================================================
// FUNCIONES PRINCIPALES - RQ (Domain -> XML)
// ===================================================================
func GIBookRequestToProvider(req *domain.BookRequest, cfg config.ProviderConfig) ProviderBookRequest {
	return ProviderBookRequest{}
}

// ===================================================================
// FUNCIONES PRINCIPALES - RS (XML -> Domain)
// ===================================================================
func ProviderBookResponseToGI(graphqlResp *ProviderBookResponse, req *domain.BookRequest) domain.BaseJsonRS[*domain.BookResponse] {
	return domain.BaseJsonRS[*domain.BookResponse]{}
}
