package auth

import (
	"github.com/stretchr/testify/require"
	"testing"
)

/*
All testing jwt obtained via this online tool:
http://jwtbuilder.jamiekurtz.com/
*/
func TestJwtToProfileID(t *testing.T) {
	hmacSecret := "test"
	profileID := uint(123)
	testCases := []struct {
		jwt   string
		error string
	}{
		{
			jwt: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiIiLCJpYXQiOjMwLCJleHAiOjMyNTAzNjgwMDMwLCJhdWQiOiIiLCJzdWIiOiIxMjMifQ.FBEthiUJD3_fUeHwXmWdkvAbdgoh7lwr0xkRV3drAio",
		},
		{
			jwt:   "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiIiLCJpYXQiOjMwLCJleHAiOjk0NjY4NDgzMCwiYXVkIjoiIiwic3ViIjoiMTIzIn0.TON3T_m2dFP5N0dAcDRbi-whU3t6G6bh6omCue9NOEM",
			error: "Token is expired",
		},
		{
			jwt:   "some_invalid_stuff",
			error: "token contains an invalid number of segments",
		},
	}

	for _, tc := range testCases {
		extractedProfileID, err := JwtToProfileID(tc.jwt, hmacSecret)

		if tc.error != "" {
			require.ErrorContains(t, err, tc.error)
		} else {
			require.NoError(t, err)
			require.Equal(t, profileID, extractedProfileID)
		}
	}
}
