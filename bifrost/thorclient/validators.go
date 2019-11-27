package thorclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

func GetValidators(c *http.Client, chainHost string) (*types.ValidatorsResp, error) {
	uri := url.URL{
		Scheme: "http",
		Host:   chainHost,
		Path:   "/thorchain/validators",
	}
	resp, err := c.Get(uri.String())
	if nil != err {
		return nil, fmt.Errorf("fail to get validators from thorchain,err:%w", err)
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fail to get validators from thorchain,statusCode:%d", resp.StatusCode)
	}
	var vr types.ValidatorsResp

	buf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return nil, fmt.Errorf("fail to read response body,err:%w", err)
	}
	cdc := MakeCodec()
	if err := cdc.UnmarshalJSON(buf, &vr); nil != err {
		return nil, fmt.Errorf("fail to unmarshal validator response,err:%w", err)
	}
	return &vr, nil
}
