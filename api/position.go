package api

import (
	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func HandlePositionsList(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.GetOpenPositions(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res...)
}
