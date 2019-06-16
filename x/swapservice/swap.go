package swapservice

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func swap(ctx sdk.Context, keeper Keeper, source, target, amount, requester, destination string) error {
	isDoubleSwap := source != "ATOM" || target != "ATOM"
	source = strings.ToUpper(source)
	target = strings.ToUpper(target)

	if isDoubleSwap {
		err := swapOne(ctx, keeper, source, "ATOM", amount, requester, requester)
		if err != nil {
			return err
		}
		return swapOne(ctx, keeper, "ATOM", target, amount, requester, destination)
	} else {
		return swapOne(ctx, keeper, source, target, amount, requester, destination)
	}

}

func swapOne(ctx sdk.Context, keeper Keeper, source, target, amount, requester, destination string) error {

	poolID := fmt.Sprintf("pool-%s", source)
	if source == "ATOM" {
		poolID = fmt.Sprintf("pool-%s", target)
	}

	pool := keeper.GetPoolStruct(ctx, poolID)
	if pool.Ticker == "" {
		return fmt.Errorf("No pool available (%s)", poolID)
	}

	amt, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return err
	}
	sourceAmount := keeper.GetAccData(ctx, fmt.Sprintf("acc-%s", requester), source)
	if sourceAmount == "" {
		return fmt.Errorf("Insufficient funds")
	}
	sourceCoins, err := strconv.ParseFloat(sourceAmount, 64)
	if err != nil {
		return err
	}
	if amt > sourceCoins {
		return fmt.Errorf("Insufficient funds")
	}
	keeper.SetAccData(
		ctx,
		fmt.Sprintf("acc-%s", requester),
		requester,
		source,
		fmt.Sprintf("%g", amt-sourceCoins),
	)

	balanceAtom, err := strconv.ParseFloat(pool.BalanceAtom, 64)
	if err != nil {
		return err
	}
	balanceToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	if err != nil {
		return err
	}

	var balanceY float64
	if source == "ATOM" {
		balanceAtom += amt
		balanceToken = (amt * balanceToken) / (amt + balanceAtom)
		balanceY, err := strconv.ParseFloat(pool.BalanceToken, 64)
		if err != nil {
			return err
		}
		balanceY = balanceY - balanceToken
	} else {
		balanceToken += amt
		balanceAtom = (balanceAtom * amt) / (amt + balanceToken)
		balanceY, err := strconv.ParseFloat(pool.BalanceAtom, 64)
		if err != nil {
			return err
		}
		balanceY = balanceY - balanceAtom
	}

	pool.BalanceAtom = fmt.Sprintf("%g", balanceAtom)
	pool.BalanceToken = fmt.Sprintf("%g", balanceToken)
	keeper.SetPoolStruct(ctx, poolID, pool)

	keeper.SetAccData(
		ctx,
		fmt.Sprintf("acc-%s", requester),
		requester,
		target,
		fmt.Sprintf("%g", balanceY),
	)

	return nil
}
