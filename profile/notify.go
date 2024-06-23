package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/eapache/go-resiliency/retrier"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	centrifugo "github.com/strips-finance/rabbit-dex-backend/pkg/centrifugo"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

var (
	DefaultNotifyClientOptions = NotifyClientOptions{
		Retries: 1,
	}
)

type NotifyClientOptions struct {
	Retries    uint
	RetryDelay time.Duration
}

type NotifyClient struct {
	client  centrifugo.CentrifugoApiClient
	backoff *retrier.Retrier
}

func NewNotifyClient(address string, options NotifyClientOptions) (*NotifyClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "NewNotifyClient")
	}

	return &NotifyClient{
		client: centrifugo.NewCentrifugoApiClient(conn),
		backoff: retrier.New(retrier.ConstantBackoff(int(options.Retries), options.RetryDelay), nil),
	}, nil
}

func (c *NotifyClient) PublishExtendedProfiles(ctx context.Context, data []*model.ExtendedProfileTierStatusData) error {
	req := centrifugo.BatchRequest{}

	for i, p := range data {
		channel := fmt.Sprint("account@", p.ProfileID)

		encoded, err := json.Marshal(p)
		if err != nil {
			return errors.Wrapf(err, "Marshal ProfileCache: profile-id=%d", p.ProfileID)
		}

		req.Commands = append(req.Commands, &centrifugo.Command{
			Id: uint32(i),
			Method: centrifugo.Command_PUBLISH,
			Publish: &centrifugo.PublishRequest{
				Channel: channel,
				Data: encoded,
				SkipHistory: true,
			},
		})
	}

	return c.backoff.Run(func() error {
		resp, err := c.client.Batch(ctx, &req)
		if err != nil {
			return errors.Wrap(err, "PublishExtendedProfiles.Batch")
		}
		for _, r := range resp.Replies {
			if e := r.Error; e != nil {
				return fmt.Errorf("PublishExtendedProfiles.Batch: %s, reply-id=%d", e.GetMessage(), r.GetId())
			}
		}

		return nil
	})
}
