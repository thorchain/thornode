package common

import (
	"errors"
	"net/http"
	"io/ioutil"
	"github.com/avast/retry-go"
)

func GetWithRetry(uri string) ([]byte, error) {
	var body []byte

	err := retry.Do(
		func() error {
			resp, err := http.Get(uri)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == 404 {
				return errors.New("404")
			}

			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return body, err
}
