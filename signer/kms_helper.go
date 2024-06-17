package signer

import (
	"crypto/ecdsa"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	REGION = "ap-northeast-1"
)

func getKmsClient() (client *kms.KMS, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(REGION),
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating AWS session: %v", err)
	}
	kmsClient := kms.New(sess)
	return kmsClient, nil
}

func Encrypt(plaintext string, key string) (ciphertext string, err error) {
	kmsClient, err := getKmsClient()
	if err != nil {
		return "", err
	}
	encryptInput := &kms.EncryptInput{
		KeyId:     aws.String(key),
		Plaintext: []byte(plaintext),
	}
	encryptOutput, err := kmsClient.Encrypt(encryptInput)
	if err != nil {
		return "", fmt.Errorf("Error encrypting string: %v", err)
	}

	return string(encryptOutput.CiphertextBlob), nil
}

func Decrypt(ciphertext string) (plaintext string, err error) {
	kmsClient, err := getKmsClient()
	if err != nil {
		return "", err
	}
	decryptInput := &kms.DecryptInput{
		CiphertextBlob: []byte(ciphertext),
	}
	decryptOutput, err := kmsClient.Decrypt(decryptInput)
	if err != nil {
		return "", fmt.Errorf("Error decrypting: %v", err)
	}
	plaintext = string(decryptOutput.Plaintext)
	return plaintext, nil
}

func CreateEncryptDecryptKey(keyDescription string) (*kms.CreateKeyOutput, error) {
	return _createKey(keyDescription, "ENCRYPT_DECRYPT")
}

func CreateSignVerifyKey(keyDescription string) (*kms.CreateKeyOutput, error) {
	return _createKey(keyDescription, "SIGN_VERIFY")
}

func _createKey(keyDescription string, keyUsage string) (*kms.CreateKeyOutput, error) {
	kmsClient, err := getKmsClient()
	if err != nil {
		return nil, err
	}
	createKeyInput := &kms.CreateKeyInput{
		Description: aws.String(keyDescription),
		KeyUsage:    aws.String(keyUsage),
		Origin:      aws.String("AWS_KMS"),
	}
	return kmsClient.CreateKey(createKeyInput)
}

func SignData(message []byte, keyID string, messageType, signingAlgorithm string) ([]byte, error) {
	kmsClient, err := getKmsClient()
	if err != nil {
		return nil, err
	}

	//messageType := kms.MessageTypeDigest
	//signingAlgorithm := kms.SigningAlgorithmSpecEcdsaSha256
	input := &kms.SignInput{
		KeyId:            aws.String(keyID),
		Message:          message,
		MessageType:      &messageType,
		SigningAlgorithm: &signingAlgorithm,
	}

	output, err := kmsClient.Sign(input)
	if err != nil {
		return nil, err
	}

	return output.Signature, nil
}

func GetPublicKey(keyID string) (*kms.GetPublicKeyOutput, error) {
	kmsClient, err := getKmsClient()
	if err != nil {
		return nil, err
	}

	input := &kms.GetPublicKeyInput{
		KeyId: aws.String(keyID),
	}
	output, err := kmsClient.GetPublicKey(input)
	fmt.Printf("kmsClient.GetPublicKey, output: %v, err: %v\n", output, err)
	if err != nil {
		return nil, err
	}
	return output, nil

}

func GetEthAddress(keyID string) (*common.Address, error) {
	getPubKeyOutput, err := GetPublicKey(keyID)
	if err != nil {
		return nil, err
	}
	type publicKeyInfo struct {
		Raw       asn1.RawContent
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}

	var pki publicKeyInfo
	asn1.Unmarshal(getPubKeyOutput.PublicKey, &pki)
	asn1Data := pki.PublicKey.RightAlign()
	_, x, y := asn1Data[0], asn1Data[1:33], asn1Data[33:]

	// fmt.Println("x and y : ", hex.EncodeToString(x), hex.EncodeToString(y))

	x_big := new(big.Int)
	x_big.SetBytes(x)
	y_big := new(big.Int)
	y_big.SetBytes(y)
	pubkey := ecdsa.PublicKey{Curve: crypto.S256(), X: x_big, Y: y_big}
	address := crypto.PubkeyToAddress(pubkey)

	return &address, nil
}
