package portfolio

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/dbtestsuite"
)

const (
	wallet            = "0x1a05a1507c35c763035fdb151af9286d4d90a81b"
	traderID          = uint64(1)
	traderProfileType = "trader"
)

var tables = []string{
	"app_profile_cache",
	"app_profile_cache_1m",
	"app_profile_cache_15m",
	"app_profile_cache_30m",
	"app_profile_cache_1h",
	"app_profile_cache_1d",
	"app_profile_cache_1d",
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

func (s *dbTestSuite) TestHandlePortfolioList() {
	timeAtFuture, err := time.Parse(time.RFC3339, "2035-01-02T15:04:05Z")
	s.NoError(err)
	timeRounded1m, err := time.Parse(time.RFC3339, "2035-01-02T15:04:00Z")
	s.NoError(err)
	timeRounded15m, err := time.Parse(time.RFC3339, "2035-01-02T15:00:00Z")
	s.NoError(err)
	timeRounded30m, err := time.Parse(time.RFC3339, "2035-01-02T15:00:00Z")
	s.NoError(err)
	timeRounded1h, err := time.Parse(time.RFC3339, "2035-01-02T15:00:00Z")
	s.NoError(err)
	timeRounded1d, err := time.Parse(time.RFC3339, "2035-01-02T00:00:00Z")
	s.NoError(err)

	now := time.Now()
	ts := now.UnixMicro()
	item := PortfolioData{
		Time:  ts,
		Value: decimal.NewFromFloat(10.25),
	}

	type args struct {
		requestRange string
		profileId    uint64
	}

	tests := []struct {
		name       string
		insertData func()
		args       args
		want       []PortfolioData
		wantErr    bool
	}{
		{
			name: "should return empty list when 1h",
			args: args{
				requestRange: "1h",
				profileId:    traderID,
			},
			want: []PortfolioData{},
		},
		{
			name: "should return empty list when 1d",
			args: args{
				requestRange: "1d",
				profileId:    traderID,
			},
			want: []PortfolioData{},
		},
		{
			name: "should return empty list when 1w",
			args: args{
				requestRange: "1w",
				profileId:    traderID,
			},
			want: []PortfolioData{},
		},
		{
			name: "should return empty list when 1m (one month)",
			args: args{
				requestRange: "1m",
				profileId:    traderID,
			},
			want: []PortfolioData{},
		},
		{
			name: "should return empty list when 1y",
			args: args{
				requestRange: "1y",
				profileId:    traderID,
			},
			want: []PortfolioData{},
		},
		{
			name: "should return empty list when all",
			args: args{
				requestRange: "all",
				profileId:    traderID,
			},
			want: []PortfolioData{},
		},
		{
			name: "should return 1 item when 1h",
			args: args{
				requestRange: "1h",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCachePeriod("1m", traderID, item)
			},
			want: []PortfolioData{item},
		},
		{
			name: "should return 1 item when 1d",
			args: args{
				requestRange: "1d",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCachePeriod("15m", traderID, item)
			},
			want: []PortfolioData{item},
		},
		{
			name: "should return 1 item when 1w",
			args: args{
				requestRange: "1w",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCachePeriod("30m", traderID, item)
			},
			want: []PortfolioData{item},
		},
		{
			name: "should return 1 item when 1m (one month)",
			args: args{
				requestRange: "1m",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCachePeriod("1h", traderID, item)
			},
			want: []PortfolioData{item},
		},
		{
			name: "should return 1 item when 1y",
			args: args{
				requestRange: "1y",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCachePeriod("1d", traderID, item)
			},
			want: []PortfolioData{item},
		},
		{
			name: "should return 1 item when all",
			args: args{
				requestRange: "all",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCachePeriod("1d", traderID, item)
			},
			want: []PortfolioData{item},
		},
		{
			name: "should generate 1m (one minute) range 1h",
			args: args{
				requestRange: "1h",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded1m.UnixMicro(),
					Value: item.Value,
				},
			},
		},
		{
			name: "should update the last price 1m (one minute) range 1h",
			args: args{
				requestRange: "1h",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, decimal.NewFromFloat(1.0), timeAtFuture.Add(-time.Second).UnixMicro())
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded1m.UnixMicro(),
					Value: item.Value,
				},
			},
		},
		{
			name: "should generate 15m range 1d",
			args: args{
				requestRange: "1d",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded15m.UnixMicro(),
					Value: item.Value,
				},
			},
		},
		{
			name: "should update the last price 15m range 1d",
			args: args{
				requestRange: "1d",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, decimal.NewFromFloat(1.0), timeAtFuture.Add(-time.Second).UnixMicro())
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded15m.UnixMicro(),
					Value: item.Value,
				},
			},
		},
		{
			name: "should generate 30m range 1w",
			args: args{
				requestRange: "1w",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded30m.UnixMicro(),
					Value: item.Value,
				},
			},
		},
		{
			name: "should update the last price 30m range 1w",
			args: args{
				requestRange: "1w",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, decimal.NewFromFloat(1.0), timeAtFuture.Add(-time.Second).UnixMicro())
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded30m.UnixMicro(),
					Value: item.Value,
				},
			},
		},
		{
			name: "should generate 1h range 1m (one month)",
			args: args{
				requestRange: "1m",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded30m.UnixMicro(),
					Value: item.Value,
				},
			},
		},
		{
			name: "should update the last price 1h range 1m (one month)",
			args: args{
				requestRange: "1m",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, decimal.NewFromFloat(1.0), timeAtFuture.Add(-time.Second).UnixMicro())
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded1h.UnixMicro(),
					Value: item.Value,
				},
			},
		},
		{
			name: "should generate 1d range 1y",
			args: args{
				requestRange: "1y",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded1d.UnixMicro(),
					Value: item.Value,
				},
			},
		},
		{
			name: "should update the last price 1d range 1y",
			args: args{
				requestRange: "1y",
				profileId:    traderID,
			},
			insertData: func() {
				s.insertProfileCache(traderID, traderProfileType, wallet, decimal.NewFromFloat(1.0), timeAtFuture.Add(-time.Second).UnixMicro())
				s.insertProfileCache(traderID, traderProfileType, wallet, item.Value, timeAtFuture.UnixMicro())
			},
			want: []PortfolioData{
				{
					Time:  timeRounded1d.UnixMicro(),
					Value: item.Value,
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// clean
			s.DeleteTables(tables)

			// given
			if tt.insertData != nil {
				tt.insertData()
			}

			// when
			request := PortfolioRequest{
				Range: tt.args.requestRange,
			}
			got, err := HandlePortfolioList(context.Background(), s.GetDB(), request, uint(tt.args.profileId))

			// then
			if tt.wantErr {
				s.Error(err)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tt.want, got)
			}
		})
	}
}

