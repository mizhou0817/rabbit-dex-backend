package api

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/strips-finance/rabbit-dex-backend/auth"
)

func Router() *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(RabbitMiddleware())
	router.Use(PaginationMiddleware())
	router.Use(AnalyticsMiddleware())

	setupCORS(router)

	// for syncing timestamp with client
	router.GET("/srvtimestamp", HandleSrvTimestamp)

	// for validating pm
	router.GET("/showmeta", HandleShowMeta)

	// mobile version
	router.GET("/version/mobile", HandleMobileVersion)

	// onboarding
	router.POST("/onboarding", HandleOnboarding)
	router.POST("/jwt", HandleJwt)
	router.GET("/account/validate", HandleAccountValidate)

	// frontend storage - lightweight handler
	// check JWT inside handler without request to tarantool
	jwtOnlyAuth := router.Group("/")
	jwtOnlyAuth.Use(AuthJwtMiddleware)
	jwtOnlyAuth.GET("/storage", HandleReadStorage)
	jwtOnlyAuth.POST("/storage", HandleWriteStorage)

	// markets
	router.Use(CompressionMiddleware())
	router.GET("/markets", HandleMarket)
	router.GET("/markets/coins", HandleMarketCoins)
	router.GET("/markets/trades", HandleMarketTrades)
	router.GET("/markets/orderbook", HandleMarketOrderBook)
	router.GET("/markets/fundingrate", HandleFundingRateList)
	router.GET("/candles", HandleCandleList)
	router.GET("/bfx/volume", HandleBfxVolume)
	router.GET("/rbx/volume", HandleRbxVolume)

	router.GET("/vaults", HandleVault)
	router.GET("/vaults/navhistory", HandleNavHistory)

	router.GET("/blast/points", HandleBlastPoints)

	authRequired := router.Group("/")
	authRequired.Use(AuthMiddleware)
	authRequired.POST("/orders", HandleOrderCreate)
	authRequired.GET("/orders", HandleOrdersList)
	authRequired.PUT("/orders", HandleOrderAmend)
	authRequired.DELETE("/orders", HandleOrderCancel)
	authRequired.DELETE("/orders/cancel_all", HandleOrderCancelAll)

	// vault info
	authRequired.GET("/vaults/holdings", HandleVaultHoldings)
	authRequired.GET("/vaults/balanceops", HandleVaultBalanceOperations)

	// dead man's switch
	authRequired.GET("/cancel_all_after", HandleDeadmanList)
	authRequired.POST("/cancel_all_after", HandleDeadmanCreate)
	authRequired.DELETE("/cancel_all_after", HandleDeadmanDelete)

	// portfolio
	authRequired.GET("/portfolio", HandlePortfolioList)

	// IMPORTANT: urls path has changed
	authRequired.GET("/balanceops", HandleBalanceOpsList)

	authRequired.GET("/account", HandleAccount)
	authRequired.PUT("/account/leverage", HandleAccountSetLeverage)

	authRequired.GET("/fills", HandleFillsList)
	authRequired.GET("/fills/order", HandleFillsForOrder)
	authRequired.GET("/positions", HandlePositionsList)

	authRequired.GET("/profile", HandleProfileCacheRequest)

	authRequired.GET("/airdrops", HandleAirdropList)
	authRequired.POST("/airdrops/claim", HandleClaimAll)

	authRequired.POST("/secrets/refresh", HandleSecretRefresh)
	authRequired.POST("/secrets/session/remove", HandleSecretsSessionRemove)

	authRequired.POST("/balanceops/processing", ProcessingWithdrawal)
	authRequired.POST("/balanceops/deposit", HandleDeposit)

	authRequired.POST("/balanceops/stake", HandleStake)
	authRequired.POST("/balanceops/stake_from_balance", HandleStakeFromBalance)
	authRequired.POST("/balanceops/unstake/begin", BeginUnstake)
	authRequired.POST("/balanceops/unstake/cancel", CancelUnstake)
	authRequired.GET("/balanceops/unstake/requested", HandleRequestedUnstakesList)

	authRequired.POST("/balanceops/init_vault", InitVault)
	authRequired.POST("/balanceops/reactivate_vault", ReactivateVault)
	authRequired.POST("/balanceops/unstake/process", ProcessUnstakes)

	authRequired.GET("/game_assets/blast", HandleGameAssetsBlastGet)
	authRequired.GET("/game_assets/blast_leaderboard", HandleBlastLeaderboard)
	authRequired.GET("/game_assets/bfx_points", HandleBfxGetPoints)

	authRequired.GET("/storage/profile_data", HandleProfileDataRead)
	authRequired.POST("/storage/profile_data", HandleProfileDataReplace)

	// referral
	authRequired.POST("/referral", HandleReferralCreate)
	authRequired.GET("/referral", HandleReferralGet)
	authRequired.PATCH("/referral", HandleReferralEdit)
	router.GET("/referral/leaderboard", HandleGetLeaderBoard)

	signatureRequired := router.Group("/")
	signatureRequired.Use(AuthMiddleware)
	signatureRequired.Use(MetamaskSignatureMiddleware(auth.TREASURER_ROLE))
	signatureRequired.POST("/balanceops/withdraw", HandleWithdrawal)
	signatureRequired.POST("/balanceops/claim", ClaimWithdrawal)
	signatureRequired.DELETE("/balanceops/cancel", CancelWithdrawal)

	secretsRoleRequired := router.Group("/")
	secretsRoleRequired.Use(AuthMiddleware)
	secretsRoleRequired.Use(MetamaskSignatureMiddleware(auth.SECRETS_ROLE))
	secretsRoleRequired.GET("/secrets", HandleListSecrets)
	secretsRoleRequired.POST("/secrets", HandleSecretCreate)
	secretsRoleRequired.DELETE("/secrets", HandleSecretDelete)

	adminAuthRequired := router.Group("/admin")
	adminAuthRequired.Use(AuthMiddleware)
	adminAuthRequired.Use(AdminAuthMiddleware)
	adminAuthRequired.POST("/markets/url", HandleChangeIconUrl)
	adminAuthRequired.POST("/markets/title", HandleChangeMarketTitle)

	superAdminAuthRequired := router.Group("/super/admin")
	superAdminAuthRequired.Use(AuthMiddleware)
	superAdminAuthRequired.Use(SuperAdminAuthMiddleware)
	superAdminAuthRequired.GET("/tiers", HandleGetTiers)
	superAdminAuthRequired.GET("/tiers/special", HandleGetSpecialTiers)
	superAdminAuthRequired.GET("/tiers/profile", HandleGetProfileTiers)
	superAdminAuthRequired.GET("/tiers/which", HandleWhichTier)

	superAdminAuthRequired.POST("/tiers", HandleAddTier)
	superAdminAuthRequired.POST("/tiers/special", HandleAddSpecialTier)
	superAdminAuthRequired.POST("/tiers/profile", HandleAddProfileTier)

	superAdminAuthRequired.POST("/tiers/edit", HandleEditTier)
	superAdminAuthRequired.POST("/tiers/special/edit", HandleEditSpecialTier)

	superAdminAuthRequired.DELETE("/tiers", HandleRemoveTier)
	superAdminAuthRequired.DELETE("/tiers/special", HandleRemoveSpecialTier)
	superAdminAuthRequired.DELETE("/tiers/profile", HandleRemoveProfileTier)

	jwtSuperAdminRequired := router.Group("/")
	jwtSuperAdminRequired.Use(SuperAdminJWTMiddleware)
	jwtSuperAdminRequired.POST("/game_assets/blast", HandleGameAssetsBlastPost)
	jwtSuperAdminRequired.POST("/game_assets/bfx", HandleGameAssetsBfxPost)

	return router
}

