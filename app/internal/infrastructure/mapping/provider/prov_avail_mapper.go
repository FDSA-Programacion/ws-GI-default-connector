package provider

import (
	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/infrastructure/config"
)

// ===================================================================
// FUNCIONES PRINCIPALES - RQ (Domain -> XML)
// ===================================================================
func GIAvailRequestToProvider(giDomainReq *domain.AvailRequest, cfg config.ProviderConfig) ProviderAvailRequest {
	return ProviderAvailRequest{}
}

// ===================================================================
// FUNCIONES PRINCIPALES - RS (XML -> DOMAIN)
// ===================================================================
func ProviderAvailResponseToGI(graphqlResp *ProviderAvailResponse, req *domain.AvailRequest) domain.BaseJsonRS[*domain.AvailResponse] {
	return domain.BaseJsonRS[*domain.AvailResponse]{}
}
