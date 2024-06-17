package api

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/pkg/log"
	neturl "net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const BLAST_AIRDROP_API_MAINNET_URL = "https://waitlist-api.prod.blast.io"
const BLAST_OPERATOR_ADDRESS = "0x9c9543Ec5183c2f9e05E9fCf02f056cEc38620Ca"
const BLAST_POINT_UPDATE_INTERVAL_SECONDS = 600

var blastPointContractAddresses = []string{"0x3Ba925fdeAe6B46d0BB4d424D829982Cb2F7309e", "0x27C11812aF8592375a7553d6034Cc0F189488f4C"}
var cachedResults map[string]BlastPointsResponse
var launcherBlastMutex sync.Mutex
var launcherBlastSyncer bool = false

// blast api internal our API req/response
type BlastPointsRequest struct {
	WalletAddress string `form:"wallet_address" binding:"required"`
}

type blastPoint struct {
	Points                decimal.Decimal `json:"points"`
	PendingPoints         decimal.Decimal `json:"pendingPoints"`
	NextBatchFinalizingAt time.Time       `json:"nextBatchFinalizingAt"`
}

type BlastPointsResponse struct {
	Gold  blastPoint `json:"gold"`
	Blast blastPoint `json:"blast"`
}

// blast API external definitions
type Blast struct {
	apiUrl          string
	operatorAddress string
	contractAddress string
}

type BlastApiChallengeRequest struct {
	OperatorAddress string `json:"operatorAddress"`
	ContractAddress string `json:"contractAddress"`
}

type BlastApiChallengeResponse struct {
	Success       bool   `json:"success"`
	ChallengeData string `json:"challengeData"`
	Message       string `json:"message"`
}

type BlastApiBearerTokenRequest struct {
	ChallengeData string `json:"challengeData"`
	Signature     string `json:"signature"`
}

type BlastApiBearerTokenResponse struct {
	Success     bool   `json:"success"`
	BearerToken string `json:"bearerToken"`
}

type BlastApiBatchTransfer struct {
	ToAddress string `json:"toAddress"`
	Points    string `json:"points"`
}

type BlastApiBatch struct {
	ContractAddress string                  `json:"contractAddress"`
	Id              string                  `json:"id"`
	PointType       string                  `json:"pointType"`
	CreatedAt       string                  `json:"createdAt"`
	FinalizeAt      time.Time               `json:"finalizeAt"`
	UpdatedAt       time.Time               `json:"updatedAt"`
	Status          string                  `json:"status"`
	Points          string                  `json:"points"`
	TransferCount   uint64                  `json:"transferCount"`
	Transfers       []BlastApiBatchTransfer `json:"transfers"`
}

type BlastApiBatchesResponse struct {
	Success bool            `json:"success"`
	Batches []BlastApiBatch `json:"batches"`
	Cursor  *string         `json:"cursor"`
}

type BlastApiBatchResponse struct {
	Success bool          `json:"success"`
	Batch   BlastApiBatch `json:"batch"`
}

func NewBlast(contractAddress string) *Blast {
	return &Blast{
		apiUrl:          BLAST_AIRDROP_API_MAINNET_URL,
		operatorAddress: BLAST_OPERATOR_ADDRESS,
		contractAddress: contractAddress,
	}
}

func (c *Blast) buildURL(path ...string) (string, error) {
	return neturl.JoinPath(c.apiUrl, path...)
}

func (c *Blast) signMessage(message string, key string) (string, error) {
	privKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return "", err
	}

	eip191Message := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(eip191Message))
	signature, err := crypto.Sign(hash.Bytes(), privKey)

	signed := hexutil.Encode(signature)
	return signed, nil
}

func (c *Blast) obtainChallenge() (*BlastApiChallengeResponse, error) {
	url, err := c.buildURL("/v1/dapp-auth/challenge")
	if err != nil {
		return nil, err
	}

	client := req.C()
	client.DisableDumpAll()

	var resp BlastApiChallengeResponse
	rr := client.SetBaseURL(url)
	err = rr.Post().
		SetBody(&BlastApiChallengeRequest{
			OperatorAddress: c.operatorAddress,
			ContractAddress: c.contractAddress}).
		Do().
		Into(&resp)

	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, errors.New("unable to obtain challenge")
	}

	return &resp, nil
}

func (c *Blast) obtainBearerToken(challengeResponse *BlastApiChallengeResponse) (*BlastApiBearerTokenResponse, error) {
	url, err := c.buildURL("/v1/dapp-auth/solve")
	if err != nil {
		return nil, err
	}

	privKey := os.Getenv("BLAST_OPERATOR_PRIV_KEY")
	if privKey == "" {
		return nil, fmt.Errorf("no env BLAST_OPERATOR_PRIV_KEY = %s", privKey)
	}
	signature, err := c.signMessage(challengeResponse.Message, privKey)
	if err != nil {
		return nil, err
	}

	client := req.C()
	client.DisableDumpAll()

	var resp BlastApiBearerTokenResponse
	rr := client.SetBaseURL(url)
	err = rr.Post().
		SetBody(&BlastApiBearerTokenRequest{
			ChallengeData: challengeResponse.ChallengeData,
			Signature:     signature}).
		Do().
		Into(&resp)

	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, errors.New("unable to obtain bearer token")
	}

	return &resp, nil
}