func setupCORS(router *gin.Engine) {
	allowMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodHead,
	}
	// TODO: move this to config
	allowOrigins := []string{
		"http://localhost:3000",
		"https://localhost:3000",
		"https://general-changes-sep-13.du059w5bjnwva.amplifyapp.com",
		"https://saveload.tradingview.com",
		"https://main.du059w5bjnwva.amplifyapp.com",
		"https://performance-optimizations.du059w5bjnwva.amplifyapp.com",
		"https://use-deserialization-middleware-and-naming-changes.du059w5bjnwva.amplifyapp.com",
		"https://testnet.rabbitx.io",
		"https://develop.du059w5bjnwva.amplifyapp.com",
		"https://dev.rabbitx.io",
		"https://app.rabbitx.io",
		"https://testnet.rabbitx.io/",
		"https://app.blastfutures.com",
		"https://testnet.blastfutures.com",
		"https://preview.du059w5bjnwva.amplifyapp.com/",
		"https://preview.du059w5bjnwva.amplifyapp.com",
		"https://preview-app.du059w5bjnwva.amplifyapp.com",
		"https://preview2.du059w5bjnwva.amplifyapp.com",
		"https://bfx.du059w5bjnwva.amplifyapp.com/",
		"https://bfx.du059w5bjnwva.amplifyapp.com",
		"https://bfx.trade",
		"https://bfx-app-preview.du059w5bjnwva.amplifyapp.com",
	}
	allowHeaders := []string{
		"Access-Control-Allow-Credentials",
		"Access-Control-Allow-Origin",
		"Max-Content-Length",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"Cookie",
		"RBT-SIGNATURE",
		"RBT-API-KEY",
		"RBT-TS",
		"RBT-PK-SIGNATURE",
		"RBT-PK-TS",
		"EID",
	}
	exposeHeaders := []string{
		"Access-Control-Allow-Credentials",
		"Access-Control-Allow-Origin",
		"Set-Cookie",
		"Max-Content-Length",
		"Content-Type",
		"Srv-Timestamp",
	}
	maxAge := 12 * time.Hour
	router.Use(cors.New(cors.Config{
		AllowMethods:           allowMethods,
		AllowAllOrigins:        false,
		AllowOrigins:           allowOrigins,
		AllowHeaders:           allowHeaders,
		ExposeHeaders:          exposeHeaders,
		AllowCredentials:       true,
		MaxAge:                 maxAge,
		AllowWildcard:          true,
		AllowBrowserExtensions: true,
		AllowWebSockets:        true,
		AllowFiles:             true,
	}))
}
