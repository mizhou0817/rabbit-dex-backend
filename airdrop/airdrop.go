package airdrop

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/airdrop/airdrop_l1"
	"github.com/strips-finance/rabbit-dex-backend/signer"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

type AirdropProvider struct {
	cfg         *Config
	providerUrl string
	airdropAbi  abi.ABI
	client      *ethclient.Client
	instance    *airdrop_l1.AirdropL1
}

const (
	_primaryType  = "claim"
	_strpDecimals = 18
	_domainName   = "RabbitxAirdrop"
	_version      = "1"
)

var (
	Strp_Multiplier decimal.Decimal = decimal.NewFromInt32(10).Pow(
		decimal.NewFromInt32(_strpDecimals))

	ZeroBigInt = big.NewInt(0)
)

func claimOpsAmountToStrpDecimals(amount *tdecimal.Decimal) *big.Int {
	return amount.Mul(Strp_Multiplier).Round(0).BigInt()
}

func NewAirdropSignature(id uint, trader string, amount tdecimal.Decimal) (r []byte, s []byte, v uint, bigIntAmount *big.Int, e error) {
	signer, err := airdropSigner()
	if err != nil {
		return nil, nil, 0, nil, err
	}

	bigIntAmount = claimOpsAmountToStrpDecimals(&amount)
	idBig := big.NewInt(int64(id))
	r, s, v, e = signer.KmsSignTypedData(_primaryType, apitypes.Types{
		_primaryType: []apitypes.Type{
			{Name: "id", Type: "uint"},
			{Name: "trader", Type: "address"},
			{Name: "amount", Type: "uint"},
		},
	},
		apitypes.TypedDataMessage{
			"id":     idBig,
			"trader": trader,
			"amount": bigIntAmount,
		},
	)

	return
}

func NewAirdropProvider(providerUrl string) (*AirdropProvider, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("Can't read airdrop config err=%s", err.Error())
	}

	if strings.HasPrefix(config.Service.L1AirdropAddress, "0x") {
		config.Service.L1AirdropAddress = stripPrefix(config.Service.L1AirdropAddress, 2)
	}

	airdropAbi, err := abi.JSON(strings.NewReader(
		string(airdrop_l1.AirdropL1MetaData.ABI)))
	if err != nil {
		return nil,
			fmt.Errorf("Error retrieving L1 Airdrop core contract ABI: %s", err.Error())
	}

	client, instance, err := dialAirdrop(providerUrl, config.Service.L1AirdropAddress)
	if err != nil {
		return nil,
			fmt.Errorf("dialAirdrop: %s", err.Error())
	}

	return &AirdropProvider{
		cfg:         config,
		providerUrl: providerUrl,
		airdropAbi:  airdropAbi,
		client:      client,
		instance:    instance,
	}, nil
}

func (ap *AirdropProvider) ProcessedClaims(ops_id uint) (bool, error) {
	_ops_id := big.NewInt(int64(ops_id))

	return ap.instance.ProcessedClaims(nil, _ops_id)
}

func airdropSigner() (*signer.SignerEIP712, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("Can't read airdrop config err=%s", err.Error())
	}

	chainId := big.NewInt(int64(config.Service.ChainId))

	return signer.NewEIP712Signer(_domainName,
		_version,
		config.Service.L1AirdropAddress,
		chainId,
		config.Service.SignerKeyId)
}

func dialAirdrop(providerUrl, airdropAddress string) (client *ethclient.Client, instance *airdrop_l1.AirdropL1, err error) {
	var sleepFor = 4 * time.Second
	for i := 0; i < 5; i++ {
		client, err = ethclient.Dial(providerUrl)
		if err == nil {
			addr := common.HexToAddress(airdropAddress)
			instance, err = airdrop_l1.NewAirdropL1(addr, client)
			if err == nil {
				break
			} else {
				err = fmt.Errorf("Error creating contract instance: %s", err.Error())
			}
		}
		time.Sleep(sleepFor)
		sleepFor = sleepFor * 2
	}
	if err != nil {
		return nil, nil, fmt.Errorf("Error dialing eth client: %s", err.Error())
	}
	return client, instance, nil
}

func stripPrefix(input string, charsToRemove int) string {
	asRunes := []rune(input)
	return string(asRunes[charsToRemove:])
}
