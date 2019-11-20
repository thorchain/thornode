package observer

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	expectedHeight = int64(123456789)
)

// TestBinanceHeight : Test to ensure we get back the height we expect.
func TestBinanceHeight(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/abci_info":
			if _, err := w.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "response": { "data": "BNBChain", "last_block_height": "123456789", "last_block_app_hash": "pwx4TJjXu3yaF6dNfLQ9F4nwAhjIqmzE8fNa+RXwAzQ=" } } }`)); nil != err {
				t.Error(err)
			}
		}
	})

	s := httptest.NewTLSServer(h)
	defer s.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	height := binanceHeight(fmt.Sprintf("https://%s", s.Listener.Addr().String()), *client)
	if height != expectedHeight {
		t.Errorf("Got a height of %v but expected %v!", height, expectedHeight)
	}
}
