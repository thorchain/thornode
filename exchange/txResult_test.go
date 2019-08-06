package exchange

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type TxHashSuite struct{}

var _ = Suite(&TxHashSuite{})

func (s *TxHashSuite) TestMockEndpoint(c *C) {

	okResponse := `{
"code": 0,
  "hash": "ED92EB231E176EF54CCF6C34E83E44BA971192E75D55C86953BF0FB371F042FA",
  "height": "22466368",
  "log": "Msg 0: ",
  "ok": true,
  "tx": {
	"type": "auth/StdTx",
	"value": {
	  "data": null,
	  "memo": "test",
	  "msg": [
		{
		  "type": "cosmos-sdk/Send",
		  "value": {
			"inputs": [
			  {
				"address": "tbnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7whxk9nt",
				"coins": [
				  {
					"amount": "100000000",
					"denom": "LOK-3C0"
				  }
				]
			  }
			],
			"outputs": [
			  {
				"address": "tbnb13wkwssdkxxj9ypwpgmkaahyvfw5qk823v8kqhl",
				"coins": [
				  {
					"amount": "100000000",
					"denom": "LOK-3C0"
				  }
				]
			  }
			]
		  }
		}
	  ],
	  "signatures": [
		{
		  "account_number": "678061",
		  "pub_key": {
			"type": "tendermint/PubKeySecp256k1",
			"value": "AtjepwjHVfGh9UYnTu7j2uhRhX4ZVuUKb4wBsoSyNOcP"
		  },
		  "sequence": "1",
		  "signature": "xIWxw7DeJ44Q880taoEL1/OzJJqlo/gUiUMMMg2uqOs2MTMcULOk3K0yDbKdkS2B+Sogw0ieQsky62ltX1h/HQ=="
		}
	  ],
	  "source": "1"
	}
  }
	}`

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okResponse))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	cli := NewClient()
	cli.httpClient = httpClient

	result, err := cli.GetTxInfo("ED92EB231E176EF54CCF6C34E83E44BA971192E75D55C86953BF0FB371F042FA")

	c.Assert(err, IsNil)

	c.Check(result.Memo(), Equals, "test")
	c.Check(result.Inputs(), HasLen, 1)
	c.Check(result.Inputs()[0].Address, Equals, "tbnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7whxk9nt")
	c.Assert(result.Inputs()[0].Coins, HasLen, 1)
	c.Check(result.Inputs()[0].Coins[0].Amount.String(), Equals, "100000000")
	c.Check(result.Inputs()[0].Coins[0].Denom, Equals, "LOK-3C0")
	c.Check(result.Outputs(), HasLen, 1)
	c.Check(result.Outputs()[0].Address, Equals, "tbnb13wkwssdkxxj9ypwpgmkaahyvfw5qk823v8kqhl")
	c.Assert(result.Outputs()[0].Coins, HasLen, 1)
	c.Check(result.Outputs()[0].Coins[0].Amount.String(), Equals, "100000000")
	c.Check(result.Outputs()[0].Coins[0].Denom, Equals, "LOK-3C0")

}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewTLSServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return cli, s.Close
}
