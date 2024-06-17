package sources

/*
Responsibilities: handles a web socket server to receive prices for markets. Makes the connection, subscribes to the tickers of interest, pings the server as required to keep the connection alive, reads the price data as it becomes available on the web socket.

Use the gobwas library for the web socket connection.

Used by binance_source.go, okx_source.go, etc. These each implement aspects specific
to their particular exchange, such as the format of the ticker, subscription messages, pings and the data repsonses.
*/
import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"
)

const (
	IO_TIMEOUT             = time.Minute
	PING_INTERVAL          = 25 * time.Second
	MIN_RECONNECT_INTERVAL = 20 * time.Second
	MAX_DIAL_ATTEMPTS      = 10
	INITIAL_CONNECTION     = "initial connection"
	PING_TIMEOUT           = "ping timeout"
	DATA_TIMEOUT           = "data timeout"
	WS_WRITE_ERROR         = "web socket write error"
	WS_READ_ERROR          = "web socket read error"
)

type GobwasSource struct {
	state              State
	latestPrices       *sync.Map
	muState            sync.RWMutex
	muLastPong         sync.Mutex
	lastPongReceived   time.Time
	url                string
	shortUrl           string
	conn               net.Conn
	markets            map[string]string
	coinTickers        map[string]Ticker
	dialer             ws.Dialer
	connectedAt        time.Time
	lastConnectStarted time.Time
	maxUseAge          map[string]time.Duration
	multipliers        map[string]float64
	extractPrice       func([]byte, float64) (PriceTime, error)
	subscriptionBytes  func(markets map[string]Ticker) []byte
	pingAlways         bool
	pingBytes          func() []byte
	pongBytes          func() []byte
	isPriceData        func([]byte, ws.OpCode) (bool, string)
	isPingData         func([]byte, ws.OpCode) bool
	isPongData         func([]byte, ws.OpCode) bool
}

type State string

const (
	disconnected State = "disconnected"
	connecting   State = "connecting"
	connected    State = "connected"
)

func NewGobwasSource() *GobwasSource {
	gs := GobwasSource{
		state:             disconnected,
		dialer:            ws.Dialer{Timeout: 20 * time.Second},
		latestPrices:      &sync.Map{},
		subscriptionBytes: func(markets map[string]Ticker) []byte { return nil },
		pingBytes:         func() []byte { return nil },
		pongBytes:         func() []byte { return nil },
		isPriceData: func([]byte, ws.OpCode) (bool, string) {
			return false, ""
		},
		isPingData: func(bytes []byte, op ws.OpCode) bool {
			return op == ws.OpPing || string(bytes) == "ping"
		},
		isPongData: func(bytes []byte, op ws.OpCode) bool {
			return op == ws.OpPong || string(bytes) == "pong"
		},
	}
	return &gs
}

func (gs *GobwasSource) start(ctx context.Context) {
	gs.reconnect(ctx, INITIAL_CONNECTION)
	go gs.pinger(ctx)
	go gs.wsReader(ctx)
}

func (gs *GobwasSource) listMarkets() (markets map[string]bool) {
	markets = make(map[string]bool, len(gs.coinTickers))
	for marketId := range gs.coinTickers {
		markets[marketId] = true
	}
	return markets
}

