package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/strips-finance/rabbit-dex-backend/api/types"

	"golang.org/x/exp/slices"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/auth"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/signer"
)

const (
	APIKeyHeader          = "RBT-API-KEY"
	TimestampHeader       = "RBT-TS"
	SignatureHeader       = "RBT-SIGNATURE"
	PKSignatureHeader     = "RBT-PK-SIGNATURE"
	PKTimestampHeader     = "RBT-PK-TS"
	IPHeader              = "X-Forwarded-For"
	ExchangeIdHeader      = "EID"
	SignatureLifetime     = 600
	ClientHeader          = "X-RBT-Client"
	SrvTimestampHeader    = "SRV-TIMESTAMP"
	ContentEncodingHeader = "Content-Encoding"

	ADHOC_TESTNET_BFX_MM_PROFILE_ID = 22823
	ADHOC_MAINNET_BFX_MM_PROFILE_ID = 23028
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

type RabbitContext struct {
	Config                *Config
	Timestamp             int64
	PKTimestamp           int64
	MarketMakerAPIKey     string
	Signature             string
	PKSignature           string
	IPHeader              string
	Payload               *auth.Payload
	Profile               *model.Profile
	Broker                *model.Broker
	TimeScaleDB           *pgxpool.Pool
	Pagination            types.PaginationRequestParams
	BlockchainProviderUrl string
	Signer                *signer.WithdrawalSigner
	EIP712Encoder         *signer.EIP712Encoder
	AnalyticCollector     *AnalyticsCollector
	RequiredRole          uint
	ProfileIdFromJwt      uint
	ExchangeId            string
	ExchangeCfg           ExchangeConfig
	ProviderUrl           string
	UnstakeDelayBlocks    *big.Int
	EthHelper             *model.EthHelper
	Meta                  *model.MatchingMeta
}

type Domain struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	ChainId           *big.Int
	VerifyingContract string `json:"verifyingContract"`
}

func ExtractClientAndDevice(c *gin.Context) (string, string) {
	if c == nil {
		return "", ""
	}

	clientHeader := c.GetHeader(ClientHeader)
	switch {
	case strings.Contains(clientHeader, "ios"):
		return clientHeader, "ios"
	case strings.Contains(clientHeader, "android"):
		return clientHeader, "android"
	case strings.Contains(clientHeader, "desktop"):
		return clientHeader, "desktop"
	case strings.Contains(clientHeader, "other"):
		return clientHeader, "other"
	}

	return "other", "other"
}

func EnrichContextMeta(c *gin.Context, ctx *RabbitContext) {
	if ctx.Meta == nil {
		ctx.Meta = new(model.MatchingMeta)
	}

	_, device := ExtractClientAndDevice(c)
	ctx.Meta.SetDevice(device)

	if c.GetHeader(APIKeyHeader) != "" {
		ctx.Meta.SetApi(true)
	}

	ctx.Meta.SetEid(ctx.ExchangeId)
}

func GetRabbitContext(c *gin.Context) *RabbitContext {
	return c.MustGet("context").(*RabbitContext)
}

func AddHocSilentChange(c *gin.Context, cfg *Config) string {
	apiKey := c.GetHeader(APIKeyHeader)
	if apiKey == "" {
		return ""
	}

	broker, err := model.GetBroker()
	if err != nil {
		logrus.Errorf("ADDHOCK_ERROR: %s", err)
		return ""
	}

	apiSecretModel := model.NewApiSecretModel(broker)

	err = apiSecretModel.ValidateApiKey(
		c.Request.Context(),
		apiKey,
		"",
	)
	if err != nil {
		logrus.Errorf("ADDHOCK_ERROR: for apiKey=%s  error=%s", apiKey, err)
		return ""
	}

	apiSecret, err := apiSecretModel.GetByKey(c.Request.Context(), apiKey)
	if err != nil {
		logrus.Errorf("ADDHOCK_ERROR: for apiKey=%s error=%s", apiKey, err)
		return ""
	}

	modifTo := model.EXCHANGE_BFX

	if cfg.Service.EnvMode == "testnet" {
		if apiSecret.ProfileID == ADHOC_TESTNET_BFX_MM_PROFILE_ID {
			logrus.Warnf("ADDHOCK_MODIF: env=%s profileid=%d modifTo=%s", cfg.Service.EnvMode, apiSecret.ProfileID, modifTo)
			return modifTo
		}
	} else if cfg.Service.EnvMode == "prod" {
		if apiSecret.ProfileID == ADHOC_MAINNET_BFX_MM_PROFILE_ID {
			logrus.Warnf("ADDHOCK_MODIF: env=%s profileid=%d modifTo=%s", cfg.Service.EnvMode, apiSecret.ProfileID, modifTo)
			return modifTo
		}
	}

	return ""
}

func RabbitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetSameSite(http.SameSiteNoneMode)

		// Read and store config to context
		cfg, err := ReadConfig()
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		// Ad-Hoc - for bfx bot on deltix, they can't do modif fast
		// we will specially modify payload for one account
		addHocEid := AddHocSilentChange(c, cfg)

		ExchangeId := strings.ToLower(c.GetHeader(ExchangeIdHeader))
		if addHocEid != "" {
			ExchangeId = addHocEid
		} else if ExchangeId == "" {
			ExchangeId = model.EXCHANGE_DEFAULT
		}

		if !slices.Contains(model.SupportedExchangeIds, ExchangeId) {
			ErrorResponse(c, fmt.Errorf("Unsupported exchange_id=%s", ExchangeId))
			c.Abort()
			return
		}

		ExchangeCfg, ok := cfg.Service.Exchanges[ExchangeId]
		if !ok {
			ErrorResponse(c, fmt.Errorf("No config for exchange_id=%s", ExchangeId))
			c.Abort()
			return
		}

		rabbitContext := &RabbitContext{
			MarketMakerAPIKey: c.GetHeader(APIKeyHeader),
			Signature:         c.GetHeader(SignatureHeader),
			IPHeader:          c.GetHeader(IPHeader),
			PKSignature:       c.GetHeader(PKSignatureHeader),
			ExchangeId:        ExchangeId,
			ExchangeCfg:       ExchangeCfg,
		}

		rabbitContext.BlockchainProviderUrl = ExchangeCfg.ProviderUrl

		rabbitContext.Config = cfg

		chainId := big.NewInt(int64(ExchangeCfg.ChainId))

		withdrawalSigner, err := signer.NewWithdrawalSigner(
			ExchangeCfg.DomainNameWithdraw,
			ExchangeCfg.ExchangeAddress,
			chainId,
			ExchangeCfg.SignerKeyId)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}
		rabbitContext.Signer = withdrawalSigner

		logrus.Warnf("*** WITHDRAW SIGNER for exchange_id=%s domain=%s l1=%s chain_id=%s", ExchangeId, ExchangeCfg.DomainNameWithdraw, ExchangeCfg.ExchangeAddress, chainId.String())

		// RAbbitxXId by default
		encoder := signer.NewEIP712Encoder(
			ExchangeCfg.DomainNameEncoder,
			"1",
			"",
			chainId,
		)
		rabbitContext.EIP712Encoder = encoder

		broker, err := model.GetBroker()
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}
		rabbitContext.Broker = broker

		rabbitContext.ProviderUrl = ExchangeCfg.ProviderUrl
		rabbitContext.EthHelper = model.NewEthHelper(rabbitContext.ProviderUrl, 2*time.Second)

		dbpool, err := GetTimescaleDbPool(cfg.Service.TimescaledbConnectionURI)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		rabbitContext.TimeScaleDB = dbpool

		analyticsCollector, err := GetAnalyticsCollector(dbpool)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}
		rabbitContext.AnalyticCollector = analyticsCollector

		// Set timestamp if it's provided
		var timestamp int64

		if timestampStr := c.Copy().GetHeader(TimestampHeader); timestampStr != "" {
			timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				ErrorResponse(c, err)
				c.Abort()
				return
			}
		}

		rabbitContext.Timestamp = timestamp
		currentTimestamp := time.Now().Unix()

		if timestamp > 0 && timestamp < currentTimestamp {
			err = fmt.Errorf("cannot use timestamp from past, try: %d", currentTimestamp+SignatureLifetime)
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		if timestamp-currentTimestamp > SignatureLifetime {
			err = fmt.Errorf("maximum timestamp lifetime is 30 seconds: %d", currentTimestamp+SignatureLifetime)
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		// Provide converted payload via context
		payload, err := extractPayload(c, timestamp)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		rabbitContext.Payload = payload

		// Set pktimestamp if it's provided
		var pkTimestamp int64

		if timestampStr := c.Copy().GetHeader(PKTimestampHeader); timestampStr != "" {
			pkTimestamp, err = strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				ErrorResponse(c, err)
				c.Abort()
				return
			}
		}

		rabbitContext.PKTimestamp = pkTimestamp

		// Now assign collected context to Gin's context
		c.Set("context", rabbitContext)

		EnrichContextMeta(c, rabbitContext)

		c.Next()

	}
}

