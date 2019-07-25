package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func stake(ctx sdk.Context, keeper Keeper, requester, ticker, rune_amount, token_amount string) error {

	//log.Printf("%s staking %s %s %s", requester, ticker, rune_amount, token_amount)
	//
	//ticker = strings.ToUpper(ticker)
	//
	//stakeID := fmt.Sprintf("stake-%s", ticker)
	//poolID := fmt.Sprintf("pool-%s", ticker)
	//pool := keeper.GetPoolStruct(ctx, poolID)
	//
	//token_amt, err := strconv.ParseFloat(token_amount, 64)
	//if err != nil {
	//	return err
	//}
	//rune_amt, err := strconv.ParseFloat(rune_amount, 64)
	//if err != nil {
	//	return err
	//}
	//
	//tickerCoins, err := strconv.ParseFloat(tickerAmount, 64)
	//if err != nil {
	//	return err
	//}
	//if token_amt > tickerCoins {
	//	return fmt.Errorf("Insufficient funds: Not enough tokens in account (%g/%g)", token_amt, tickerCoins)
	//}
	//
	//atomAmount := keeper.GetAccData(ctx, fmt.Sprintf("acc-%s", requester), "ATOM")
	//if atomAmount == "" {
	//	return fmt.Errorf("Insufficient funds: No runes in account")
	//}
	//runeCoins, err := strconv.ParseFloat(atomAmount, 64)
	//if err != nil {
	//	return err
	//}
	//if rune_amt > runeCoins {
	//	return fmt.Errorf("Insufficient funds: Not enough runes in account (%g/%g)", rune_amt, runeCoins)
	//}
	//
	//stake := keeper.GetStakeData(ctx, stakeID, requester)
	//stakeRune, err := strconv.ParseFloat(stake.Rune, 64)
	//if err != nil {
	//	return err
	//}
	//stakeToken, err := strconv.ParseFloat(stake.Token, 64)
	//if err != nil {
	//	return err
	//}
	//
	//if stakeRune+rune_amt < 0 {
	//	return fmt.Errorf("Insufficient funds: Not enough ATOM coins staked")
	//}
	//if stakeToken+token_amt < 0 {
	//	return fmt.Errorf("Insufficient funds: Not enough token coins staked")
	//}
	//
	//log.Printf("Pre-Account: %sAtom %sToken", atomAmount, tickerAmount)
	//log.Printf("Pre-Stake: %sAtom %sToken", stake.Rune, stake.Token)
	//log.Printf("Pre-Pool: %sAtom %sToken", pool.BalanceRune, pool.BalanceToken)
	//log.Printf("Staking: %sAtom %sToken", rune_amount, token_amount)
	//log.Printf("Post-Account: %sAtom %sToken", fmt.Sprintf("%g", runeCoins-rune_amt), fmt.Sprintf("%g", tickerCoins-token_amt))
	//log.Printf("Post-Staking: %sAtom %sToken", fmt.Sprintf("%g", stakeRune+rune_amt), fmt.Sprintf("%g", stakeToken+token_amt))
	//
	//keeper.SetAccData(
	//	ctx,
	//	fmt.Sprintf("acc-%s", requester),
	//	requester,
	//	ticker,
	//	fmt.Sprintf("%g", tickerCoins-token_amt),
	//)
	//keeper.SetAccData(
	//	ctx,
	//	fmt.Sprintf("acc-%s", requester),
	//	requester,
	//	"ATOM",
	//	fmt.Sprintf("%g", runeCoins-rune_amt),
	//)
	//keeper.SetStakeData(
	//	ctx,
	//	stakeID,
	//	requester,
	//	fmt.Sprintf("%g", stakeRune+rune_amt),
	//	fmt.Sprintf("%g", stakeToken+token_amt),
	//)
	//
	//balanceRune, err := strconv.ParseFloat(pool.BalanceRune, 64)
	//if err != nil {
	//	return err
	//}
	//balanceToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	//if err != nil {
	//	return err
	//}
	//pool.BalanceRune = fmt.Sprintf("%g", balanceRune+rune_amt)
	//pool.BalanceToken = fmt.Sprintf("%g", balanceToken+token_amt)
	//pool.Ticker = ticker
	//log.Printf("Post-Pool: %sAtom %sToken", pool.BalanceRune, pool.BalanceToken)
	//
	//keeper.SetPoolStruct(ctx, poolID, pool)

	// TODO: delete pool/stake keys if no stakes are left

	return nil
}
