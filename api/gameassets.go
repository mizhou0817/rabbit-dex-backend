package api

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/strips-finance/rabbit-dex-backend/gameassets"
)

func HandleGameAssetsBlastPost(c *gin.Context) {
	var request struct {
		Data []gameassets.BlastAssetsLoaded `json:"data"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("failed to decode request: %w", err))
		return
	}

	ctx := GetRabbitContext(c)

	result, err := gameassets.BlastLoadAssetsBatch(c, ctx.TimeScaleDB, request.Data)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("failed to load batch: %w", err))
		return
	}

	SuccessResponse(c, result)
}

func HandleGameAssetsBlastGet(c *gin.Context) {
	ctx := GetRabbitContext(c)

	assets, err := gameassets.BlastGetAssets(c, ctx.TimeScaleDB, ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("failed to get assets: %w", err))
		return
	}

	result := struct {
		gameassets.BlastAssets
		Wallet string `json:"wallet"`
	}{
		BlastAssets: assets,
		Wallet:      ctx.Profile.Wallet,
	}

	SuccessResponse(c, result)
}

func HandleBlastLeaderboard(c *gin.Context) {
	ctx := GetRabbitContext(c)
	result, err := gameassets.BlastGetLeaderboard(c, ctx.TimeScaleDB)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("failed to get leaderboard: %w", err))
		return
	}

	SuccessResponse(c, result...)
}

func HandleGameAssetsBfxPost(c *gin.Context) {
	var request struct {
		Data []gameassets.BfxAssetsLoaded `json:"data"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("failed to decode bfx post request: %w", err))
		return
	}

	ctx := GetRabbitContext(c)

	result, err := gameassets.BfxLoadAssetsBatch(c, ctx.TimeScaleDB, request.Data)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("failed to load bfx batch: %w", err))
		return
	}

	SuccessResponse(c, result)
}

func HandleBfxGetPoints(c *gin.Context) {
	var request struct {
		Data gameassets.BfxGetPointsRequest `json:"data"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("failed to decode bfx points request: %w", err))
		return
	}

	ctx := GetRabbitContext(c)

	bfxPoints, err := gameassets.BfxGetPoints(c, ctx.TimeScaleDB, request.Data)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("failed to get bfx points: %w", err))
		return
	}

	SuccessResponse(c, bfxPoints)
}