func (c *Blast) getBatches(bearerToken string) ([]BlastApiBatch, error) {
	url, err := c.buildURL("/v1/contracts", c.contractAddress, "batches")
	if err != nil {
		return nil, err
	}

	client := req.C()
	client.DisableDumpAll()

	result := make([]BlastApiBatch, 0)
	cursor := ""

	for {
		uri := url
		if cursor != "" {
			uri = uri + "?cursor=" + cursor
		}

		var resp BlastApiBatchesResponse
		rr := client.SetBaseURL(uri)
		err = rr.SetCommonBearerAuthToken(bearerToken).
			Get().
			Do().
			Into(&resp)

		if err != nil {
			return nil, err
		}

		if !resp.Success {
			return nil, errors.New("unable to get batches")
		}

		result = append(result, resp.Batches...)

		if resp.Cursor == nil {
			// no more pagination required, exit
			break
		}

		cursor = *resp.Cursor
	}

	return result, nil
}

func (c *Blast) getBatch(batchId string, bearerToken string) (*BlastApiBatchResponse, error) {
	url, err := c.buildURL("/v1/contracts", c.contractAddress, "batches", batchId)
	if err != nil {
		return nil, err
	}

	client := req.C()
	client.DisableDumpAll()

	var resp BlastApiBatchResponse
	rr := client.SetBaseURL(url)
	err = rr.SetCommonBearerAuthToken(bearerToken).
		Get().
		Do().
		Into(&resp)

	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, errors.New("unable to get batch")
	}

	return &resp, nil

}

func getBlastPoints() (map[string]BlastPointsResponse, error) {
	tmpResults := make(map[string]BlastPointsResponse)

	for _, blastContractAddress := range blastPointContractAddresses {

		blast := NewBlast(blastContractAddress)
		challenge, err := blast.obtainChallenge()
		if err != nil {
			return nil, err
		}

		token, err := blast.obtainBearerToken(challenge)
		if err != nil {
			return nil, err
		}

		batches, err := blast.getBatches(token.BearerToken)
		if err != nil {
			return nil, err
		}

		for _, batchToFetch := range batches {
			batch, err := blast.getBatch(batchToFetch.Id, token.BearerToken)
			if err != nil {
				return nil, err
			}

			if batch.Batch.Status == "FINALIZING" || batch.Batch.Status == "PENDING" {
				finalizeAt := batch.Batch.FinalizeAt
				for _, transfer := range batch.Batch.Transfers {
					walletAddress := strings.ToLower(transfer.ToAddress)
					v, ok := tmpResults[walletAddress]
					if !ok {
						tmpResults[walletAddress] = BlastPointsResponse{}
						v = tmpResults[walletAddress]
					}

					if batch.Batch.PointType == "LIQUIDITY" {
						d, _ := decimal.NewFromString(transfer.Points)
						v.Blast.PendingPoints = v.Blast.PendingPoints.Add(d)
						if finalizeAt.Unix() > v.Blast.NextBatchFinalizingAt.Unix() {
							v.Blast.NextBatchFinalizingAt = finalizeAt
						}
					} else if batch.Batch.PointType == "DEVELOPER" {
						d, _ := decimal.NewFromString(transfer.Points)
						v.Gold.PendingPoints = v.Gold.PendingPoints.Add(d)

						if finalizeAt.Unix() > v.Blast.NextBatchFinalizingAt.Unix() {
							v.Gold.NextBatchFinalizingAt = finalizeAt
						}
					}

					tmpResults[walletAddress] = v
				}
			} else if batch.Batch.Status == "FINALIZED" {
				finalizeAt := batch.Batch.FinalizeAt
				for _, transfer := range batch.Batch.Transfers {
					walletAddress := strings.ToLower(transfer.ToAddress)
					v, ok := tmpResults[walletAddress]
					if !ok {
						tmpResults[walletAddress] = BlastPointsResponse{}
						v = tmpResults[walletAddress]
					}

					if batch.Batch.PointType == "LIQUIDITY" {
						d, _ := decimal.NewFromString(transfer.Points)
						v.Blast.Points = v.Blast.Points.Add(d)
						if finalizeAt.Unix() > v.Blast.NextBatchFinalizingAt.Unix() {
							v.Blast.NextBatchFinalizingAt = finalizeAt
						}
					} else if batch.Batch.PointType == "DEVELOPER" {
						d, _ := decimal.NewFromString(transfer.Points)
						v.Gold.Points = v.Gold.Points.Add(d)
						if finalizeAt.Unix() > v.Blast.NextBatchFinalizingAt.Unix() {
							v.Gold.NextBatchFinalizingAt = finalizeAt
						}
					}

					tmpResults[walletAddress] = v
				}
			}
		}
	}

	return tmpResults, nil
}

func runBlastPointsUpdater() {
	for {
		tmpResults, err := getBlastPoints()
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Error getting blast points: err=%v", err)
		} else {
			cachedResults = tmpResults
		}

		time.Sleep(BLAST_POINT_UPDATE_INTERVAL_SECONDS * time.Second)
	}
}

func HandleBlastPoints(c *gin.Context) {
	launcherBlastMutex.Lock()
	defer launcherBlastMutex.Unlock()

	if launcherBlastSyncer == false {
		go runBlastPointsUpdater()
		launcherBlastSyncer = true
	}

	var request BlastPointsRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	walletAddress := strings.ToLower(request.WalletAddress)
	v, ok := cachedResults[walletAddress]
	if !ok {
		ErrorResponse(c, errors.New("no such wallet"))
		return
	}

	SuccessResponse(c, v)
}
