package slipstopper

import (
	"encoding/json"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/golang-jwt/jwt/v4"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

type WSClient struct {
	matcherByMarket map[string]*Matcher
	cfg             *Config
}

type OrderEvent struct {
	Orders []model.OrderData `json:"orders"`
}

type PriceEvent struct {
	FairPrice  *decimal.Decimal `json:"fair_price"`
	IndexPrice *decimal.Decimal `json:"index_price"`
}

func (ws *WSClient) connToken(secretToken string, user string) (string, error) {
	claims := jwt.MapClaims{"sub": user}
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretToken))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (ws *WSClient) subscribe(client *centrifuge.Client, market string) {
	// each market has its own Matcher
	matcher := NewMatcher()
	ws.matcherByMarket[market] = matcher

	// we throttle the calls for price update to 10 milliseconds
	throttle := rate.Sometimes{Interval: time.Millisecond * 200}

	privateChannel := "conditional:" + market
	marketChannel := "market:" + market

	privateSub, err := client.NewSubscription(privateChannel, centrifuge.SubscriptionConfig{
		Recoverable: true,
	})
	if err != nil {
		logrus.Fatalln(err)
	}

	marketSub, err := client.NewSubscription(marketChannel, centrifuge.SubscriptionConfig{
		Recoverable: true,
	})
	if err != nil {
		logrus.Fatalln(err)
	}

	privateSub.OnSubscribed(func(e centrifuge.SubscribedEvent) {
		var orders []model.OrderData
		err := json.Unmarshal(e.Data, &orders)
		if err != nil {
			logrus.Warn("could not parse initial data ", err)
		} else {
			matcher.ClearAll()

			matcher.mu.Lock()
			defer matcher.mu.Unlock()
			for _, order := range orders {
				matcher.Insert(order)
			}
		}

		logrus.Infof("Subscribed to channel=%s got orders=%d", privateSub.Channel, len(orders))
	})

	marketSub.OnSubscribed(func(e centrifuge.SubscribedEvent) {
		logrus.Info("Subscribed to channel: ", marketSub.Channel)
	})

	privateSub.OnPublication(func(e centrifuge.PublicationEvent) {
		var data OrderEvent
		logrus.Info("order updated: ", privateSub.Channel, ": ", string(e.Data))
		err := json.Unmarshal(e.Data, &data)
		if err != nil {
			logrus.Warn(err)
			return
		}

		matcher.mu.Lock()
		defer matcher.mu.Unlock()

		for _, order := range data.Orders {
			if order.Status == model.PLACED {
				// if an order exists already then this is an update. (such as price etc...)
				_, ok := matcher.nodesByOrderId[order.OrderId]
				if ok {
					matcher.Remove(order)
				}
				// insert new order
				matcher.Insert(order)
			} else {
				// any other status means the execution of the order has been sent to the engine already.
				matcher.Remove(order)
			}
		}
	})

	marketSub.OnPublication(func(e centrifuge.PublicationEvent) {
		throttle.Do(func() {
			var data PriceEvent
			err = json.Unmarshal(e.Data, &data)
			if err != nil {
				logrus.Warn(err)
				return
			}
			if data.FairPrice != nil {
				logrus.Info("Received price update: ", marketSub.Channel, " ", data.FairPrice)
				matcher.mu.Lock()
				defer matcher.mu.Unlock()
				matcher.OnPriceUpdate(*data.FairPrice)
			}
		})
	})

	err = privateSub.Subscribe()
	if err != nil {
		logrus.Fatalln(err)
	}

	err = marketSub.Subscribe()
	if err != nil {
		logrus.Fatalln(err)
	}
}

func (ws *WSClient) Run(readyChan chan bool) {
	logrus.Info("Starting slipstopper...")
	logrus.Info("Connection info: ", ws.cfg.Service.WebsocketURI)
	logrus.Info("Markets: ", ws.cfg.Service.Markets)

	token, err := ws.connToken(ws.cfg.Service.CentrifugoHMACSecretToken, "slipstopper")
	if err != nil {
		logrus.Fatalln(err)
	}

	client := centrifuge.NewJsonClient(
		ws.cfg.Service.WebsocketURI,
		centrifuge.Config{
			Token: token,
		},
	)
	defer client.Close()

	client.OnConnecting(func(e centrifuge.ConnectingEvent) {
		logrus.Printf("Connecting - %d (%s)", e.Code, e.Reason)
	})

	client.OnConnected(func(e centrifuge.ConnectedEvent) {
		logrus.Printf("Connected with ID %s", e.ClientID)
	})

	client.OnDisconnected(func(e centrifuge.DisconnectedEvent) {
		logrus.WithField(log.AlertTag, log.AlertHigh).Warnf("Disconnected: %d (%s)", e.Code, e.Reason)
	})

	client.OnError(func(e centrifuge.ErrorEvent) {
		logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Error: %v", e.Error)
	})

	err = client.Connect()
	if err != nil {
		logrus.Fatalln(err)
	}

	for _, market := range ws.cfg.Service.Markets {
		ws.subscribe(client, market)
	}

	readyChan <- true
	// Run until CTRL+C.
	select {}
}

func NewWSClient(cfg *Config) *WSClient {
	return &WSClient{
		matcherByMarket: make(map[string]*Matcher),
		cfg:             cfg,
	}
}
