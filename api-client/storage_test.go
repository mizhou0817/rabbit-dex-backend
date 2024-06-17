package api_client

import (
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type StorageSuite struct {
	APITestSuite
}

func generateJwt(
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

func (s *StorageSuite) TestReadWrite() {
	cfg, err := api.ReadConfig()
	require.NoError(s.T(), err)

	token, err := generateJwt(0, cfg.Service.HMACSecret, 1000)
	require.NoError(s.T(), err)

	err = s.Client().WriteStorage(token, &api.WriteStorageRequest{
		Data: []byte("SOME DATA1"),
	})
	require.Error(s.T(), err)

	token, err = generateJwt(model.MAX_PROFILE_ID, cfg.Service.HMACSecret, 1000)
	require.NoError(s.T(), err)

	err = s.Client().WriteStorage(token, &api.WriteStorageRequest{
		Data: []byte("SOME DATA2"),
	})
	require.Error(s.T(), err)

	token, err = generateJwt(10, cfg.Service.HMACSecret, 1000)
	require.NoError(s.T(), err)

	jsonStr := `"FAVORITE_MARKETS": {"PEPE1000-USD": true,"ARB-USD": true}`
	jsonBytes := []byte(jsonStr)

	err = s.Client().WriteStorage(token, &api.WriteStorageRequest{
		Data: jsonBytes,
	})
	require.NoError(s.T(), err)

	resp, err := s.Client().ReadStorage(token)
	require.NoError(s.T(), err)

	resJson1 := string(resp.Result[0])
	require.NotEmpty(s.T(), resJson1)
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, &StorageSuite{})
}
