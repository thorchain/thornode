package statechain

// import (
// 	"fmt"
// 	"net/url"
// 	"strconv"
// 	"encoding/json"

// 	log "github.com/rs/zerolog/log"

// 	"gitlab.com/thorchain/bepswap/observe/common"
// 	types "gitlab.com/thorchain/bepswap/observe/common/types"
// )

// func TxOut(blockHeight string) types.OutTx {
// 	uri := url.URL{
// 		Scheme: "http",
// 		Host: types.ChainHost,
// 		Path: fmt.Sprintf("/swapservice/txoutarray/%s", blockHeight),
// 	}

// 	body, _ := common.GetWithRetry(uri.String())

// 	var outTx types.OutTx
// 	err := json.Unmarshal(body, &outTx)
// 	if err != nil {
// 		log.Error().Msgf("Error: %v", err)
// 	}

// 	for i, txArr := range outTx.TxArray {
// 		for j, coin := range txArr.Coins {
// 			parsedAmt, _ := strconv.ParseFloat(coin.Amount, 64)
// 			amount := parsedAmt
// 			outTx.TxArray[i].Coins[j].Amount = fmt.Sprintf("%.0f", amount)
// 		}
// 	}

// 	return outTx
// }