// return true if a reconnect was done, false otherwise
func (gs *GobwasSource) reconnect(ctx context.Context, reason string) bool {

	if !gs.enterConnectingState() {
		// already connecting or recently connected so don't try again
		return false
	}

	if reason == INITIAL_CONNECTION {
		logrus.Infof("connecting to %s", gs.url)
	} else {
		logrus.Warnf("reconnecting to %s, reason: %s", gs.url, reason)
	}
	// establish a new connection
	numPrevAttempts := 0
	interval := time.Second
	for {
		if numPrevAttempts > 0 {
			if numPrevAttempts >= MAX_DIAL_ATTEMPTS {
				logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("can't connect to %s, %d failed attempts to dial", gs.url, numPrevAttempts)
				gs.muState.Lock()
				defer gs.muState.Unlock()
				gs.state = disconnected
				return false
			}
			time.Sleep(interval)
			interval = interval * 2
		}
		numPrevAttempts++
		// close any existing connection
		if gs.conn != nil {
			gs.conn.Close()
			gs.conn = nil
			logrus.Infof("disconnected %s", gs.url)
		}
		conn, _, _, err := gs.dialer.Dial(ctx, gs.url)
		if err != nil {
			logrus.Warnf("error dialing %s: %s", gs.url, err.Error())
			continue
		}
		gs.conn = conn

		// send subscription message to server
		subscriptionBytes := gs.subscriptionBytes(gs.coinTickers)
		if len(subscriptionBytes) != 0 {
			err = gs.conn.SetWriteDeadline(time.Now().Add(IO_TIMEOUT))
			if err != nil {
				logrus.Warnf("error setting write deadline for %s: %s", gs.url, err.Error())
				continue
			}
			err = wsutil.WriteClientMessage(gs.conn, ws.OpText, subscriptionBytes)
			if err != nil {
				logrus.Warnf("error subscribing to %s: %s", gs.url, err.Error())
				continue
			}
		}
		logrus.Warnf("connected to %s", gs.url)
		break
	}
	gs.muState.Lock()
	defer gs.muState.Unlock()
	gs.state = connected
	gs.connectedAt = time.Now()
	return true
}

// Enters the connecting state if we are not already in it and
// sufficient time has elapsed since the last connection attempt.
// Returns true if we have entered the connecting state and false
// if we haven't, either because we were already in the connecting
// state or because insufficient time has elapsed since we last
// attempted a connection
func (gs *GobwasSource) enterConnectingState() bool {
	gs.muState.Lock()
	defer gs.muState.Unlock()
	if gs.state == connecting ||
		time.Since(gs.lastConnectStarted) < MIN_RECONNECT_INTERVAL {
		if gs.state == connecting {
			logrus.Warnf("reconnect to %s rejected because already in progress", gs.url)
		} else {
			logrus.Warnf("reconnect to %s rejected because only %v since last attempt", gs.url, time.Since(gs.lastConnectStarted))
		}
		return false
	}
	gs.state = connecting
	gs.lastConnectStarted = time.Now()
	return true
}

func (gs *GobwasSource) pinger(ctx context.Context) {
	ticker := time.NewTicker(PING_INTERVAL)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			gs.handlePingsAndTimeouts(ctx)
		}
	}
}

func (gs *GobwasSource) wsReader(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			gs.readWSData(ctx)
		}
	}
}

func (gs *GobwasSource) handlePingsAndTimeouts(ctx context.Context) {

	// check for data timeouts
	minDataElapsed := time.Hour
	for _, market := range gs.markets {
		elapsed := gs.getDataElapsed(market)
		if elapsed < minDataElapsed {
			minDataElapsed = elapsed
		}
		if elapsed > 3*gs.maxUseAge[market] {
			logrus.Warnf("%s received for %s from %s in %v", NO_PRICE_DATA, market, gs.url, elapsed)
			gs.reconnect(ctx, DATA_TIMEOUT)
			return
		}
	}

	// if pinging, check for ping timeouts
	pingBytes := gs.pingBytes()
	if len(pingBytes) != 0 {
		// find elapsed time since we last received any data or a pong
		elapsed := gs.getPongElapsed()
		if elapsed > minDataElapsed {
			elapsed = minDataElapsed
		}

		// check if we need a ping timeout
		if elapsed > 3*PING_INTERVAL {
			gs.reconnect(ctx, PING_TIMEOUT)
			return
		}

		// check if we need to send a ping
		if gs.pingAlways || elapsed > PING_INTERVAL {
			gs.writeWSData(ctx, ws.OpPing, pingBytes)
		}
	}

}

