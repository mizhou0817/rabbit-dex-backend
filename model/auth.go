package model

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang-jwt/jwt/v4"
)

const (
	remove_session_secrets                  = "remove_session_secrets"
	api_secret_create                       = "api_secret_create"
	api_secret_by_key                       = "api_secret_by_key"
	api_secret_by_profile_id                = "api_secret_by_profile_id"
	get_or_refresh_api_secret_by_profile_id = "get_or_refresh_api_secret_by_profile_id"
	api_secret_pair_create                  = "api_secret_pair_create"
	api_delete_secret_key                   = "api_delete_secret_key"
	api_secret_validate                     = "api_secret_validate"
	api_list_secrets                        = "api_list_secrets"
	api_refresh_secret                      = "api_refresh_secret"
	api_update_api_secret_expire            = "api_update_api_secret_expire"

	frontend_secret_create           = "frontend_secret_create"
	frontend_secret_by_jwt           = "frontend_secret_by_jwt"
	frontend_secret_by_refresh_token = "frontend_secret_by_refresh_token"
	frontend_secret_update           = "frontend_secret_update"
)

/*
Tarantool model to persist api keys for market maker. The one important detail
is that market maker uses Secret to sign payload he sends in POST, PUT, DELETE
request. And Key is actually public, and it can be passed through headers, it's
used just to find associated Secret key to verify signature. Api keys supposed
to be used by technically experienced users.
*/
type APISecret struct {
	Key        string `msgpack:"key" json:"Key"`
	ProfileID  uint   `msgpack:"profile_id" json:"ProfileID"`
	Secret     string `msgpack:"secret" json:"Secret"`
	Tag        string `msgpack:"tag" json:"Tag"`
	Expiration uint   `msgpack:"expiration" json:"Expiration"`
	Status     string `msgpack:"status" json:"Status"`
}

type Secret struct {
	APISecret     *APISecret `msgpack:"api_secret" json:"api_secret"`
	JwtPrivate    string     `msgpack:"jwt_private" json:"jwt_private"`
	JwtPublic     string     `msgpack:"jwt_public" json:"jwt_public"`
	RefreshToken  string     `msgpack:"refresh_token" json:"refresh_token"`
	AllowedIpList []string   `msgpack:"allowed_ip_list" json:"allowed_ip_list"`
	CreatedAt     int64      `msgpack:"created_at" json:"created_at"`
}

type (
	apiSecretRequest struct {
		APISecret *APISecret `msgpack:"api_secret"`
	}

	apiSecretResponse struct {
		APISecret *APISecret `msgpack:"api_secret"`
		Error     string     `msgpack:"error"`
	}

	apiSecretsResponse struct {
		APISecrets []*APISecret `msgpack:"api_secrets"`
		Error      string       `msgpack:"error"`
	}

	listSecretsResponse struct {
		Secrets []*Secret `msgpack:"secrets"`
		Error   string    `msgpack:"error"`
	}

	refreshSecretResponse struct {
		Secret *Secret `msgpack:"secret"`
		Error  string  `msgpack:"error"`
	}

	validateKeyResponse struct {
		IsValid bool   `msgpack:"is_valid"`
		Error   string `msgpack:"error"`
	}

	errorOnlyResponse struct {
		Error string `msgpack:"error"`
	}
)

type ApiSecretModel struct {
	broker *Broker
}

func generateSecret(profileID uint, expiredAt int64, tag string) (*APISecret, error) {
	key := make([]byte, 32)
	secret := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	_, err = rand.Read(secret)
	if err != nil {
		return nil, err
	}

	apiSecret := &APISecret{
		Key:        base64.StdEncoding.EncodeToString(key),
		ProfileID:  profileID,
		Secret:     hexutil.Encode(secret),
		Expiration: uint(expiredAt),
		Tag:        tag,
	}

	return apiSecret, nil

}

func NewApiSecretModel(broker *Broker) *ApiSecretModel {
	return &ApiSecretModel{
		broker: broker,
	}
}

func (s ApiSecretModel) UpdateApiSecretExpire(
	ctx context.Context,
	key string,
	newExpiresAt int64) error {

	var result []errorOnlyResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, api_update_api_secret_expire, []interface{}{
		key,
		newExpiresAt,
	}, &result)
	if err != nil {
		return err
	}

	if len(result) < 1 {
		return errors.New("UNKNOWN_ERROR")
	}

	if err := result[0].Error; err != "" {
		return errors.New(err)
	}

	return nil
}

func (s ApiSecretModel) RefreshSecret(
	ctx context.Context,
	profileID uint,
	refreshToken, newJwt, newRefreshToken string,
	newExpiredAt int64,
) (*Secret, error) {
	var result []refreshSecretResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, api_refresh_secret, []interface{}{
		profileID,
		refreshToken,
		newJwt,
		newRefreshToken,
		newExpiredAt,
	}, &result)
	if err != nil {
		return nil, err
	}

	if len(result) < 1 {
		return nil, errors.New("UNKNOWN_ERROR")
	}

	if err := result[0].Error; err != "" {
		return nil, errors.New(err)
	}

	return result[0].Secret, nil
}

