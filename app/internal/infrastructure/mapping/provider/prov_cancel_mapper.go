package provider

import (
	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/infrastructure/config"
)

// ===================================================================
// FUNCIONES PRINCIPALES - RQ (Domain -> XML)
// ===================================================================
func GICancelRequestToProvider(req *domain.CancelRequest, cfg config.ProviderConfig) ProviderCancelRequest {
	return ProviderCancelRequest{}
}

// ===================================================================
// FUNCIONES PRINCIPALES - RS (XML -> Domain)
// ===================================================================
func ProviderCancelResponseToGI(graphqlResp *ProviderCancelResponse, req *domain.CancelRequest) domain.BaseJsonRS[*domain.CancelResponse] {
	return domain.BaseJsonRS[*domain.CancelResponse]{}
}
