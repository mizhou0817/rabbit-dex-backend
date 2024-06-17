package api

import (
	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/portfolio"
)

func HandlePortfolioList(c *gin.Context) {
	var request portfolio.PortfolioRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	response, err := portfolio.HandlePortfolioList(c.Request.Context(), ctx.TimeScaleDB, request, ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, response...)
}
