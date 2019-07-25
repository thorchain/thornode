package exchange

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/binance-chain/go-sdk/client"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/gorilla/websocket"
	"github.com/jpthor/cosmos-swap/config"
)

type Service struct {
	cfg    config.Settings
	logger zerolog.Logger
	ws     *Wallets
	wg     *sync.WaitGroup
	dc     client.DexClient
	quit   chan struct{}
}

// Binance account struct (WS).
type BinanceAcct struct {
	Stream string `json:"stream"`
	Data   struct {
		Event       string `json:"e"`
		EventHeight int    `json:"E"`
		H           string `json:"H"`
		TimeInForce string `json:"f"`
		T           []struct {
			O string `json:"o"`
			C []struct {
				Asset string `json:"a"`
				A     string `json:"A"`
			} `json:"c"`
		} `json:"t"`
	} `json:"data"`
}

// Bianance transaction struct (API).
type BinanceTxn struct {
	Tx []struct {
		TxHash        string      `json:"txHash"`
		BlockHeight   int         `json:"blockHeight"`
		TxType        string      `json:"txType"`
		TimeStamp     time.Time   `json:"timeStamp"`
		FromAddr      string      `json:"fromAddr"`
		ToAddr        string      `json:"toAddr"`
		Value         string      `json:"value"`
		TxAsset       string      `json:"txAsset"`
		TxFee         string      `json:"txFee"`
		TxAge         int         `json:"txAge"`
		OrderID       interface{} `json:"orderId"`
		Code          int         `json:"code"`
		Data          interface{} `json:"data"`
		ConfirmBlocks int         `json:"confirmBlocks"`
		Memo          string      `json:"memo"`
	} `json:"tx"`
	Total int `json:"total"`
}

// Send a PONG every 30 seconds.
const pongWait = 30 * time.Second

func NewService(cfg config.Settings, ws *Wallets, logger zerolog.Logger) (*Service, error) {
	return &Service{
		cfg:    cfg,
		ws:     ws,
		wg:     &sync.WaitGroup{},
		logger: logger.With().Str("module", "service").Logger(),
	}, nil
}

func (s *Service) Start() error {
	s.logger.Info().Msg("start")
	for _, symbol := range s.cfg.Pools {
		s.logger.Info().Msgf("start to process %s", symbol)

		w, err := s.ws.GetWallet(symbol)
		if nil != err {
			return errors.Wrap(err, "fail to get wallet")
		}

		s.logger.Info().Msgf("wallet:%s", w)
		if err := s.startProcess(w); nil != err {
			return errors.Wrapf(err, "fail to start processing, %s", symbol)
		}
	}

	return nil
}

func (s *Service) startProcess(wallet *Bep2Wallet) error {
	u := url.URL{Scheme: "wss", Host: s.cfg.DexBaseUrl, Path: fmt.Sprintf("/api/ws/%s", string(wallet.PublicAddress))}
	s.logger.Info().Msgf("Listening to: %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		s.logger.Fatal().Msgf("Error: %s", err)
	}

	s.keepAlive(c, pongWait)

	ch := make(chan []byte)
	go s.receiveEvent(ch, wallet.PublicAddress)

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
		}
		ch <- message
	}
}

func (s *Service) receiveEvent(ch chan []byte, poolAddress string) {
	for {
		mm := <-ch
		log.Printf("recv: %s", mm)

		var binance BinanceAcct

		err := json.Unmarshal(mm, &binance)
		if err != nil {
			s.logger.Info().Msgf("There was an error: %s", err)
		}

		if binance.Data.Event == "outboundTransferInfo" {
			s.getTxn(&binance, poolAddress)
		}
	}
}

func (s *Service) getTxn(binance *BinanceAcct, poolAddress string) BinanceTxn {
	select {
	case <-time.After(time.Second * 3):
		u := url.URL{Scheme: "https", Host: s.cfg.DexBaseUrl, Path: "/api/v1/transactions"}

		q := u.Query()
		q.Set("address", poolAddress)
		q.Set("blockHeight", fmt.Sprintf("%v", binance.Data.EventHeight))

		u.RawQuery = q.Encode()

		s.logger.Info().Msgf("Getting transaction from: %s", u.String())

		res, _ := http.Get(u.String())
		body, _ := ioutil.ReadAll(res.Body)

		var txn BinanceTxn

		err := json.Unmarshal(body, &txn)
		if err != nil {
			s.logger.Info().Msgf("There was an error: %s", err)
		}

		s.logger.Info().Msgf("Got Txn: %v", txn)

		return txn
	}
}

func (s *Service) keepAlive(c *websocket.Conn, timeout time.Duration) {
	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()

		return nil
	})

	go func() {
		for {
			err := c.WriteMessage(websocket.PingMessage, []byte("pong"))
			if err != nil {
				return
			}

			time.Sleep(timeout / 2)

			if time.Now().Sub(lastResponse) > timeout {
				c.Close()

				return
			}
		}
	}()
}

func (s *Service) Stop() error {
	os.Exit(1)

	return nil
}
