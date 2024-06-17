package tests

import (
	"context"
	"testing"

	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/strips-finance/rabbit-dex-backend/auth"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type TestAuthSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc

	api *model.ApiSecretModel
}

func (s *TestAuthSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), time.Minute)

	broker := ClearAll(s.T(), SkipInstances("api-gateway"))
	require.NotNil(s.T(), broker)

	s.api = model.NewApiSecretModel(broker)
	require.NotNil(s.T(), s.api)
}

func (s *TestAuthSuite) TearDownTest() {
	s.cancel()
}

func TestCexAuth(t *testing.T) {
	suite.Run(t, new(TestCexAuthSuite))
}

type TestCexAuthSuite struct {
	TestAuthSuite
}

func (s *TestCexAuthSuite) TestCexFlow() {
	var (
		profileId  uint   = 111
		hmacSecret string = "hmacsecret"
	)

	expiresAt := time.Now().Add(time.Second * time.Duration(24*60*60)).Unix()

	jwt, err := s.api.GenerateJwt(profileId, hmacSecret, expiresAt)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), jwt)

	for i := 0; i < 3; i++ {
		res, err := s.api.CreatePairFromProfileID(
			context.Background(),
			"tag",
			profileId,
			&expiresAt,
			jwt,
			jwt,
			[]string{"1", "2", "3"})
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), res)
		require.NotEmpty(s.T(), res.Key)
		require.NotEmpty(s.T(), res.Secret)
		require.Equal(s.T(), "tag", res.Tag)
		require.Equal(s.T(), profileId, res.ProfileID)
		require.Equal(s.T(), expiresAt, int64(res.Expiration))
	}

	secrets, err := s.api.GetSecrets(context.Background(), profileId)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 3, len(secrets))

	var someSecret *model.Secret = secrets[0]

	for _, sec := range secrets {
		require.Equal(s.T(), "tag", sec.APISecret.Tag)
		require.Equal(s.T(), expiresAt, int64(sec.APISecret.Expiration))
		require.Equal(s.T(), jwt, sec.JwtPrivate)
		require.Equal(s.T(), jwt, sec.RefreshToken)

		apiSecret, err := s.api.GetByKey(context.Background(), sec.APISecret.Key)
		require.NoError(s.T(), err)
		require.Equal(s.T(), sec.APISecret.Key, apiSecret.Key)

		err = s.api.ValidateApiKey(
			context.Background(),
			sec.APISecret.Key,
			"1",
		)
		require.NoError(s.T(), err)

		err = s.api.ValidateApiKey(
			context.Background(),
			sec.APISecret.Key,
			"111",
		)
		require.Equal(s.T(), "NOT_ALLOWED_FOR_IP", err.Error())
	}

	// Create with empty list and check that any IP allowed
	res, err := s.api.CreatePairFromProfileID(
		context.Background(),
		"tag",
		profileId,
		&expiresAt,
		jwt,
		jwt,
		nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), res)

	err = s.api.ValidateApiKey(
		context.Background(),
		res.Key,
		"1qqq",
	)
	require.NoError(s.T(), err)

	err = s.api.DeletePairForProfile(context.Background(), profileId, res.Key)
	require.NoError(s.T(), err)

	err = s.api.ValidateApiKey(
		context.Background(),
		res.Key,
		"1qqq",
	)
	require.Equal(s.T(), "NO_SUCH_KEY", err.Error())

	//Test refresh
	restoredProfileId, err := auth.JwtToProfileID(someSecret.JwtPrivate, hmacSecret)
	require.NoError(s.T(), err)
	require.Equal(s.T(), profileId, restoredProfileId)

	newExpiresAt := time.Now().Add(time.Second * time.Duration(30*24*60*60)).Unix()

	newJwt, err := s.api.GenerateJwt(profileId, hmacSecret, newExpiresAt)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), newJwt)

	newRefreshToken, err := s.api.GenerateJwt(profileId, hmacSecret, newExpiresAt*2)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), newRefreshToken)

	new_secret, err := s.api.RefreshSecret(context.Background(),
		profileId,
		someSecret.RefreshToken,
		newJwt,
		newRefreshToken,
		newExpiresAt,
	)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), new_secret)
	require.Equal(s.T(), someSecret.APISecret.Key, new_secret.APISecret.Key)

	require.Equal(s.T(), newExpiresAt, int64(new_secret.APISecret.Expiration))
	require.Equal(s.T(), newJwt, new_secret.JwtPrivate)
	require.Equal(s.T(), newRefreshToken, new_secret.RefreshToken)
	require.Equal(s.T(), someSecret.JwtPublic, new_secret.JwtPublic)
}

func (s *TestCexAuthSuite) TestMaxPerProfile() {
	var (
		profileId1 uint   = 211
		profileId2 uint   = 311
		hmacSecret string = "hmacsecret"
	)

	expiresAt := time.Now().Add(time.Second * time.Duration(24*60*60)).Unix()

	jwt, err := s.api.GenerateJwt(profileId1, hmacSecret, expiresAt)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), jwt)

	for i := 0; i < 100; i++ {
		res, err := s.api.CreatePairFromProfileID(
			context.Background(),
			"tag",
			profileId1,
			&expiresAt,
			jwt,
			jwt,
			[]string{"1", "2", "3"})
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), res)
		require.NotEmpty(s.T(), res.Key)
		require.NotEmpty(s.T(), res.Secret)
	}

	jwt, err = s.api.GenerateJwt(profileId2, hmacSecret, expiresAt)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), jwt)

	for i := 0; i < 100; i++ {
		res, err := s.api.CreatePairFromProfileID(
			context.Background(),
			"tag",
			profileId2,
			&expiresAt,
			jwt,
			jwt,
			[]string{"1", "2", "3"})
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), res)
		require.NotEmpty(s.T(), res.Key)
		require.NotEmpty(s.T(), res.Secret)
	}

}
