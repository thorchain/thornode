package swapservice

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func stake(ctx sdk.Context, keeper Keeper, requester, ticker, atom_amount, token_amount string) error {

	fmt.Println("")
	log.Printf("%s staking %s %s %s", requester, ticker, atom_amount, token_amount)

	ticker = strings.ToUpper(ticker)

	stakeID := fmt.Sprintf("stake-%s", ticker)
	poolID := fmt.Sprintf("pool-%s", ticker)
	pool := keeper.GetPoolStruct(ctx, poolID)

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
		return fmt.Errorf("Insufficient funds: No tokens in account")
	}
	tickerCoins, err := strconv.ParseFloat(tickerAmount, 64)
	if err != nil {
		return err
	}
	if token_amt > tickerCoins {
		return fmt.Errorf("Insufficient funds: Not enough tokens in account (%g/%g)", token_amt, tickerCoins)
	}

	atomAmount := keeper.GetAccData(ctx, fmt.Sprintf("acc-%s", requester), "ATOM")
	if atomAmount == "" {
		return fmt.Errorf("Insufficient funds: No atoms in account")
	}
	atomCoins, err := strconv.ParseFloat(atomAmount, 64)
	if err != nil {
		return err
	}
	if atom_amt > atomCoins {
		return fmt.Errorf("Insufficient funds: Not enough atoms in account (%g/%g)", atom_amt, atomCoins)
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

	log.Printf("Pre-Account: %sAtom %sToken", atomAmount, tickerAmount)
	log.Printf("Pre-Stake: %sAtom %sToken", stake.Atom, stake.Token)
	log.Printf("Pre-Pool: %sAtom %sToken", pool.BalanceAtom, pool.BalanceToken)
	log.Printf("Staking: %sAtom %sToken", atom_amount, token_amount)
	log.Printf("Post-Account: %sAtom %sToken", fmt.Sprintf("%g", atomCoins-atom_amt), fmt.Sprintf("%g", tickerCoins-token_amt))
	log.Printf("Post-Staking: %sAtom %sToken", fmt.Sprintf("%g", atom_amt-stakeAtom), fmt.Sprintf("%g", token_amt-stakeToken))

	keeper.SetAccData(
		ctx,
		fmt.Sprintf("acc-%s", requester),
		requester,
		ticker,
		fmt.Sprintf("%g", tickerCoins-token_amt),
	)
	keeper.SetAccData(
		ctx,
		fmt.Sprintf("acc-%s", requester),
		requester,
		"ATOM",
		fmt.Sprintf("%g", atomCoins-atom_amt),
	)
	keeper.SetStakeData(
		ctx,
		stakeID,
		requester,
		fmt.Sprintf("%g", atom_amt-stakeAtom),
		fmt.Sprintf("%g", token_amt-stakeToken),
	)

	balanceAtom, err := strconv.ParseFloat(pool.BalanceAtom, 64)
	if err != nil {
		return err
	}
	balanceToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	if err != nil {
		return err
	}
	pool.BalanceAtom = fmt.Sprintf("%g", balanceAtom+atom_amt)
	pool.BalanceToken = fmt.Sprintf("%g", balanceToken+token_amt)
	pool.Ticker = ticker
	log.Printf("Post-Pool: %sAtom %sToken", pool.BalanceAtom, pool.BalanceToken)

	keeper.SetPoolStruct(ctx, poolID, pool)

	// TODO: delete pool/stake keys if no stakes are left

	return nil
}
