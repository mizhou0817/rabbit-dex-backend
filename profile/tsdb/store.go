package tsdb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

const (
	batchSize = 50000
)

type Store struct {
	db      DB
	builder sq.StatementBuilderType
}

func NewStore(db DB) *Store {
	return &Store{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *Store) GetProfilesIdsAfterCreatedAt(ctx context.Context, afterTsMicro int64) ([]ProfileId, error) {
	columns := []string{"id"}

	builder := s.builder.Select(columns...).From("app_profile")
	if afterTsMicro > 0 {
		builder = builder.Where(sq.Gt{"created_at": afterTsMicro})
	}
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build query")
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "run query")
	}
	defer rows.Close()

	var profilesIds []ProfileId
	for rows.Next() {
		var profileId ProfileId

		err := rows.Scan(&profileId)
		if err != nil {
			return nil, errors.Wrap(err, "scan rows")
		}

		profilesIds = append(profilesIds, profileId)
	}

	return profilesIds, nil
}
