package handlers

import (
	"errors"
	"testing"

	"ws-int-httr/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAvailErrorResponse_IncludesDetailedErrorInErrorsType(t *testing.T) {
	detailErr := errors.New("Wrong status on booking response PRODUCT_ERROR")
	customErr := domain.ErrorAvailCode
	customErr.Err = detailErr

	resp := createAvailErrorResponse(&customErr)

	require.NotNil(t, resp.Errors)
	require.Len(t, resp.Errors.ErrorsType, 1)
	assert.Equal(t, customErr.Message, resp.Errors.ErrorsType[0].Error)
	assert.Equal(t, 400, resp.Errors.ErrorsType[0].Type)
	require.NotNil(t, resp.InternalCondition)
	assert.Equal(t, detailErr.Error(), resp.InternalCondition.ProviderStatusDescription)
}

func TestCreatePreBookErrorResponse_IncludesDetailedErrorInErrorsType(t *testing.T) {
	detailErr := errors.New("Wrong status on booking response PRODUCT_ERROR")
	customErr := domain.ErrorPrebookCode
	customErr.Err = detailErr

	resp := createPreBookErrorResponse(&customErr)

	require.NotNil(t, resp.Errors)
	require.Len(t, resp.Errors.ErrorsType, 1)
	assert.Equal(t, customErr.Message, resp.Errors.ErrorsType[0].Error)
	assert.Equal(t, 400, resp.Errors.ErrorsType[0].Type)
	require.NotNil(t, resp.InternalCondition)
	assert.Equal(t, detailErr.Error(), resp.InternalCondition.ProviderStatusDescription)
}

func TestCreateBookErrorResponse_IncludesDetailedErrorInErrorsType(t *testing.T) {
	detailErr := errors.New("Wrong status on booking response PRODUCT_ERROR")
	customErr := domain.ErrorBookCode
	customErr.Err = detailErr

	resp := createBookErrorResponse(&customErr)

	require.NotNil(t, resp.Errors)
	require.Len(t, resp.Errors.ErrorsType, 1)
	assert.Equal(t, customErr.Message, resp.Errors.ErrorsType[0].Error)
	assert.Equal(t, 400, resp.Errors.ErrorsType[0].Type)
	require.NotNil(t, resp.InternalCondition)
	require.NotNil(t, resp.InternalCondition.ProviderStatusDescription)
	assert.Equal(t, detailErr.Error(), *resp.InternalCondition.ProviderStatusDescription)
}

func TestCreateCancelErrorResponse_IncludesDetailedErrorInErrorsType(t *testing.T) {
	detailErr := errors.New("Cancel rejected by provider BAD_STATUS")
	customErr := domain.ErrorCancelCode
	customErr.Err = detailErr

	resp := createCancelErrorResponse(&customErr)

	require.NotNil(t, resp.Errors)
	require.Len(t, resp.Errors.ErrorsType, 1)
	assert.Equal(t, customErr.Message, resp.Errors.ErrorsType[0].Error)
	assert.Equal(t, 400, resp.Errors.ErrorsType[0].Type)
	require.NotNil(t, resp.InternalCondition)
	assert.Equal(t, detailErr.Error(), resp.InternalCondition.ProviderStatusDescription)
}

func TestGetErrorDescription_FallsBackToMessage(t *testing.T) {
	ce := domain.CustomError{Message: "plain message"}
	require.Equal(t, "plain message", getErrorDescription(&ce))
}
