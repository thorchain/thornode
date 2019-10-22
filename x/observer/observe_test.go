package observer

import (
	"crypto/tls"
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
		case "/api/v1/validators":
			if _, err := w.Write([]byte(`{"block_height":123456789,"validators":[]}`)); nil != err {
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
	height := binanceHeight(s.Listener.Addr().String(), *client)
	if height != expectedHeight {
		t.Errorf("Got a height of %v but expected %v!", height, expectedHeight)
	}
}
