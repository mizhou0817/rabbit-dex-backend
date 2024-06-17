package tsdb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

type ReferralLink struct {
	ProfileId ProfileId `db:"profile_id"`
	InvitedId ProfileId `db:"invited_id"`
}

func (s *Store) GetReferralsByInvitedProfiles(ctx context.Context, invitedIds ...ProfileId) ([]ReferralLink, error) {
	if len(invitedIds) == 0 {
		return s.getReferralsByInvitedProfiles(ctx)
	}

	var (
		referrals []ReferralLink
	)
	for i, n, size := 0, batchSize, len(invitedIds); i < size; i += n {
		if i+n > size {
			n = size - i
		}

		invitedIdsBatch := invitedIds[i : i+n]
		refs, err := s.getReferralsByInvitedProfiles(ctx, invitedIdsBatch...)
		if err != nil {
			return nil, errors.Wrap(err, "getReferralsByInvitedProfiles")
		}

		for _, r := range refs {
			referrals = append(referrals, r)
		}
	}

	return referrals, nil
}

func (s *Store) getReferralsByInvitedProfiles(ctx context.Context, invitedIds ...ProfileId) ([]ReferralLink, error) {
	columns := []string{"profile_id", "invited_id"}

	builder := s.builder.Select(columns...).From("app_referral_link")
	if len(invitedIds) > 0 {
		builder = builder.Where(sq.Eq{"invited_id": invitedIds})
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

	var referralLinks []ReferralLink
	for rows.Next() {
		rl := ReferralLink{}

		err := rows.Scan(
			&rl.ProfileId,
			&rl.InvitedId,
		)
		if err != nil {
			return nil, errors.Wrap(err, "scan rows")
		}

		referralLinks = append(referralLinks, rl)
	}

	return referralLinks, nil
}
