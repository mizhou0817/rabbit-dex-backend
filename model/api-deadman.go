package model

import (
	"context"
)

const (
	DEADMAN_CREATE = "public.deadman_create"
	DEADMAN_DELETE = "public.deadman_delete"
	DEADMAN_LIST   = "public.deadman_get"
)

type DeadmanData struct {
	ProfileId   uint   `msgpack:"profile_id"  json:"profile_id,omitempty"`
	Timeout     uint   `msgpack:"timeout"  json:"timeout,omitempty"`
	LastUpdated uint   `msgpack:"last_updated"  json:"last_updated,omitempty"`
	Status      string `msgpack:"status"  json:"status,omitempty"`
}

func (api *ApiModel) DeadmanCreate(ctx context.Context, profile_id, timeout uint) (*DeadmanData, error) {
	return DataResponse[*DeadmanData]{}.Request(ctx, API_INSTANCE, api.broker, DEADMAN_CREATE, []interface{}{
		profile_id,
		timeout,
	})
}

func (api *ApiModel) DeadmanDelete(ctx context.Context, profile_id uint) (*DeadmanData, error) {
	return DataResponse[*DeadmanData]{}.Request(ctx, API_INSTANCE, api.broker, DEADMAN_DELETE, []interface{}{
		profile_id,
	})
}

func (api *ApiModel) DeadmanGet(ctx context.Context, profile_id uint) (*DeadmanData, error) {
	return DataResponse[*DeadmanData]{}.Request(ctx, API_INSTANCE, api.broker, DEADMAN_LIST, []interface{}{
		profile_id,
	})
}