func SuperAdminAuthMiddleware(c *gin.Context) {
	ctx := GetRabbitContext(c)

	if ctx.Profile == nil || ctx.Profile.Wallet == "" {
		ErrorResponse(c, errors.New("NOT_AUTHENTICATED"))
		c.Abort()
		return
	}

	is_super_admin := false
	for _, admin := range ctx.Config.Service.SuperAdminWallets {
		if strings.EqualFold(admin, ctx.Profile.Wallet) {
			is_super_admin = true
			break
		}
	}

	if !is_super_admin {
		ErrorResponse(c, errors.New("NOT_SUPER_ADMIN"))
		c.Abort()
		return

	}
	c.Next()
}

func SuperAdminJWTMiddleware(c *gin.Context) {
	ctx := GetRabbitContext(c)

	jwt, err := c.Cookie("jwt")
	if err != nil {
		ErrorResponse(c, err)
		c.Abort()
		return
	}

	profileID, err := auth.JwtToProfileID(jwt, ctx.Config.Service.HMACSecret)
	if err != nil {
		ErrorResponse(c, err)
		c.Abort()
		return
	}

	if !IsAllowedProfileId(profileID) {
		ErrorResponse(c, errors.New("BROKEN_PROFILE_ID"))
		c.Abort()
		return
	}

	if ctx.Config == nil {
		ErrorResponse(c, errors.New("NO_CONFIG"))
		c.Abort()
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)

	profile, err := apiModel.GetProfileById(c.Request.Context(), profileID)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("NO_PROFILE: %w", err))
		c.Abort()
		return
	}

	isSuperAdmin := false
	for _, admin := range ctx.Config.Service.SuperAdminWallets {
		if strings.EqualFold(admin, profile.Wallet) {
			isSuperAdmin = true
			break
		}
	}
	if !isSuperAdmin {
		ErrorResponse(c, errors.New("NOT_SUPER_ADMIN"))
		c.Abort()
		return
	}

	ctx.ProfileIdFromJwt = profileID
	ctx.Profile = profile

	c.Next()
}

func AdminAuthMiddleware(c *gin.Context) {
	ctx := GetRabbitContext(c)

	if ctx.Profile == nil || ctx.Profile.Wallet == "" {
		ErrorResponse(c, errors.New("NOT_AUTHENTICATED"))
		c.Abort()
		return
	}

	is_admin := false
	for _, admin := range ctx.Config.Service.AdminWallets {
		if strings.EqualFold(admin, ctx.Profile.Wallet) {
			is_admin = true
			break
		}
	}

	if !is_admin {
		ErrorResponse(c, errors.New("NOT_ADMIN"))
		c.Abort()
		return

	}
	c.Next()
}

func AuthJwtMiddleware(c *gin.Context) {
	// light-weight auth middleware to check only JWT
	ctx := GetRabbitContext(c)

	// This is Frontend branch
	jwt, err := c.Cookie("jwt")
	if err != nil {
		ErrorResponse(c, err)
		c.Abort()
		return
	}

	profileID, err := auth.JwtToProfileID(jwt, ctx.Config.Service.HMACSecret)
	if err != nil {
		ErrorResponse(c, err)
		c.Abort()
		return
	}

	// To not allow use public token (0 or big number) - until we will fix this bug or for future
	if !IsAllowedProfileId(profileID) {
		ErrorResponse(c, errors.New("BROKEN_PROFILE_ID"))
		c.Abort()
		return
	}

	ctx.ProfileIdFromJwt = profileID

	c.Next()
}

