package thorclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

var EmptyNodeAccount types.NodeAccount

// GetNodeAccount from thorchain
func GetNodeAccount(c *http.Client, chainHost, thorAddr string) (types.NodeAccount, error) {
	uri := url.URL{
		Scheme: "http",
		Host:   chainHost,
		Path:   "/thorchain/nodeaccount/" + thorAddr,
	}
	resp, err := c.Get(uri.String())
	if nil != err {
		return EmptyNodeAccount, fmt.Errorf("fail to get node account from thorchain,err:%w", err)
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return EmptyNodeAccount, fmt.Errorf("fail to get node account from thorchain,statusCode:%d", resp.StatusCode)
	}
	var na types.NodeAccount

	buf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return EmptyNodeAccount, fmt.Errorf("fail to read response body,err:%w", err)
	}
	cdc := MakeCodec()
	if err := cdc.UnmarshalJSON(buf, &na); nil != err {
		return EmptyNodeAccount, fmt.Errorf("fail to unmarshal node account response,err:%w", err)
	}
	return na, nil
}
