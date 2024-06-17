package api

import (
	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type DeadmanCreateRequest struct {
	Timeout uint `form:"timeout" binding:"gt=0,required"` // timeout in milliseconds
}

func HandleDeadmanCreate(c *gin.Context) {
	var request DeadmanCreateRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	response, err := apiModel.DeadmanCreate(c.Request.Context(), ctx.Profile.ProfileId, request.Timeout)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, response)
}

func HandleDeadmanDelete(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	response, err := apiModel.DeadmanDelete(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, response)
}

func HandleDeadmanList(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	response, err := apiModel.DeadmanGet(c.Request.Context(), ctx.Profile.ProfileId)

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, response)
}
