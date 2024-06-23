package model

import (
	"context"
	"fmt"
	"strings"
)

const (
	IS_VALID_SIGNER = "profile.is_valid_signer"
	ADD_PERMISSION  = "profile.add_permission"
)

func (api *ApiModel) IsValidSigner(ctx context.Context, vault, wallet string, requireRole uint) error {
	vault = strings.ToLower(vault)
	wallet = strings.ToLower(wallet)

	isValid, err := DataResponse[bool]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, IS_VALID_SIGNER, []interface{}{
		vault,
		wallet,
		requireRole,
	})

	if err != nil {
		return err
	} else if !isValid {
		return fmt.Errorf("requiredRole=%d on Vault=%s for wallet%s not allowed", requireRole, vault, wallet)
	}

	return nil
}

func (api *ApiModel) AddPermission(ctx context.Context, vault, wallet string, requireRole uint) error {
	vault = strings.ToLower(vault)
	wallet = strings.ToLower(wallet)

	_, err := DataResponse[interface{}]{}.Request(ctx, PROFILE_INSTANCE, api.broker, ADD_PERMISSION, []interface{}{
		vault,
		wallet,
		requireRole,
	})

	return err
}
