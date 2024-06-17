package profiledata

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileData struct {
	Version uint           `json:"version"`
	Data    map[string]any `binding:"required" json:"data"`
}

type Error string

func (e Error) Error() string { return string(e) }

const (
	errVersionMismatch = Error("version in db does not match version in request")
	errDB              = Error("db operation error")
)

type Storage struct {
	profileID uint
	db        *pgxpool.Pool
}

func NewStorage(profileID uint, db *pgxpool.Pool) *Storage {
	return &Storage{
		profileID: profileID,
		db:        db,
	}
}

var sqlBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Get retrieves the profile data from the storage.
// It returns a pointer to ProfileData and an error, if any.
func (s *Storage) Get(ctx context.Context) (*ProfileData, error) {
	data := ProfileData{
		Version: 0,
		Data:    map[string]any{},
	}

	sql, args := sqlBuilder.
		Select("version", "data").
		From("app_profile_data").
		Where(sq.Eq{"profile_id": s.profileID}).
		MustSql()

	err := s.db.QueryRow(ctx, sql, args...).Scan(&data.Version, &data.Data)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("%w: %w", errDB, err)
	}

	return &data, nil
}

// Replace replaces the existing profile data with the provided data.
// It compares and swaps the version field to avoid race conditions.
// If the version of the existing data does not match the provided version,
// it returns existing data and an error.
// Otherwise, it updates the version and data fields in the database and returns
// the updated profile data.
//
// Parameters:
//   - ctx: The context.Context object for the request.
//   - data: The ProfileData object containing the new data to be replaced.
//
// Returns:
//   - *ProfileData: A pointer to the existing profile data after the replace operation
//     if the version matches the provided version. Otherwise, it returns
//     the existing data.
//   - error: An error if any occurred during the replace operation.
func (s *Storage) Replace(ctx context.Context, data ProfileData) (*ProfileData, error) {
	sql, args := sqlBuilder.
		// Compare and swap on Version field.
		Insert("app_profile_data").
		Columns("profile_id", "version", "data").
		Values(s.profileID, 1, data.Data). // Only insert version == 1 otherwise increment
		Suffix(`ON CONFLICT (profile_id) DO
					UPDATE SET
						version = ? + 1,
						data = ?
					WHERE
						app_profile_data.version = ?
				RETURNING version, data`,
			data.Version, data.Data, data.Version).
		MustSql()

	var res ProfileData
	err := s.db.QueryRow(ctx, sql, args...).Scan(&res.Version, &res.Data)
	if err == pgx.ErrNoRows {
		existing, err := s.Get(ctx)
		if err != nil {
			return nil, err
		}
		return existing, errVersionMismatch
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errDB, err)
	}
	return &res, nil
}
