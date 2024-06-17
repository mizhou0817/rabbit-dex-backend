package auth

import (
	"encoding/json"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

/*
With this one we can extract profile id from jwt string
and ensure that it;s signature is valid and the most
important we can ensure expiration time of provided jwt.
*/
func JwtToProfileID(jwt_ string, hmacSecret string) (uint, error) {
	// Verify RefreshToken is valid and not expired by jwt package
	verifiedJwt, err := jwt.Parse(jwt_, func(_ *jwt.Token) (interface{}, error) {
		hmacSecret := []byte(hmacSecret)

		return hmacSecret, nil
	})
	if err != nil {
		return 0, err
	}

	// Next extract claims from verified jwt
	claimsMap, ok := verifiedJwt.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}

	claimsData, err := json.Marshal(claimsMap)
	if err != nil {
		return 0, err
	}

	var claims jwt.RegisteredClaims

	if err = json.Unmarshal(claimsData, &claims); err != nil {
		return 0, err
	}

	// Convert extracted from claims subject to profile id
	profileID, err := strconv.ParseUint(claims.Subject, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(profileID), nil
}
