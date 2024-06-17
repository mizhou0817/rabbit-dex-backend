package api

import (
	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/vaultdata"
)

func HandleNavHistory(c *gin.Context) {
	var request vaultdata.VaultHistoryRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	response, err := vaultdata.HandleNavHistory(c.Request.Context(), ctx.TimeScaleDB, request, ctx.ExchangeId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, response...)
}

func HandleVault(c *gin.Context) {
	var request vaultdata.VaultRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	response, pagination, err := vaultdata.HandleVault(c.Request.Context(), ctx.TimeScaleDB, request, ctx.ExchangeId, ctx.Pagination)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponsePaginated(c, pagination, response...)
}

func HandleVaultBalanceOperations(c *gin.Context) {
	var request vaultdata.VaultBalanceOperationsRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	response, pagination, err := vaultdata.HandleVaultBalanceOperations(c.Request.Context(), ctx.TimeScaleDB, request, ctx.Profile.ProfileId, ctx.ExchangeId, ctx.Pagination)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponsePaginated(c, pagination, response...)
}

func HandleVaultHoldings(c *gin.Context) {
	var request vaultdata.VaultHoldingsRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	response, pagination, err := vaultdata.HandleVaultHoldings(c.Request.Context(), ctx.TimeScaleDB, request, ctx.Profile.ProfileId, ctx.ExchangeId, ctx.Pagination)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponsePaginated(c, pagination, response...)
}
