package addressmanager

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain"
)

func Test(t *testing.T) {
	TestingT(t)
}

type AddressMangerSuite struct {
	server *httptest.Server
}

var _ = Suite(&AddressMangerSuite{})

func (s *AddressMangerSuite) SetUpSuite(c *C) {
	thorchain.SetupConfigForTest()
	fmt.Println("SetUpSuite!!")
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.RequestURI {
		case "/thorchain/pooladdresses":
			poolsAddresses(c, rw)
		}
	}))
}

func poolsAddresses(c *C, rw http.ResponseWriter) {
	content, err := ioutil.ReadFile("../../test/fixtures/endpoints/poolAddresses/pooladdresses.json")
	if err != nil {
		c.Fatal(err)
	}
	rw.Header().Set("Content-Type", "application/json")
	if _, err := rw.Write(content); err != nil {
		c.Fatal(err)
	}
}

func (s *AddressMangerSuite) TestGetPoolAddresses(c *C) {
	addrMr, err := NewAddressManager(s.server.URL, &metrics.Metrics{})
	if err != nil {
		c.Error(err.Error())
		return
	}

	pa, err := addrMr.getPoolAddresses()
	if err != nil {
		c.Error(err.Error())
	}
	c.Assert(pa, NotNil)
	c.Assert(pa.Current.GetByChain(common.BNBChain).PubKey.String(), Equals, "thorpub1addwnpepq0c8wahkfpc3s65rl6ut262jwd57tp2qtp4dfvdtqllcmccdepp8usg7d47")
}
