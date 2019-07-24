package swapservice

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func swap(ctx sdk.Context, keeper Keeper, source, target, amount, requester, destination string) error {
	isDoubleSwap := source != "ATOM" && target != "ATOM"
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

	fmt.Println("")
	log.Printf("%s Swapping %s(%s) -> %s to %s", requester, source, amount, target, destination)
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
	targetAmount := keeper.GetAccData(ctx, fmt.Sprintf("acc-%s", requester), target)
	keeper.SetAccData(
		ctx,
		fmt.Sprintf("acc-%s", requester),
		requester,
		source,
		fmt.Sprintf("%g", sourceCoins-amt),
	)

	balanceRune, err := strconv.ParseFloat(pool.BalanceRune, 64)
	if err != nil {
		return err
	}
	balanceToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	if err != nil {
		return err
	}

	log.Printf("Pre-Account: %sSource %sTarget", sourceAmount, targetAmount)
	log.Printf("Pre-Pool: %sRune %sToken", pool.BalanceRune, pool.BalanceToken)

	if source == "ATOM" {
		balanceRune += amt
		balanceToken = (amt * balanceToken) / (amt + balanceRune)
		balanceY, err := strconv.ParseFloat(pool.BalanceToken, 64)
		if err != nil {
			return err
		}
		log.Printf("FNew Y: %g", balanceY)
		log.Printf("balanceToken %g", balanceToken)
		balanceY = balanceY - balanceToken
		log.Printf("NNEW Y: %g", balanceY)
		keeper.SetAccData(
			ctx,
			fmt.Sprintf("acc-%s", requester),
			requester,
			target,
			fmt.Sprintf("%g", balanceY),
		)
		log.Printf("Post-Account: %g %s", sourceCoins-amt, fmt.Sprintf("%g", balanceY))
	} else {
		balanceToken += amt
		balanceRune = (balanceRune * amt) / (amt + balanceToken)
		balanceY, err := strconv.ParseFloat(pool.BalanceRune, 64)
		if err != nil {
			return err
		}
		log.Printf("BNew Y: %g", balanceY)
		balanceY = balanceY - balanceRune
		keeper.SetAccData(
			ctx,
			fmt.Sprintf("acc-%s", requester),
			requester,
			target,
			fmt.Sprintf("%g", balanceY),
		)
		log.Printf("Post-Account: %g %s", sourceCoins-amt, fmt.Sprintf("%g", balanceY))
	}

	pool.BalanceRune = fmt.Sprintf("%g", balanceRune)
	pool.BalanceToken = fmt.Sprintf("%g", balanceToken)
	keeper.SetPoolStruct(ctx, poolID, pool)

	log.Printf("Post-Pool: %sAtom %sToken", pool.BalanceRune, pool.BalanceToken)

	return nil
}
