package archiver

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
)

type anyMapJsonbCodec struct {
	pgtype.JSONBCodec
}

func (c *anyMapJsonbCodec) PlanEncode(m *pgtype.Map, oid uint32, format int16, value any) pgtype.EncodePlan {
	return &anyMapPlan{
		EncodePlan: c.JSONBCodec.PlanEncode(m, oid, format, value),
	}
}

type anyMapPlan struct {
	pgtype.EncodePlan
}

func (p *anyMapPlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	if anyMap, ok := value.(map[any]any); ok {
		strMap := make(map[string]any)
		for anyKey, anyVal := range anyMap {
			strKey, ok := anyKey.(string)
			if !ok {
				return nil, errors.New("map key should be string")
			}
			strMap[strKey] = anyVal
		}
		value = strMap
	}
	return p.EncodePlan.Encode(value, buf)
}
