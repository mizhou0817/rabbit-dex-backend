package volume

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/model"

	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/dbtestsuite"
)

const (
	wallet            = "0x1a05a1507c35c763035fdb151af9286d4d90a81b"
	traderID1         = uint64(1)
	traderProfileType = "trader"
)

var tables = []string{
	"app_profile",
	"app_fill",
}

type dbTestSuite struct {
	dbtestsuite.DBTestSuite
}

func Test_dbTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timescaledb integration test")
	}
	testSuite := new(dbTestSuite)
	suite.Run(t, testSuite)
}

func (s *dbTestSuite) SetupSuite() {
	s.BaseSetupSuite()
	ApplyTestMigrations(s.T(), s.MigrationConnectionString())
}

func (s *dbTestSuite) TearDownSuite() {
	s.BaseTearDownSuite()
}

func (s *dbTestSuite) SetupTest() {
	s.DeleteTables(tables)
}

type handlerFunc = func(
	ctx context.Context,
	db *pgxpool.Pool,
	request VolumeRequest,
) (*VolumeResponse, error)

func (s *dbTestSuite) TestHandleVolume() {
	now := time.Now()

	tests := []struct {
		exchange string
		handler  handlerFunc
	}{
		{model.EXCHANGE_BFX, HandleBfxVolume},
		{model.EXCHANGE_RBX, HandleRbxVolume},
	}
	for _, ttt := range tests {
		s.Run(
			ttt.exchange, func() {

				insertCommon := func() {
					s.insertProfile(traderID1, traderProfileType, wallet, ttt.exchange)
					s.insertFill(
						traderID1,
						now,
						decimal.NewFromFloat(5.0),
						decimal.NewFromFloat(20.0),
					)
				}

				type args struct {
					request VolumeRequest
				}
				tests := []struct {
					name       string
					args       args
					insertData func()
					want       *VolumeResponse
					wantErr    bool
				}{
					{
						name: "should return volume when 1 fill",
						args: args{
							request: VolumeRequest{
								StartDate: now.Add(-time.Hour).UnixMicro(),
								EndDate:   now.Add(+time.Hour).UnixMicro(),
							},
						},
						insertData: insertCommon,
						want: &VolumeResponse{
							Volume: decimal.NewFromFloat(100.00),
						},
					},
					{
						name: "should return zero volume when no fills",
						args: args{
							request: VolumeRequest{
								StartDate: now.Add(+time.Hour).UnixMicro(),
								EndDate:   now.Add(2 * time.Hour).UnixMicro(),
							},
						},
						insertData: insertCommon,
						want: &VolumeResponse{
							Volume: decimal.NewFromFloat(0.0),
						},
					},
				}
				for _, tt := range tests {
					s.Run(
						tt.name, func() {
							s.DeleteTables(tables)
							tt.insertData()

							// when
							got, err := ttt.handler(
								context.Background(),
								s.GetDB(),
								tt.args.request,
							)

							// then
							if tt.wantErr {
								s.NotNil(err)
								s.Nil(got)
							} else {
								s.NoError(err)
								s.assertJSON(tt.want, got)
							}
						},
					)
				}
			},
		)
	}
}

func (s *dbTestSuite) assertJSON(expected, got any) {
	expectedB, err := json.Marshal(expected)
	s.NoError(err)
	gotB, err := json.Marshal(got)
	s.NoError(err)
	s.Equal(string(expectedB), string(gotB))

}

func (s *dbTestSuite) insertProfile(profileID uint64, profileType, wallet, exchangeId string) {
	q := `INSERT INTO app_profile
	(id, profile_type, status, wallet, created_at, shard_id, archive_id, archive_timestamp, exchange_id)
	VALUES(@id, @profileType, 'active',@wallet, @now, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_profile), @now, @exchangeId)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":          profileID,
		"profileType": profileType,
		"wallet":      wallet,
		"exchangeId":  exchangeId,
		"now":         now,
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertFill(profileID uint64, now time.Time, price, size decimal.Decimal) {
	q := `INSERT INTO app_fill
	(id, profile_id, market_id, order_id, "timestamp", trade_id, price, "size", side, is_maker, fee, liquidation, shard_id, archive_id, archive_timestamp, client_order_id)
	VALUES(@id, @profile_id, '', '', @timestamp, '', @price, @size, '', false, 0, false, '', (select COALESCE(max(archive_id),0) + 1 from app_fill), @archive_timestamp, '')
	`
	args := pgx.NamedArgs{
		"id":                uuid.NewString(),
		"profile_id":        profileID,
		"timestamp":         now.UnixMicro(),
		"price":             price,
		"size":              size,
		"archive_timestamp": time.Now().Add(-100 * time.Hour).UnixMicro(),
	}
	s.Execute(q, args)
}
