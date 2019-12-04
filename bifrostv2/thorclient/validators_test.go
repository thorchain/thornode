package thorclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/x/thorchain"
)

func TestValidators(t *testing.T) { TestingT(t) }

type ValidatorsTestSuite struct{}

var _ = Suite(&ValidatorsTestSuite{})

func (*ValidatorsTestSuite) SetUpSuite(c *C) {
	thorchain.SetupConfigForTest()
}

const normalResp = `{
  "active_nodes": [
    {
      "node_address": "thor146z3xkdlyzmda639ljsk3qvucpem0f60d2lz62",
      "status": "active",
      "node_pub_keys": {
        "secp256k1": "thorpub1addwnpepq0t2qpwk0rx4da68zzvl6w7vdcdygyzau49ffc2kqnx0624ard576060nyk",
        "ed25519": "thorpub1addwnpepq0t2qpwk0rx4da68zzvl6w7vdcdygyzau49ffc2kqnx0624ard576060nyk"
      },
      "validator_cons_pub_key": "thorcpub1zcjduepqx34rzj073kcqzwefz3j70rfcps2e3he7yx4qu45pv09nektnk67swkgalp",
      "bond": "0",
      "bond_address": "tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5",
      "status_since": "1",
      "observer_active": false,
      "signer_active": false,
      "signer_membership": [
        "thorpub1addwnpepqw9gv9ffmua8xpgyqj5nn4slw62wpwvcfgj6xxwx9tuyg0x80xlf24a8s6c"
      ],
      "version": "1"
    }
  ],
  "nominated": null,
  "queued": null,
  "rotate_at": "17281",
  "rotate_window_open_at": "16081"
}`

func (*ValidatorsTestSuite) TestGetValidators(c *C) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Assert(req.RequestURI, Equals, "/thorchain/validators")
		_, err := rw.Write([]byte(normalResp))
		c.Assert(err, IsNil)
	}))
	defer server.Close()
	httpClient := &http.Client{
		Timeout: time.Second,
	}
	resp, err := GetValidators(httpClient, server.Listener.Addr().String())
	c.Assert(err, IsNil)
	c.Assert(resp, NotNil)
	c.Assert(resp.Nominated, IsNil)
	c.Assert(resp.Queued, IsNil)
	c.Assert(resp.RotateWindowOpenAt, Equals, uint64(16081))
	c.Assert(resp.RotateAt, Equals, uint64(17281))
	c.Assert(resp.ActiveNodes, HasLen, 1)
}
