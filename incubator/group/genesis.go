package group

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

func (s GenesisState) Validate() error {
	return s.Params.Validate()
}

// ExportGenesis returns a GenesisState for a given context and Keeper.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return GenesisState{
		Params: k.getParams(ctx),
	}
}