func AuthMiddleware(c *gin.Context) {
	var secret string
	ctx := GetRabbitContext(c)
	broker, err := model.GetBroker()
	if err != nil {
		ErrorResponse(c, err)
		c.Abort()
		return
	}

	apiModel := model.NewApiModel(broker)

	if ctx.MarketMakerAPIKey != "" {
		// This is MarketMaker branch
		apiSecretModel := model.NewApiSecretModel(broker)

		err := apiSecretModel.ValidateApiKey(
			c.Request.Context(),
			ctx.MarketMakerAPIKey,
			ctx.IPHeader,
		)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		apiSecret, err := apiSecretModel.GetByKey(c.Request.Context(), ctx.MarketMakerAPIKey)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		profile, err := apiModel.GetProfileById(c.Request.Context(), apiSecret.ProfileID)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		ctx.Profile = profile
		secret = apiSecret.Secret
		c.Set("context", ctx)
	} else {
		// This is Frontend branch
		jwt, err := c.Cookie("jwt")
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		profileID, err := auth.JwtToProfileID(jwt, ctx.Config.Service.HMACSecret)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}
		profile, err := apiModel.GetProfileById(c.Request.Context(), profileID)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		frontendSecretModel := model.NewFrontendSecretModel(broker)

		frontendSecret, err := frontendSecretModel.GetByJwt(c.Request.Context(), jwt)
		if err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		ctx.Profile = profile
		secret = frontendSecret.RandomSecret
		c.Set("context", ctx)
	}

	// Next we should verify payload
	rMethod := c.Request.Method
	if !(rMethod == http.MethodPost || rMethod == http.MethodPut || rMethod == http.MethodDelete) {
		c.Next()
		return
	}

	if err = ctx.Payload.Verify(ctx.Signature, secret, ctx.Config.Service.EnvMode); err != nil {
		ErrorResponse(c, err)
		c.Abort()
		return
	}

	c.Next()
}

func extractPayload(c *gin.Context, timestamp int64) (*auth.Payload, error) {
	var data map[string]json.RawMessage

	rMethod := c.Request.Method
	if !(rMethod == http.MethodPost || rMethod == http.MethodPut || rMethod == http.MethodDelete) {
		return nil, nil
	}

	// Read raw request body to convert later
	jsonData, err := c.GetRawData()
	if err != nil {
		return nil, err
	}

	// Make it readable again for remain api handlers
	c.Request.Body = io.NopCloser(bytes.NewReader(jsonData))

	if err = json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	// Convert json.RawMessage to string
	payloadData := map[string]string{}
	for k, v := range data {
		payloadData[k] = strings.Trim(string(v), "\"")
	}

	payloadData["method"] = rMethod
	payloadData["path"] = c.FullPath()

	return auth.NewPayload(timestamp, payloadData)
}

func AnalyticsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// catch the Request Body
		var requestBodyBytes []byte
		if c.Request.Body != nil {
			// consume and restore
			requestBodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
		}

		// catch the Response Body
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		c.Writer.Header().Set(SrvTimestampHeader, strconv.Itoa(int(time.Now().Unix())))

		c.Next()

		ctx := GetRabbitContext(c)

		if strings.EqualFold(ctx.Config.Service.EnvMode, "dev") {
			logrus.Warn("Ananlytics middleware turned off")
			return
		}

		path := c.Request.URL.Path
		method := strings.ToLower(c.Request.Method)
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}

		if c.Request.Method != http.MethodGet {
			logrus.WithFields(logrus.Fields{
				"client_ip":       c.ClientIP(),
				"method":          method,
				"path":            path,
				"request_body":    string(requestBodyBytes),
				"response_status": w.Status(),
				"response_body":   w.body.String(),
			}).Info("HTTP Response")
		}

		isAllowedPath := false
		for p, methods := range ctx.Config.Service.AnalyticsConfig.AllowedURLPaths {
			if strings.HasPrefix(path, p) && slices.Contains(methods, method) {
				isAllowedPath = true
				break
			}
		}

		if !isAllowedPath {
			return
		}

		var profileId uint
		if ctx.Profile != nil {
			profileId = ctx.Profile.ProfileId
		}

		if slices.Contains(ctx.Config.Service.AnalyticsConfig.IgnoreProfileIds, profileId) {
			return
		}

		clientHeader, _ := ExtractClientAndDevice(c)

		analyticEvent := AnalyticEvent{
			ProfileId:          profileId,
			ClientHeader:       clientHeader,
			ClientIPAddress:    c.ClientIP(),
			URLPath:            path,
			HTTPMethod:         method,
			RequestBody:        string(requestBodyBytes),
			ResponseStatusCode: w.Status(),
			ResponseBody:       string(w.body.Bytes()),
		}

		go ctx.AnalyticCollector.Push(analyticEvent)
	}
}

func PaginationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := GetRabbitContext(c)

		// page (default=1)
		val, err := strconv.ParseInt(c.DefaultQuery("p_page", "0"), 10, 64)
		if err != nil || val < 0 {
			ctx.Pagination.Page = 0
		} else {
			ctx.Pagination.Page = val
		}

		// limit (default=50)
		val, err = strconv.ParseInt(c.DefaultQuery("p_limit", "50"), 10, 64)
		if err != nil || val <= 0 || val > 1000 {
			ctx.Pagination.Limit = 50
		} else {
			ctx.Pagination.Limit = val
		}

		// order (default=DESCENDING)
		order := c.DefaultQuery("p_order", "DESC")
		if order != "ASC" && order != "DESC" {
			ctx.Pagination.Order = "DESC"
		} else {
			ctx.Pagination.Order = order
		}

		c.Next()
	}
}

func MetamaskSignatureMiddleware(requireRole uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := GetRabbitContext(c)
		if ctx.PKSignature == "" {
			ErrorResponse(c, fmt.Errorf("RBT-PK-SIGNATURE header is empty: %s", ctx.PKSignature))
			c.Abort()
			return
		}

		currentTimestamp := time.Now().Unix()

		if ctx.PKTimestamp > 0 && ctx.PKTimestamp < currentTimestamp {
			err := fmt.Errorf("cannot use timestamp from past, try: %d", currentTimestamp+SignatureLifetime)
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		if ctx.PKTimestamp-currentTimestamp > SignatureLifetime {
			err := fmt.Errorf("maximum timestamp lifetime is 30 seconds: %d", currentTimestamp+SignatureLifetime)
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		ctx.RequiredRole = requireRole

		verifyRequest := &auth.MetamaskVerifyRequest{
			Wallet:        ctx.Profile.Wallet,
			Timestamp:     ctx.PKTimestamp,
			Signature:     ctx.PKSignature,
			ProfileType:   ctx.Profile.Type,
			EIP712Encoder: ctx.EIP712Encoder,
		}

		logrus.
			WithField("Wallet", ctx.Profile.Wallet).
			WithField("Timestamp", ctx.PKTimestamp).
			WithField("Signature", ctx.PKSignature).
			Info("MetamaskSignatureMiddleware")

		if err := auth.VerifyProfile(auth.EIP_712, verifyRequest, ctx.RequiredRole, ctx.ExchangeCfg.OnboardingMessages); err != nil {
			ErrorResponse(c, err)
			c.Abort()
			return
		}

		c.Next()
	}
}

func CompressionMiddleware() gin.HandlerFunc {
	gzipMiddleware := gzip.Gzip(gzip.DefaultCompression)

	return func(c *gin.Context) {
		// Read and store config to context
		cfg, err := ReadConfig()
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		// configMinResponseSize := 1200
		configMimeTypes := "application/json"

		compression := cfg.Service.Compression

		if compression.Enabled {
			// configMinResponseSize = compression.MinResponseSize
			configMimeTypes = compression.MimeTypes
		}

		mimeType := c.GetHeader("Content-Type")

		if compression.Enabled {
			if mimeType != "" && strings.Contains(configMimeTypes, mimeType) {
				gzipMiddleware(c)
			} else {
				c.Next()
			}
		} else {
			c.Next()
		}
	}
}