func (s ApiSecretModel) GetSecrets(ctx context.Context, profileID uint) ([]*Secret, error) {
	var result []listSecretsResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, api_list_secrets, []interface{}{profileID}, &result)
	if err != nil {
		return nil, err
	}

	if len(result) < 1 {
		return nil, errors.New("UNKNOWN_ERROR")
	}

	if err := result[0].Error; err != "" {
		return nil, errors.New(err)
	}

	return result[0].Secrets, nil

}

func (s ApiSecretModel) CreatePair(ctx context.Context, apiSecret *APISecret, jwt, refresh_token string, ips []string) (*APISecret, error) {
	var result []apiSecretResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, api_secret_pair_create, []interface{}{apiSecret, jwt, refresh_token, ips}, &result)
	if err != nil {
		return nil, err
	}

	return s.decodeAPISecretResponse(result)
}

func (s ApiSecretModel) DeletePairForProfile(ctx context.Context, profileID uint, key string) error {
	var result []apiSecretResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, api_delete_secret_key, []interface{}{profileID, key}, &result)
	if err != nil {
		return err
	}

	if err := result[0].Error; err != "" {
		return errors.New(err)
	}

	return nil
}

func (s ApiSecretModel) CreatePairFromProfileID(ctx context.Context, tag string, profileID uint, expiration *int64, jwt, refresh_token string, ips []string) (*APISecret, error) {
	// By default market maker api key expires in 6 months
	expiredAt := time.Now().Add(time.Hour * 24 * 30 * 6).Unix()
	if expiration != nil {
		expiredAt = *expiration
	}

	apiSecret, err := generateSecret(profileID, expiredAt, tag)

	if err != nil {
		return nil, err
	}

	return s.CreatePair(ctx, apiSecret, jwt, refresh_token, ips)
}

func (s ApiSecretModel) ValidateApiKey(ctx context.Context, key string, ip string) error {
	var result []validateKeyResponse

	now := time.Now().Unix()
	err := s.broker.Execute(AUTH_INSTANCE, ctx, api_secret_validate, []interface{}{key, ip, now}, &result)
	if err != nil {
		return err
	}

	if err := result[0].Error; err != "" {
		return errors.New(err)
	}

	return nil
}

func (s ApiSecretModel) GetByKey(ctx context.Context, key string) (*APISecret, error) {
	var result []apiSecretResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, api_secret_by_key, []interface{}{key}, &result)
	if err != nil {
		return nil, err
	}

	return s.decodeAPISecretResponse(result)
}

func (s ApiSecretModel) GetOrRefreshSecretByProfileID(ctx context.Context, profileID uint) ([]*APISecret, error) {
	var result []apiSecretsResponse

	// By default market maker api key expires in 6 months
	expiredAt := time.Now().Add(time.Hour * 24 * 30 * 6).Unix()
	newApiSecret, err := generateSecret(profileID, expiredAt, "")
	if err != nil {
		return nil, err
	}

	//TODO: for backward compability, will be removed in the future
	newApiSecret.Status = API_SECRET_GEN_STATUS
	err = s.broker.Execute(AUTH_INSTANCE, ctx, get_or_refresh_api_secret_by_profile_id, []interface{}{
		profileID,
		newApiSecret}, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return make([]*APISecret, 0), nil
	}

	if result[0].Error != "" {
		return nil, fmt.Errorf(result[0].Error)
	}

	return result[0].APISecrets, nil
}

func (s ApiSecretModel) GetByProfileID(ctx context.Context, profileID uint) ([]*APISecret, error) {
	var result []apiSecretsResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, api_secret_by_profile_id, []interface{}{profileID}, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return make([]*APISecret, 0), nil
	}

	return result[0].APISecrets, nil
}

func (s ApiSecretModel) RemoveSessionSecrets(ctx context.Context, profileID uint) error {
	var result []errorOnlyResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, remove_session_secrets, []interface{}{profileID}, &result)
	if err != nil {
		return err
	}

	if len(result) < 1 {
		return errors.New("UNKNOWN_ERROR")
	}

	if err := result[0].Error; err != "" {
		return errors.New(err)
	}

	return nil
}

func (s ApiSecretModel) GenerateJwt(
	profileID uint,
	hmacSecret string,
	expiresAt int64,
) (string, error) {
	subject := strconv.Itoa(int(profileID))

	expiresAtTime := time.Unix(expiresAt, 0)
	registerClaims := jwt.RegisteredClaims{Subject: subject, ExpiresAt: jwt.NewNumericDate(expiresAtTime)}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, registerClaims)

	return jwtToken.SignedString([]byte(hmacSecret))
}