func (s *dbTestSuite) assertJSON(expected, got any) {
	expectedB, err := json.Marshal(expected)
	s.NoError(err)
	gotB, err := json.Marshal(got)
	s.NoError(err)
	s.Equal(string(expectedB), string(gotB))

}

func (s *dbTestSuite) insertProfileCache(vaultID uint64, profileType, wallet string, accountEquity decimal.Decimal, archiveTimestamp int64) {
	q := `INSERT INTO app_profile_cache
	(id, profile_type, status, wallet, last_update, balance, account_equity, total_position_margin, total_order_margin, total_notional, account_margin, withdrawable_balance, cum_unrealized_pnl, health, account_leverage, cum_trading_volume, leverage, last_liq_check, shard_id, archive_id, archive_timestamp)
	VALUES(@id, @profileType, 'active',@wallet, @now, 0, @accountEquity, 0, 0, 0, 0, 0, 0, 0, 0, 0, '{}', 0, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_profile), @archive_timestamp)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":                vaultID,
		"profileType":       profileType,
		"wallet":            wallet,
		"now":               now,
		"accountEquity":     accountEquity,
		"archive_timestamp": archiveTimestamp,
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertProfileCachePeriod(period string, id uint64, item PortfolioData) {
	q := fmt.Sprintf(`INSERT INTO app_profile_cache_%s
	(id, archive_timestamp, account_equity)
	VALUES(@id, @archive_timestamp, @account_equity)
	`, period)
	args := pgx.NamedArgs{
		"id":                id,
		"archive_timestamp": item.Time,
		"account_equity":    item.Value,
	}
	s.Execute(q, args)
}
