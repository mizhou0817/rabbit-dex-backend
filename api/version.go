package api

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

type VersionRequest struct {
	Version  string `form:"version" binding:"required"`
	Platform string `form:"platform" binding:"required,oneof=android ios"`
}

type VersionResponse struct {
	ForceUpgrade bool   `json:"force_upgrade"`
	AppUrl       string `json:"app_url"`
}

var (
	ErrExternalVersionBroken = errors.New("EXTERNAL_VERSION_NUMBER_BROKEN")
	ErrInternalVersionBroken = errors.New("INTERNAL_VERSION_NUMBER_BROKEN")
)

func HandleMobileVersion(c *gin.Context) {
	var request VersionRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	defaulResponse := &VersionResponse{
		ForceUpgrade: false,
		AppUrl:       "",
	}

	ctx := GetRabbitContext(c)

	platformConfig, ok := ctx.Config.Service.Platforms[request.Platform]
	if !ok {
		SuccessResponse(c, defaulResponse)
		return
	}

	// We are using semantic notation for versions Major.Minor.Patch
	external := strings.Split(request.Version, ".")
	if len(external) != 3 {
		ErrorResponse(c, ErrExternalVersionBroken)
		return
	}

	current := strings.Split(platformConfig.CurrentVersion, ".")
	if len(current) != 3 {
		ErrorResponse(c, ErrInternalVersionBroken)
		return
	}

	//If major differ then upgrade
	if external[0] != current[0] {
		SuccessResponse(c, VersionResponse{
			ForceUpgrade: true,
			AppUrl:       platformConfig.AppUrl,
		})
		return
	}

	//If minor differ use config
	if external[1] != current[1] {
		SuccessResponse(c, VersionResponse{
			ForceUpgrade: platformConfig.ForceUpdate,
			AppUrl:       platformConfig.AppUrl,
		})
		return
	}

	// Do nothing in other way
	SuccessResponse(c, defaulResponse)
}
