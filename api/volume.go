package api

import (
	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/volume"
)

func HandleBfxVolume(c *gin.Context) {
	var request volume.VolumeRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	response, err := volume.HandleBfxVolume(c.Request.Context(), ctx.TimeScaleDB, request)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, response)
}

func HandleRbxVolume(c *gin.Context) {
	var request volume.VolumeRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	response, err := volume.HandleRbxVolume(c.Request.Context(), ctx.TimeScaleDB, request)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, response)
}