func (gs *GobwasSource) writeWSData(ctx context.Context, op ws.OpCode, data []byte) {
	err := func() error {
		gs.muState.RLock()
		defer gs.muState.RUnlock()
		if gs.state != connected {
			return nil
		}
		err := gs.conn.SetWriteDeadline(time.Now().Add(IO_TIMEOUT))
		if err != nil {
			return err
		}
		return wsutil.WriteClientMessage(gs.conn, op, data)
	}()
	if err != nil {
		logrus.Warnf("error writing to ws server at %s: %s", gs.url, err.Error())
		gs.reconnect(ctx, WS_WRITE_ERROR)
	}
}

func (gs *GobwasSource) readWSData(ctx context.Context) {
	var byteResponse []byte
	var err error
	var op ws.OpCode
	var state State
	byteResponse, op, err = func() ([]byte, ws.OpCode, error) {
		gs.muState.RLock()
		defer gs.muState.RUnlock()
		state = gs.state
		if state != connected {
			return nil, ws.OpClose, nil
		}
		err := gs.conn.SetReadDeadline(time.Now().Add(IO_TIMEOUT))
		if err != nil {
			return nil, ws.OpClose, err
		}
		return wsutil.ReadServerData(gs.conn)
	}()

	if err != nil {
		err = fmt.Errorf("error reading price from %s: %s", gs.url, err.Error())
	} else if op == ws.OpClose {
		err = fmt.Errorf("web socket connection to %s closed", gs.url)
	} else if isPrice, coinId := gs.isPriceData(byteResponse, op); isPrice {
		gs.handlePriceData(byteResponse, coinId)
	} else if gs.isPongData(byteResponse, op) {
		gs.handlePongData(byteResponse)
	} else if gs.isPingData(byteResponse, op) {
		gs.handlePingData(ctx, byteResponse)
	}

	if err != nil {
		if state != connecting {
			logrus.Warnf("web socket read error: %s", err.Error())
		}
		if state == connecting || !gs.reconnect(ctx, WS_READ_ERROR) {
			// don't keep hammering away with rejecteds reconnects
			time.Sleep(time.Second)
		}
	}
}

func (gs *GobwasSource) handlePriceData(byteResponse []byte, coinId string) {
	market := gs.markets[coinId]
	priceTime, err := gs.extractPrice(byteResponse, gs.multipliers[market])
	if err != nil {
		logrus.Warnf("error extracting price from %s, got `%s`, error is: %s",
			gs.url, byteResponse, err.Error())
		return
	}
	store(market, priceTime, gs.latestPrices, gs.shortUrl)
}

func (gs *GobwasSource) getPongElapsed() time.Duration {
	gs.muState.RLock()
	defer gs.muState.RUnlock()
	gs.muLastPong.Lock()
	defer gs.muLastPong.Unlock()
	if gs.lastPongReceived.IsZero() || gs.lastPongReceived.Before(gs.connectedAt) {
		return time.Since(gs.connectedAt)
	}
	return time.Since(gs.lastPongReceived)
}

func (gs *GobwasSource) getDataElapsed(market string) time.Duration {
	latest := load(market, gs.latestPrices)
	gs.muState.RLock()
	defer gs.muState.RUnlock()
	if latest.time.IsZero() || latest.time.Before(gs.connectedAt) {
		return time.Since(gs.connectedAt)
	}
	return time.Since(latest.time)
}

func (gs *GobwasSource) handlePongData(byteResponse []byte) {
	gs.muLastPong.Lock()
	defer gs.muLastPong.Unlock()
	gs.lastPongReceived = time.Now()
}

func (gs *GobwasSource) handlePingData(ctx context.Context, pingData []byte) {
	if pongBytes := gs.pongBytes(); len(pongBytes) != 0 {
		gs.writeWSData(ctx, ws.OpPong, pongBytes)
	}
}
