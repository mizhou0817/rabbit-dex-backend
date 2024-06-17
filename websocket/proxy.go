package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type MarketRequest struct {
	Channel string `json:"channel"`
}

type SubscriptionRequest struct {
	Channel string `json:"channel"`
	User    string `json:"user"`
}

type jsonMessageParams struct {
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

type jsonMessage struct {
	Method string            `json:"method"`
	Params jsonMessageParams `json:"params"`
}

func publishToCentrifugoChannel(channel string, marketResponse api.MarketResponse) error {
	url := "http://centrifugo:8000/api?api_key=T7gQqqloMObC30i5MNwjZYPb60Y7E6pFesONxB5s5YE"

	msg := jsonMessage{
		Method: "publish",
		Params: jsonMessageParams{
			Channel: channel,
			Data:    marketResponse,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func sendMarketViewData(db *pgxpool.Pool, sql string) error {
	resp := api.MarketResponse{}
	rows, err := db.Query(context.Background(), sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var channel string
		err = rows.Scan(
			&channel,
			&resp.AverageDailyVolume,
			&resp.LastTradePrice24High,
			&resp.LastTradePrice24Low,
			&resp.LastTradePrice24hChangePremium,
			&resp.LastTradePrice24hChangeBasis,
			&resp.AverageDailyVolumeChangePremium,
			&resp.AverageDailyVolumeChangeBasis)

		if err != nil {
			return err
		}
		channel = "market:" + channel
		err = publishToCentrifugoChannel(channel, resp)
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func runMarketViewDataPusher() error {
	db, err := pgxpool.New(context.Background(), "postgres://rabbitx:rabbitx@timescaledb:5432/rabbitx")
	if err != nil {
		return err
	}
	defer db.Close()

	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, _, err := sqlBuilder.
		Select(
			"market_id",
			"COALESCE(average_daily_volume, 0)",
			"COALESCE(last_trade_price_24high, 0)",
			"COALESCE(last_trade_price_24low, 0)",
			"COALESCE(last_trade_price_24h_change_premium, 0)",
			"COALESCE(last_trade_price_24h_change_basis, 0)",
			"COALESCE(average_daily_volume_change_premium, 0)",
			"COALESCE(average_daily_volume_change_basis, 0)").
		From("market_data_view").
		ToSql()

	if err != nil {
		return err
	}

	for {
		// Call the function to execute
		err := sendMarketViewData(db, sql)
		if err != nil {
			logrus.Info("runMarketViewDataPusher:", err)
		}

		// Wait for 1 seconds
		time.Sleep(1 * time.Second)
	}
}

func RunProxy() {
	logrus.Info("WebSocket Proxy started")

	broker, err := model.GetBroker()
	if err != nil {
		logrus.Panic(err)
	}
	apiModel := model.NewApiModel(broker)

	go runMarketViewDataPusher()

	router := gin.Default()
	router.POST("/centrifugo/subscribe", func(c *gin.Context) {
		var request SubscriptionRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			logrus.Error(err)
			c.Status(http.StatusBadRequest)
			c.JSON(http.StatusOK, gin.H{
				"error": gin.H{},
			})
			return
		}

		logrus.
			WithField("channel", request.Channel).
			WithField("user", request.User).
			Info("new request")

		var (
			data      interface{}
			channel   string
			profileId uint
			argument  string
		)

		if strings.Contains(request.Channel, "@") {
			parts := strings.Split(request.Channel, "@")
			profileId_, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				logrus.Error(err)
				c.Status(http.StatusBadRequest)
				return
			}
			profileId = uint(profileId_)

			requiredId, err := strconv.Atoi(request.User)
			if err != nil {
				logrus.Error(err)
				c.Status(http.StatusBadRequest)
				return
			}

			if uint(requiredId) != profileId {
				logrus.Error("ERROR_PROFILE_ID")
				api.ErrorResponse(c, errors.New("ERROR_PROFILE_ID"))
				c.Abort()
				return
			}

			request.Channel = parts[0]
		}

		if strings.Contains(request.Channel, ":") {
			parts := strings.Split(request.Channel, ":")
			argument = parts[len(parts)-1]
			channel = parts[0]

			if argument == "TEST-MARKET" {
				c.Status(http.StatusBadRequest)
				return
			}
		} else {
			channel = request.Channel
		}

		switch channel {
		case "market":
			market, err := apiModel.GetMarketData(context.Background(), argument)
			if err != nil {
				logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("apiModel.GetMarketData err = %s", err.Error())
				c.Status(http.StatusBadRequest)
				return
			} else if market == nil {
				logrus.WithField(log.AlertTag, log.AlertHigh).Error("market is nil")
				c.Status(http.StatusBadRequest)
				return
			}

			logrus.
				Info("Sending initial state for market")

			data = market

		case "account":
			profile, err := apiModel.GetExtendedProfileData(context.Background(), profileId)
			if err != nil {
				logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("apiModel.GetProfileData err = %s", err.Error())
				c.Status(http.StatusBadRequest)
				return
			} else if profile == nil {
				logrus.WithField(log.AlertTag, log.AlertHigh).Error("profile is nil")
				c.Status(http.StatusBadRequest)
				return
			}

			logrus.
				WithField("id", profile.ProfileID).
				WithField("wallet", profile.Wallet).
				Info("Sending initial state for profile")

			data = profile

		case "orderbook":
			orderbook, err := apiModel.GetOrderbookData(context.Background(), argument)
			if err != nil {
				logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("apiModel.GetOrderbookData err = %s", err.Error())
				c.Status(http.StatusBadRequest)
				return
			} else if orderbook == nil {
				logrus.Error("orderbook is nil")
				c.Status(http.StatusBadRequest)
				return
			}

			logrus.
				WithField("marketID", orderbook.MarketID).
				WithField("sequence", orderbook.Sequence).
				Info("Sending initial state for orderbook")

			data = orderbook

		case "trade":
			trades, err := apiModel.GetTrades(context.Background(), argument, 15)
			if err != nil {
				logrus.WithField(log.AlertTag, log.AlertHigh).Error(err)
				c.Status(http.StatusBadRequest)
				return
			}

			logrus.
				WithField("len", len(trades)).
				Info("Sending initial state for trades")

			data = trades

		case "conditional":
			orders, err := apiModel.GetPlacedOrders(c.Request.Context(), argument, nil)
			if err != nil {
				logrus.WithField(log.AlertTag, log.AlertHigh).Error(err)
				c.Status(http.StatusBadRequest)
				return
			}

			logrus.
				WithField("len", len(orders)).
				Info("Sending initial state for conditional orders")

			data = orders

		default:
			logrus.
				WithField("channel", request.Channel).
				Error("initial state not implemented")

			c.JSON(http.StatusOK, gin.H{
				"result": gin.H{},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"result": gin.H{
				"data": data,
			},
		})
	})

	// TODO: move this to the config
	if err := router.Run("0.0.0.0:7778"); err != nil {
		logrus.Panic(err)
	}
}
