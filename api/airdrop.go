package api

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/airdrop"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

type ResetAirdropRequest struct {
	Title          string  `form:"title" binding:"required"`
	StartTimestamp int64   `form:"start_timestamp" binding:"required,min=1672551995000000"`
	EndTimestamp   int64   `form:"end_timestamp" binding:"required,min=1672551995000000"`
	ProfileId      uint    `form:"profile_id" binding:"required"`
	TotalRewards   float64 `form:"total_rewards" binding:"required,min=1"`
	Claimable      float64 `form:"claimable" binding:"required,min=0"`
}

type AirdropCreateRequest struct {
	Title          string `json:"title" binding:"required"`
	StartTimestamp int64  `json:"start_timestamp" binding:"required,min=1672551995000000"`
	EndTimestamp   int64  `json:"end_timestamp" binding:"required,min=1672551995000000"`
}

type ProfileAirdropInitRequest struct {
	AirdropTitle string  `json:"title" binding:"required"`
	ProfileId    uint    `json:"profile_id" binding:"required"`
	TotalRewards float64 `json:"total_rewards" binding:"required,min=1"`
	Claimable    float64 `json:"claimable" binding:"required,min=0"`
}

type UpdateClaimableRequest struct {
	AirdropTitle string `json:"title" binding:"required"`
	ProfileId    uint   `json:"profile_id" binding:"required"`
}

type ClaimAllRequest struct {
	AirdropTitle string `json:"title" binding:"required"`
}

type ClaimAllResponse struct {
	ClaimOps *model.AirdropClaimOps `json:"claim_ops"`
	BnAmount string                 `json:"bn_amount"`
	R        string                 `json:"r"`
	S        string                 `json:"s"`
	V        uint                   `json:"v"`
}

var (
	ErrPendingClaimOpsExist = errors.New("PENDING_CLAIM_EXIST")
	ErrClaimFinishFailed    = errors.New("CLAIM_FINISH_FAILED")
	ErrInvalidWallet        = errors.New("INVALID_WALLET")
)

func _signClaimOps(claimOps *model.AirdropClaimOps, wallet string) (*ClaimAllResponse, error) {
	if !common.IsHexAddress(wallet) {
		return nil, ErrInvalidWallet
	}

	r, s, v, bigIntAmount, err := airdrop.NewAirdropSignature(claimOps.Id, wallet, claimOps.Amount)
	if err != nil {
		return nil, err
	}

	R := fmt.Sprintf("0x%x", r)
	S := fmt.Sprintf("0x%x", s)

	return &ClaimAllResponse{
		ClaimOps: claimOps,
		BnAmount: bigIntAmount.String(),
		R:        R,
		S:        S,
		V:        v,
	}, nil
}

func HandleResetAirdrop(c *gin.Context) {
	var request ResetAirdropRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	err := apiModel.TestAirdropResetAll(c.Request.Context(),
		request.Title,
		request.StartTimestamp,
		request.EndTimestamp,
		request.ProfileId,
		request.TotalRewards,
		request.Claimable)

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse[int](c)
}

func HandleAirdropList(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	//We update claimable when we want to show airdrop list
	err := apiModel.UpdateAllProfileAirdrops(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	provider, err := airdrop.NewAirdropProvider(ctx.BlockchainProviderUrl)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	claimOps, err := apiModel.PendingClaimOps(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	} else if claimOps != nil {

		/*
			There is A PENDING claimOps
			Try to finalize it
		*/
		claimed, err := provider.ProcessedClaims(claimOps.Id)
		if err != nil {
			ErrorResponse(c, err)
			return
		} else if claimed {

			/*
				It's claimed - let's update
			*/
			claimOps, err = apiModel.FinishClaim(c.Request.Context(), ctx.Profile.ProfileId)
			if err != nil {
				ErrorResponse(c, err)
				return
			} else if claimOps.Status != model.AIRDROP_CLAIMED_STATUS {
				ErrorResponse(c, ErrClaimFinishFailed)
				return
			}
		}
	}

	res, err := apiModel.GetProfileAirdrops(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	/*
		ADHOC for frontend UX - it relies on claimable. And never checks pending TXs
		So we just replace this number on the fly, from amount from pending TX
	*/
	claimOps, err = apiModel.PendingClaimOps(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	// Replace on the fly.
	if claimOps != nil {
		for i, pa := range res {
			if pa.AirdropTitle == claimOps.AirdropTitle {
				res[i].Claimable = claimOps.Amount
				res[i].Claimed = *tdecimal.NewDecimal(res[i].Claimed.Sub(claimOps.Amount.Decimal))
			}
		}
	}

	SuccessResponse(c, res...)

}

func HandleCreateAirdrop(c *gin.Context) {
	var request AirdropCreateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	// profile_id uint, market_id, order_type, side string, price, size float64

	res, err := apiModel.CreateAirdrop(c.Request.Context(),
		request.Title,
		request.StartTimestamp,
		request.EndTimestamp)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleClaimAll(c *gin.Context) {
	var request ClaimAllRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	provider, err := airdrop.NewAirdropProvider(ctx.BlockchainProviderUrl)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	/*
		We allow: only 1 claim per airdrop per time

		If pending claim exist just return signature
	*/
	claimOps, err := apiModel.PendingClaimOps(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	} else if claimOps != nil { // There is A PENDING claimOps

		//Check if claim ops claimed
		claimed, err := provider.ProcessedClaims(claimOps.Id)
		if err != nil {
			ErrorResponse(c, err)
			return
		} else if !claimed { // Still pending

			// Return error if pending claimOps for another airdrop exist
			if claimOps.AirdropTitle != request.AirdropTitle {
				ErrorResponse(c, errors.New("PENDING_CLAIM_EXIST"))
				return
			}

			// Return signature
			res, err := _signClaimOps(claimOps, ctx.Profile.Wallet)
			if err != nil {
				ErrorResponse(c, err)
				return
			}

			SuccessResponse(c, res)
			return
		}

		//claimed == true Finish claim
		claimOps, err = apiModel.FinishClaim(c.Request.Context(), ctx.Profile.ProfileId)
		if err != nil {
			ErrorResponse(c, err)
			return
		} else if claimOps.Status != model.AIRDROP_CLAIMED_STATUS {
			ErrorResponse(c, ErrClaimFinishFailed)
			return
		}
	}

	//Create new claim
	claimOps, err = apiModel.ProfileClaimAll(c.Request.Context(), ctx.Profile.ProfileId, request.AirdropTitle)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	res, err := _signClaimOps(claimOps, ctx.Profile.Wallet)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleInitProfileAirdrop(c *gin.Context) {
	var request ProfileAirdropInitRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	// profile_id uint, market_id, order_type, side string, price, size float64

	res, err := apiModel.SetProfileTotal(c.Request.Context(),
		request.ProfileId,
		request.AirdropTitle,
		request.TotalRewards,
		request.Claimable)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleUpdateProfileClaimable(c *gin.Context) {
	var request UpdateClaimableRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	// profile_id uint, market_id, order_type, side string, price, size float64

	res, err := apiModel.UpdateProfileClaimable(c.Request.Context(),
		request.ProfileId,
		request.AirdropTitle)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)

}
