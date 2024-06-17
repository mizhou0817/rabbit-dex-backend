package api_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type WSClientCallback interface {
	AccountInit(data *model.ProfileData)
	AccountData(data *model.ProfileData)
	OrderBookInit(data *model.OrderbookData)
	OrderBookData(data *model.OrderbookData)
	MarketInit(data *model.MarketData)
	MarketData(data *model.MarketData)
	TradeInit(data []*model.TradeData)
	TradeData(data []*model.TradeData)
}

type WSClient struct {
	client     *Client
	connection *websocket.Conn
	marketIDs  []string
	callback   WSClientCallback

	idToChannel map[int]string
}

type (
	wsMessage struct {
		ID int `json:"id"`
	}

	// Auth related structures
	wsAuthMessage struct {
		*wsMessage

		Connect wsAuthConnect `json:"connect"`
	}

	wsAuthConnect struct {
		Token string `json:"token"`
		Name  string `json:"name"`
	}

	// Channel subscription related structures
	wsSubscribeMessage struct {
		*wsMessage

		Subscribe wsChannelSubscribe `json:"subscribe"`
	}

	wsChannelSubscribe struct {
		Channel string
		Name    string
	}

	// Response related structures
	wsResponse struct {
		ID        int          `json:"id"`
		Error     *wsError     `json:"error"`
		Subscribe *wsSubscribe `json:"subscribe"`
		Push      *wsPush      `json:"push"`
	}

	wsError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	wsSubscribe struct {
		Recoverable bool        `json:"recoverable"`
		Epoch       string      `json:"epoch"`
		Positioned  bool        `json:"positioned"`
		Data        interface{} `json:"data"`
	}

	wsPush struct {
		Channel string `json:"channel"`
		Pub     wsPub  `json:"pub"`
	}

	wsPub struct {
		Data interface{} `json:"data"`
	}
)

func NewWSClient(
	url string,
	client *Client,
	marketIDs []string,
	callback WSClientCallback,
) (*WSClient, error) {
	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	idToChannel := make(map[int]string)

	return &WSClient{
		client:      client,
		connection:  connection,
		marketIDs:   marketIDs,
		callback:    callback,
		idToChannel: idToChannel,
	}, nil
}

func (c *WSClient) Start() error {
	defer c.connection.Close()

	// Prepare channels for subscription
	accountChannel := fmt.Sprintf("account@%d", c.client.Credentials.ProfileID)
	channels := make([]string, 0, len(c.marketIDs)+1)
	channels = append(channels, accountChannel)

	if c.marketIDs != nil {
		for _, marketID := range c.marketIDs {
			orderBookChannel := fmt.Sprintf("orderbook:%s", marketID)
			tradeChannel := fmt.Sprintf("trade:%s", marketID)
			marketChannel := fmt.Sprintf("market:%s", marketID)
			channels = append(channels, orderBookChannel, tradeChannel, marketChannel)
		}
	}

	logrus.
		WithField("channels", strings.Join(channels, ", ")).
		Info("subscribing to channels")

	// Now send msg with auth Credentials
	authMsg := wsAuthMessage{
		wsMessage: &wsMessage{ID: 1},
		Connect:   wsAuthConnect{Token: c.client.Credentials.Jwt, Name: "js"},
	}
	if err := c.connection.WriteJSON(authMsg); err != nil {
		return err
	}

	// Subscribe to all channels related to selected markets
	for idx, channel := range channels {
		c.idToChannel[1+idx] = channel
		subMsg := wsSubscribeMessage{
			wsMessage: &wsMessage{ID: 1 + idx},
			Subscribe: wsChannelSubscribe{Channel: channel, Name: "js"},
		}

		if err := c.connection.WriteJSON(subMsg); err != nil {
			return err
		}
	}

	// Now read messages after auth and subscribing to channels
	for {
		var resp wsResponse

		if err := c.connection.ReadJSON(&resp); err != nil {
			logrus.Error(err)
			continue
		}

		// Deal with ping messages, Centrifugo has own format for ping messages
		if resp.ID == 0 && resp.Error == nil && resp.Subscribe == nil && resp.Push == nil {
			c.connection.WriteMessage(websocket.TextMessage, []byte("{}"))
			continue
		}

		if err := c.processMessage(&resp); err != nil {
			logrus.Error(err)
		}
	}
}

func (c *WSClient) processMessage(resp *wsResponse) error {
	var (
		channel string
		data    []byte
		err     error
		initial bool
	)

	if sub := resp.Subscribe; sub != nil {
		// Let's find out what in what channel we received initial data
		var ok bool
		channel, ok = c.idToChannel[resp.ID]
		if !ok {
			errText := fmt.Sprintf("channel with id %d not found in %v", resp.ID, c.idToChannel)

			return errors.New(errText)
		}

		data, err = json.Marshal(sub.Data)
		if err != nil {
			return err
		}

		initial = true
	}

	if push := resp.Push; push != nil {
		channel = push.Channel
		data, err = json.Marshal(push.Pub.Data)
		if err != nil {
			return err
		}
	}

	channelName := strings.Split(channel, ":")[0]
	channelName = strings.Split(channelName, "@")[0] //  changed from "#" to "@"

	switch channelName {
	case "account":
		var profile model.ProfileData

		if err = json.Unmarshal(data, &profile); err != nil {
			return err
		}

		if initial {
			c.callback.AccountInit(&profile)
		} else {
			c.callback.AccountData(&profile)
		}

	case "orderbook":
		var orderbook model.OrderbookData

		if err = json.Unmarshal(data, &orderbook); err != nil {
			return err
		}

		if initial {
			c.callback.OrderBookInit(&orderbook)
		} else {
			c.callback.OrderBookData(&orderbook)
		}

	case "trade":
		trades := make([]*model.TradeData, 0)

		if err = json.Unmarshal(data, &trades); err != nil {
			return err
		}

		if initial {
			c.callback.TradeInit(trades)
		} else {
			c.callback.TradeData(trades)
		}

	case "market":
		var market model.MarketData

		if err = json.Unmarshal(data, &market); err != nil {
			return err
		}

		if initial {
			c.callback.MarketInit(&market)
		} else {
			c.callback.MarketData(&market)
		}
	}

	return nil
}
