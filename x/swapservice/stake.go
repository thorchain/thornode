package swapservice

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func stake(ctx sdk.Context, keeper Keeper, requester, ticker, atom_amount, token_amount string) error {

	ticker = strings.ToUpper(ticker)

	stakeID := fmt.Sprintf("stake-%s", ticker)
	poolID := fmt.Sprintf("pool-%s", ticker)
	pool := keeper.GetPoolStruct(ctx, poolID)
	if pool.Ticker == "" {
		return fmt.Errorf("No pool available (%s)", poolID)
	}

	token_amt, err := strconv.ParseFloat(token_amount, 64)
	if err != nil {
		return err
	}
	atom_amt, err := strconv.ParseFloat(atom_amount, 64)
	if err != nil {
		return err
	}
	tickerAmount := keeper.GetAccData(ctx, fmt.Sprintf("acc-%s", requester), ticker)
	if tickerAmount == "" {
		return fmt.Errorf("Insufficient funds")
	}
	tickerCoins, err := strconv.ParseFloat(tickerAmount, 64)
	if err != nil {
		return err
	}
	if token_amt > tickerCoins {
		return fmt.Errorf("Insufficient funds")
	}

	atomAmount := keeper.GetAccData(ctx, fmt.Sprintf("acc-%s", requester), "ATOM")
	if atomAmount == "" {
		return fmt.Errorf("Insufficient funds")
	}
	atomCoins, err := strconv.ParseFloat(atomAmount, 64)
	if err != nil {
		return err
	}
	if atom_amt > atomCoins {
		return fmt.Errorf("Insufficient funds")
	}

	stake := keeper.GetStakeData(ctx, stakeID, requester)
	stakeAtom, err := strconv.ParseFloat(stake.Atom, 64)
	if err != nil {
		return err
	}
	stakeToken, err := strconv.ParseFloat(stake.Token, 64)
	if err != nil {
		return err
	}

	keeper.SetAccData(
		ctx,
		fmt.Sprintf("acc-%s", requester),
		requester,
		ticker,
		fmt.Sprintf("%g", token_amt-tickerCoins),
	)
	keeper.SetAccData(
		ctx,
		fmt.Sprintf("acc-%s", requester),
		requester,
		"ATOM",
		fmt.Sprintf("%g", atom_amt-atomCoins),
	)
	keeper.SetStakeData(
		ctx,
		stakeID,
		requester,
		fmt.Sprintf("%g", stakeAtom-atom_amt),
		fmt.Sprintf("%g", stakeToken-token_amt),
	)

	balanceAtom, err := strconv.ParseFloat(pool.BalanceAtom, 64)
	if err != nil {
		return err
	}
	balanceToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	if err != nil {
		return err
	}
	pool.BalanceAtom = fmt.Sprintf("%g", balanceAtom+atomCoins)
	pool.BalanceToken = fmt.Sprintf("%g", balanceToken+tickerCoins)

	keeper.SetPoolStruct(ctx, poolID, pool)

	return nil
}
