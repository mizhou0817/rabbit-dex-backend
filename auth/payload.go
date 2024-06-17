package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
)

const (
	PayloadKeyMethod = "method"
	PayloadKeyPath   = "path"
)

var requiredKeys = []string{PayloadKeyMethod, PayloadKeyPath}

type Payload struct {
	Timestamp int64
	Data      map[string]string
}

func NewPayload(timestamp int64, data map[string]string) (*Payload, error) {
	// Ensure payload has all required keys
	for _, k := range requiredKeys {
		if _, ok := data[k]; !ok {
			return nil, MissingPayloadKeyError{MissedKey: k}
		}
	}

	return &Payload{
		Timestamp: timestamp,
		Data:      data,
	}, nil
}

/*
Every Hash output is collision resistant even for small changes
in payload. Also important thing is that we use timestamp during
hash calculation to ensure later during HMAC verification that
payload signature cannot be reused after expiration.
*/
func (s Payload) Hash() []byte {
	// Sort payload keys and prepare an alphabetically ordered string.
	var message string
	sortedKeys := make([]string, 0, len(s.Data))

	for k := range s.Data {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		message += fmt.Sprintf("%s=%s", k, s.Data[k])
	}

	// Convert int64 timestamp to bytes to pass to the hash input.
	//timestampBytes := make([]byte, 8)
	//binary.BigEndian.PutUint64(timestampBytes, uint64(s.Timestamp))
	timestamp := strconv.FormatInt(s.Timestamp, 10)

	// Calculate hash itself with given input.
	input := make([]byte, 0, len(message)+len(timestamp))
	input = append(input, []byte(message)...)
	input = append(input, []byte(timestamp)...)
	hash := sha256.Sum256(input)

	logrus.
		WithField("hash_input", string(input)).
		WithField("hash_output", hexutil.Encode(hash[:])).
		Info("payload hash")

	return hash[:]
}

/*
Returns HMAC-SHA256 signature after signing payload hash with
provided by user secret. Later this signature is used to ensure
signer of payload was valid. From high level overview payload
can be signed by frontend user by rotating random secret or by
market maker with constant api key secret.
*/
func (s Payload) Sign(secret string) (string, error) {
	secretBytes, err := hexutil.Decode(secret)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write(s.Hash())
	signature := mac.Sum(nil)

	return hexutil.Encode(signature), nil
}

/*
From high level architect overview by using Verify method result we can ensure
API user is eligible to perform some actions i.e. order creation. Also remember
there can be two types of users with their secret:
1) Frontend user with rotating RandomSecret
2) Market maker with "constant" api key secret
*/
func (s Payload) Verify(signature string, secret string, envmode string) error {
	expectedSignature, err := s.Sign(secret)
	if err != nil {
		return err
	}

	if expectedSignature != signature {
		if strings.EqualFold(envmode, "prod") {
			return InvalidPayloadSignatureError{
				Expected: "protected",
				Actual:   "protected",
			}
		} else {
			return InvalidPayloadSignatureError{
				Expected: expectedSignature,
				Actual:   signature,
			}
		}
	}

	// Also check signature is expired or not
	currentTimestamp := time.Now().Unix()

	if currentTimestamp >= s.Timestamp {
		return ExpiredSignatureError{
			SignatureTimestamp: s.Timestamp,
			CurrentTimestamp:   currentTimestamp,
		}
	}

	return nil
}
