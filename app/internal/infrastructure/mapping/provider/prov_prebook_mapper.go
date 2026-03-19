package provider

import (
	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/infrastructure/config"
)

// ===================================================================
// FUNCIONES PRINCIPALES - RQ (Domain -> XML)
// ===================================================================
func GIPrebookRequestToProvider(domainReq *domain.PreBookRequest, cfg config.ProviderConfig) ProviderPrebookRequest {
	return ProviderPrebookRequest{}

}

// ===================================================================
// FUNCIONES PRINCIPALES - RS (XML -> Domain)
// ===================================================================
func ProviderPrebookResponseToGI(providerPrebookRS *ProviderPrebookResponse, req *domain.PreBookRequest) domain.BaseJsonRS[*domain.PreBookResponse] {
	return domain.BaseJsonRS[*domain.PreBookResponse]{}

}