func (s ApiSecretModel) decodeAPISecretResponse(res []apiSecretResponse) (*APISecret, error) {
	if len(res) < 1 {
		return nil, errors.New("UNKNOWN_ERROR")
	}

	if err := res[0].Error; err != "" {
		return nil, errors.New(err)
	}

	return res[0].APISecret, nil
}

/*
This model is intended to store rotating secrets to perform payload signing for
multiple browser session. LONG story short it holds secrets for frontend users
which interacting with Rabbit DEX via browser interface.
*/
type FrontendSecret struct {
	Jwt          string `msgpack:"jwt"`
	ProfileID    uint   `msgpack:"profile_id"`
	RandomSecret string `msgpack:"random_secret"`
	RefreshToken string `msgpack:"refresh_token"`
	Status       string `msgpack:"status"`
}

type (
	frontendSecretRequest struct {
		FrontendSecret *FrontendSecret `msgpack:"frontend_secret"`
		OldJwt         string          `msgpack:"old_jwt"`
	}

	frontendSecretResponse struct {
		FrontendSecret *FrontendSecret `msgpack:"frontend_secret"`
		Error          string          `msgpack:"error"`
	}

	frontendSecretsResponse struct {
		FrontendSecrets []*FrontendSecret `msgpack:"frontend_secrets"`
		Error           string            `msgpack:"error"`
	}
)

type FrontendSecretModel struct {
	broker *Broker
}

func NewFrontendSecretModel(broker *Broker) *FrontendSecretModel {
	return &FrontendSecretModel{
		broker: broker,
	}
}

func (s FrontendSecretModel) Create(
	ctx context.Context,
	frontendSecret *FrontendSecret,
	oldJwt string,
) (*FrontendSecret, error) {
	var result []frontendSecretResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, frontend_secret_create, []interface{}{frontendSecret, oldJwt}, &result)
	if err != nil {
		return nil, err
	}

	return s.decodeFrontendSecretResponse(result)
}

func (s FrontendSecretModel) CreateFromProfileID(
	ctx context.Context,
	profileID uint,
	oldJwt string,
	hmacSecret string,
	jwtLifetime uint64,
	refreshTokenLifetime uint64,
) (*FrontendSecret, error) {
	randomSecret := make([]byte, 32)

	_, err := rand.Read(randomSecret)
	if err != nil {
		return nil, err
	}

	jwt_, err := s.generateJwt(profileID, hmacSecret, jwtLifetime)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateJwt(profileID, hmacSecret, refreshTokenLifetime)
	if err != nil {
		return nil, err
	}

	apiSecret := &FrontendSecret{
		Jwt:          jwt_,
		ProfileID:    profileID,
		RandomSecret: hexutil.Encode(randomSecret),
		RefreshToken: refreshToken,
	}

	return s.Create(ctx, apiSecret, oldJwt)
}

func (s FrontendSecretModel) GetByJwt(ctx context.Context, jwt string) (*FrontendSecret, error) {
	var result []frontendSecretResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, frontend_secret_by_jwt, []interface{}{jwt}, &result)
	if err != nil {
		return nil, err
	}

	return s.decodeFrontendSecretResponse(result)
}

func (s FrontendSecretModel) GetByRefreshToken(ctx context.Context, refreshToken string) (*FrontendSecret, error) {
	var result []frontendSecretResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, frontend_secret_by_refresh_token, []interface{}{refreshToken}, &result)
	if err != nil {
		return nil, err
	}

	return s.decodeFrontendSecretResponse(result)
}

func (s FrontendSecretModel) Update(ctx context.Context, frontendSecret *FrontendSecret) (*FrontendSecret, error) {
	var result []frontendSecretResponse

	err := s.broker.Execute(AUTH_INSTANCE, ctx, frontend_secret_update, []interface{}{frontendSecret}, &result)
	if err != nil {
		return nil, err
	}

	return s.decodeFrontendSecretResponse(result)
}

func (s FrontendSecretModel) generateJwt(
	profileID uint,
	hmacSecret string,
	lifetime uint64,
) (string, error) {
	subject := strconv.Itoa(int(profileID))
	expiresAt := time.Now().Add(time.Second * time.Duration(lifetime))
	registerClaims := jwt.RegisteredClaims{Subject: subject, ExpiresAt: jwt.NewNumericDate(expiresAt)}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, registerClaims)

	return jwtToken.SignedString([]byte(hmacSecret))
}

func (s FrontendSecretModel) decodeFrontendSecretResponse(res []frontendSecretResponse) (*FrontendSecret, error) {
	if len(res) < 1 {
		return nil, errors.New("UNKNOWN_ERROR")
	}

	if err := res[0].Error; err != "" {
		return nil, errors.New(err)
	}

	return res[0].FrontendSecret, nil
}
