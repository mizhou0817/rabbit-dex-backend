package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type WriteStorageRequest struct {
	Data []byte `json:"data" binding:"required"`
}

func HandleReadStorage(c *gin.Context) {
	ctx := GetRabbitContext(c)
	if !IsAllowedProfileId(ctx.ProfileIdFromJwt) {
		ErrorResponse(c, errors.New("BROKEN_PROFILE_ID"))
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)

	data, err := apiModel.ReadFrontendStorage(c.Request.Context(), ctx.ProfileIdFromJwt)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	// Do nothing in other way
	SuccessResponse(c, data)
}

func HandleWriteStorage(c *gin.Context) {
	var request WriteStorageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	if !IsAllowedProfileId(ctx.ProfileIdFromJwt) {
		ErrorResponse(c, errors.New("BROKEN_PROFILE_ID"))
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)

	err := apiModel.WriteFrontendStorage(c.Request.Context(), ctx.ProfileIdFromJwt, request.Data)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	// Do nothing in other way
	SuccessResponse(c, true)
}
