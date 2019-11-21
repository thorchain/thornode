package observer

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"

	. "gopkg.in/check.v1"
)

const (
	expectedHeight = int64(123456789)
)

type ObserverSuite struct{}

var _ = Suite(&ObserverSuite{})

// TestBinanceHeight : Test to ensure we get back the height we expect.
func (s *ObserverSuite) TestBinanceHeight(c *C) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/abci_info":
			_, err := w.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "response": { "data": "BNBChain", "last_block_height": "123456789", "last_block_app_hash": "pwx4TJjXu3yaF6dNfLQ9F4nwAhjIqmzE8fNa+RXwAzQ=" } } }`))
			c.Assert(err, IsNil)
		}
	})

	ser := httptest.NewServer(h)
	defer ser.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	height, err := binanceHeight(ser.URL, *client)
	c.Assert(err, IsNil)
	c.Check(height, Equals, expectedHeight)
}
