package statechain

// import (
// 	"fmt"
// 	"net/url"
// 	"encoding/json"

// 	log "github.com/rs/zerolog/log"

// 	"gitlab.com/thorchain/bepswap/observe/common"
// 	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
// 	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
// )

// func TxnBlockHeight(txn string) string {
// 	uri := url.URL{
// 		Scheme: "http",
// 		Host: ctypes.ChainHost,
// 		Path: fmt.Sprintf("/txs/%s", txn),
// 	}

// 	body, _ := common.GetWithRetry(uri.String())

// 	var txs types.Txs
// 	err := json.Unmarshal(body, &txs)
// 	if err != nil {
// 		log.Error().Msgf("Error: %v", err)
// 	}

// 	return txs.Height
// }
