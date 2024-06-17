//go:generate go run github.com/golang/mock/mockgen -source=$PWD/volume_cache.go -destination=$PWD/mock/volume_cache.go -package=mock
package profile

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/strips-finance/rabbit-dex-backend/profile/tsdb"
)

type VolumeStore interface {
	GetVolumesAggregatesLast30d(context.Context) ([]tsdb.CumVolume, error)
}

type VolumeStoreCache struct {
	store VolumeStore

	mu      sync.RWMutex
	volumes map[ProfileId]decimal.Decimal
}

func NewVolumeStoreCache(store VolumeStore) *VolumeStoreCache {
	return &VolumeStoreCache{
		store:   store,
		volumes: map[ProfileId]decimal.Decimal{},
	}
}

func (c *VolumeStoreCache) Refresh(ctx context.Context) error {
	vols, err := c.store.GetVolumesAggregatesLast30d(ctx)
	if err != nil {
		return errors.Wrap(err, "GetVolumesAggregatesLast30d")
	}

	volumes := make(map[ProfileId]decimal.Decimal, len(vols))
	for _, p := range vols {
		volumes[p.ProfileId] = p.CumVolume
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.volumes = volumes

	return nil
}

func (c *VolumeStoreCache) GetVolume(profileId ProfileId) (decimal.Decimal, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	vol, ok := c.volumes[profileId]
	return vol, ok
}
