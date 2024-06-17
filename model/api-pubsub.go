package model

import (
	"context"
)

type (
	ChannelName = string
	PubsubBatch = map[ChannelName][]byte
)

func (api *ApiModel) PublishBatch(ctx context.Context, batch PubsubBatch, ttl, size, meta_ttl int) error {
	_, err := DataResponse[any]{}.Request(ctx, PUBSUB_INSTANCE, api.broker, "publish_data_batch", []any{
		batch,
		ttl,
		size,
		meta_ttl,
	})

	return err
}
