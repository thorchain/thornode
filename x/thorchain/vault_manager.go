package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// const values used to emit events
const (
	EventTypeActiveVault   = "ActiveVault"
	EventTypeInactiveVault = "InactiveVault"
)

// VaultManager
type VaultManager interface {
	TriggerKeygen(ctx sdk.Context, nas NodeAccounts) error
	RotateVault(ctx sdk.Context, vault Vault) error
	EndBlock(ctx sdk.Context, constAccessor constants.ConstantValues) error
}

// VaultMgr is going to manage the vaults
type VaultMgr struct {
	k          Keeper
	txOutStore TxOutStore
}

// NewVaultMgr create a new vault manager
func NewVaultMgr(k Keeper, txOutStore TxOutStore) *VaultMgr {
	return &VaultMgr{
		k:          k,
		txOutStore: txOutStore,
	}
}

// EndBlock: move funds from retiring asgard vaults
func (vm *VaultMgr) EndBlock(ctx sdk.Context, constAccessor constants.ConstantValues) error {
	migrateInterval := constAccessor.GetInt64Value(constants.FundMigrationInterval)

	retiring, err := vm.k.GetAsgardVaultsByStatus(ctx, RetiringVault)
	if err != nil {
		return err
	}

	active, err := vm.k.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		return err
	}

	// if we have no active asgards to move funds to, don't move funds
	if len(active) == 0 {
		return nil
	}

	for _, vault := range retiring {
		if !vault.HasFunds() {
			// no more funds to move, delete the vault
			if err := vm.k.DeleteVault(ctx, vault.PubKey); err != nil {
				return err
			}
			continue
		}

		// move partial funds every 30 minutes
		if (ctx.BlockHeight()-vault.StatusSince)%migrateInterval == 0 {
			for _, coin := range vault.Coins {

				// determine which active asgard vault is the best to send
				// these coins to. We target the vault with the least amount of
				// this particular coin
				cn := active[0].GetCoin(coin.Asset)
				pk := active[0].PubKey
				for _, asgard := range active {
					if cn.Amount.GT(asgard.GetCoin(coin.Asset).Amount) {
						cn = asgard.GetCoin(coin.Asset)
						pk = asgard.PubKey
					}
				}

				// get address of asgard pubkey
				addr, err := pk.GetAddress(coin.Asset.Chain)
				if err != nil {
					return err
				}

				// figure the nth time, we've sent migration txs from this vault
				nth := (ctx.BlockHeight()-vault.StatusSince)/migrateInterval + 1

				// Default amount set to total remaining amount. Relies on the
				// signer, to successfully send these funds while respecting
				// gas requirements (so it'll actually send slightly less)
				amt := coin.Amount
				if nth <= 5 { // migrate partial funds 5 times
					// migrating 20% of our funds. This will make this 20%
					// unavailable to the system for the time it takes the txs
					// to complete. For BNB, its instant, but for Bitcoin it
					// can be 10 minutes.
					amt = amt.QuoUint64(4)
				}

				toi := &TxOutItem{
					Chain:       coin.Asset.Chain,
					InHash:      common.BlankTxID,
					ToAddress:   addr,
					VaultPubKey: vault.PubKey,
					Coin: common.Coin{
						Asset:  coin.Asset,
						Amount: amt,
					},
					Memo: "migrate",
				}
				vm.txOutStore.AddTxOutItem(ctx, toi)
			}
		}
	}

	return nil
}

func (vm *VaultMgr) TriggerKeygen(ctx sdk.Context, nas NodeAccounts) error {
	keygen := make(common.PubKeys, len(nas))
	for i := range nas {
		keygen[i] = nas[i].PubKeySet.Secp256k1
	}
	keygens := NewKeygens(uint64(ctx.BlockHeight()))
	keygens.Keygens = []common.PubKeys{keygen}
	return vm.k.SetKeygens(ctx, keygens)
}

func (vm *VaultMgr) RotateVault(ctx sdk.Context, vault Vault) error {
	active, err := vm.k.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		return err
	}

	// find vaults the new vault conflicts with, mark them as inactive
	for _, asgard := range active {
		for _, member := range asgard.Membership {
			if vault.Contains(member) {
				asgard.UpdateStatus(RetiringVault, ctx.BlockHeight())
				if err := vm.k.SetVault(ctx, asgard); err != nil {
					return err
				}

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(EventTypeInactiveVault,
						sdk.NewAttribute("set asgard vault to inactive", asgard.PubKey.String())))
				break
			}
		}
	}

	// Update Node account membership
	for _, member := range vault.Membership {
		na, err := vm.k.GetNodeAccountByPubKey(ctx, member)
		if err != nil {
			return err
		}
		na.TryAddSignerPubKey(vault.PubKey)
		if err := vm.k.SetNodeAccount(ctx, na); err != nil {
			return err
		}
	}

	if err := vm.k.SetVault(ctx, vault); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeActiveVault,
			sdk.NewAttribute("add new asgard vault", vault.PubKey.String())))
	return nil
}
