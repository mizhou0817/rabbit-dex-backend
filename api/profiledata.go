package api

import (
	"github.com/gin-gonic/gin"

	"github.com/strips-finance/rabbit-dex-backend/profiledata"
)

func HandleProfileDataRead(c *gin.Context) {
	rabbitCtx := GetRabbitContext(c)
	storage := profiledata.NewStorage(rabbitCtx.Profile.ProfileId, rabbitCtx.TimeScaleDB)
	data, err := storage.Get(c)
	if err != nil {
		ErrorResponse(c, err)
	}
	SuccessResponse(c, data)
}

func HandleProfileDataReplace(c *gin.Context) {
	var newData profiledata.ProfileData
	if err := c.ShouldBindJSON(&newData); err != nil {
		ErrorResponse(c, err)
		return
	}
	rabbitCtx := GetRabbitContext(c)

	storage := profiledata.NewStorage(rabbitCtx.Profile.ProfileId, rabbitCtx.TimeScaleDB)
	dataInStorage, err := storage.Replace(c, newData)
	if err != nil {
		ErrorResponse(c, err, dataInStorage)
		return
	}
	SuccessResponse(c, dataInStorage)
}
