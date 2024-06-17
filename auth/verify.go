package auth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/signer"
)

type SignatureType string

const (
	TRADER_ROLE                  = 1
	TREASURER_ROLE               = 2
	SECRETS_ROLE                 = 3
	EIP_712        SignatureType = "EIP_712"
	EIP_191        SignatureType = "EIP_191"
)

/*
This should be used mainly in tests to check signature
verification mechanism.
*/
type MetamaskSignRequest struct {
	Message   string
	Timestamp int64
}

/*
This struct used only to collect all required argument to
call VerifyECDSA function to ensure signature is valid or not.
*/
type MetamaskVerifyRequest struct {
	Wallet        string
	Timestamp     int64
	Signature     string
	ProfileType   string
	EIP712Encoder *signer.EIP712Encoder
}

/*
Converts message and timestamp to a hash following EIP-191 format.
*/
func eip191Message(message string, timestamp int64) []byte {
	metamaskMessage := fmt.Sprintf("%s\n%d", message, timestamp)
	eip191Message := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(metamaskMessage), metamaskMessage)

	return crypto.Keccak256Hash([]byte(eip191Message)).Bytes()
}

/*
Converts message and timestamp to a hash following EIP-712 format.
*/
func eip712Message(message string, timestamp *big.Int, encoder *signer.EIP712Encoder) []byte {
	eip712Msg, err := encoder.EncodeData(
		"signin",
		apitypes.Types{
			"signin": []apitypes.Type{
				{Name: "message", Type: "string"},
				{Name: "timestamp", Type: "uint256"},
			},
		},
		apitypes.TypedDataMessage{
			"message":   message,
			"timestamp": timestamp,
		},
	)
	if err != nil {
		logrus.Errorf("Failed to encode EIP712 message: %s", err.Error())
		return nil
	}
	return eip712Msg
}

/*
EthSign signs a message with timestamp by encoding them in accordance with
either EIP 712 or EIP 191, as requested, hashing the result and then signing
the hash using the private key provided.

If EIP 712 encoding is requeted then an EIP712Encoder must be provided. If
EIP 191 encoding is requested then the EIP712Encoder ref can be nil.

The signature is returned as a hex encoded string ("0x3623...").

This function should be generally used to test obtained signature is valid
and can be successfully verified by MetamaskVerify function. It's not supposed
to be used in production, it does what metamask should do on frontend side
to generate signature from private key it stores.
*/
func EthSign(signatureType SignatureType, request *MetamaskSignRequest, privateKey *ecdsa.PrivateKey, encoder *signer.EIP712Encoder) (string, error) {
	var message []byte
	if signatureType == EIP_712 {
		message = eip712Message(request.Message, big.NewInt(request.Timestamp), encoder)
	} else {
		message = eip191Message(request.Message, request.Timestamp)
	}
	signature, err := crypto.Sign(message, privateKey)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(signature), nil
}

/*
RecoverSigner recovers the address whose private key was used to sign a
message with timestamp. It expects that the signature was created by encoding
the message and timestamp in accordance with either EIP 712 or EIP 191, then
hashing the encoded data and signing that. Which of EIP 712 or EIP 191 was
used is indicated by the signatureType argument.

The signature should be a hex encoded string ("0x3623...").

If EIP 712 was used during the creation of the signature then an EIP712Encoder
must be provided abd its Domain must match the one used to create the signature.
If EIP 191 was used then the EIP712Encoder ref can be nil.
*/
func RecoverSigner(signature string, message string, timestamp int64, signatureType SignatureType, encoder *signer.EIP712Encoder) (common.Address, error) {
	var messageHash []byte
	if signatureType == EIP_712 {
		messageHash = eip712Message(message, big.NewInt(timestamp), encoder)
	} else {
		messageHash = eip191Message(message, timestamp)
	}
	signatureBytes, err := hexutil.Decode(signature)
	if err != nil {
		return common.Address{}, err
	}
	recoveredPublicKey, err := crypto.SigToPub(messageHash, signatureBytes)
	if err != nil {
		return common.Address{}, err
	}
	recoveredWallet := crypto.PubkeyToAddress(*recoveredPublicKey)

	return recoveredWallet, nil
}

func VerifyProfile(signatureType SignatureType, request *MetamaskVerifyRequest, requiredRole uint, messages []string) error {
	wallet := common.HexToAddress(request.Wallet)

	// Now ensure signature is not expired.
	currentTimestamp := time.Now().Unix()

	if currentTimestamp >= request.Timestamp {
		err := ExpiredSignatureError{
			SignatureTimestamp: request.Timestamp,
			CurrentTimestamp:   currentTimestamp,
		}

		return err
	}

	// Supporting both EIP 191 and 712 signatures and old and new sign in
	// messages during the transition period.
	// Try each combination and return nil, indicating success, if one works.
	var err, firstErr error
	for _, message := range messages {
		for _, signatureType := range []SignatureType{EIP_712, EIP_191} {
			err = checkSigner(message, request, wallet, signatureType, requiredRole)
			if err == nil {
				return nil
			} else if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func checkSigner(message string, request *MetamaskVerifyRequest,
	wallet common.Address, signatureType SignatureType,
	requiredRole uint) error {

	recoveredWallet, err := RecoverSigner(request.Signature, message, request.Timestamp, signatureType, request.EIP712Encoder)
	if err != nil {
		return err
	}

	logrus.
		WithField("recoveredWallet", recoveredWallet.String()).
		Info("Deconding step3")

	if request.ProfileType == model.PROFILE_TYPE_VAULT {

		broker, err := model.GetBroker()
		if err != nil {
			logrus.Error(err)
			return err
		}

		apiModel := model.NewApiModel(broker)
		err = apiModel.IsValidSigner(
			context.Background(),
			wallet.String(),
			recoveredWallet.String(),
			requiredRole,
		)

		//TODO: deal with it about why it was an error, and move back
		//err = vault.IsValidSigner(wallet, recoveredWallet, requiredRole)

		if err != nil {
			logrus.
				WithField("traderWallet", recoveredWallet.String()).
				WithField("vaultContract", wallet.String()).
				WithField("requireRole", requiredRole).
				Error(err)

			return NotValidSignerError{
				Msg:          err.Error(),
				Wallet:       recoveredWallet.String(),
				Vault:        wallet.String(),
				RequiredRole: requiredRole,
			}
		}

	} else {
		// Compare recovered wallet equals the user provided wallet.
		if wallet != recoveredWallet {
			err := InvalidMetamaskSignatureError{
				ExpectedWallet:  wallet,
				RecoveredWallet: recoveredWallet,
			}

			return err
		}

	}
	return nil
}
