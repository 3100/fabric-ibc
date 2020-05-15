package client

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/datachainlab/fabric-ibc/x/ibc/02-client/keeper"
)

const (
	AttributeKeyClientID  = types.AttributeKeyClientID
	AttrbuteKeyClientType = types.AttributeKeyClientType
	SubModuleName         = types.SubModuleName
	RouterKey             = types.RouterKey
	QuerierRoute          = types.QuerierRoute
	QueryAllClients       = types.QueryAllClients
	QueryClientState      = types.QueryClientState
	QueryConsensusState   = types.QueryConsensusState
)

var (
	// functions aliases
	NewKeeper                 = keeper.NewKeeper
	QuerierClients            = keeper.QuerierClients
	RegisterCodec             = types.RegisterCodec
	ErrClientExists           = types.ErrClientExists
	ErrClientNotFound         = types.ErrClientNotFound
	ErrClientFrozen           = types.ErrClientFrozen
	ErrConsensusStateNotFound = types.ErrConsensusStateNotFound
	ErrInvalidConsensus       = types.ErrInvalidConsensus
	ErrClientTypeNotFound     = types.ErrClientTypeNotFound
	ErrInvalidClientType      = types.ErrInvalidClientType
	ErrRootNotFound           = types.ErrRootNotFound
	ErrInvalidHeader          = types.ErrInvalidHeader
	ErrInvalidEvidence        = types.ErrInvalidEvidence
	DefaultGenesisState       = types.DefaultGenesisState
	NewGenesisState           = types.NewGenesisState
	NewClientConsensusStates  = types.NewClientConsensusStates

	// variable aliases
	SubModuleCdc           = types.SubModuleCdc
	EventTypeCreateClient  = types.EventTypeCreateClient
	EventTypeUpdateClient  = types.EventTypeUpdateClient
	AttributeValueCategory = types.AttributeValueCategory
)

type (
	Keeper          = keeper.Keeper
	StakingKeeper   = types.StakingKeeper
	GenesisState    = types.GenesisState
	ConsensusStates = types.ClientConsensusStates
)
