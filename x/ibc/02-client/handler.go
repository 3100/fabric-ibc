package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	fabric "github.com/datachainlab/fabric-ibc/x/ibc/xx-fabric"
)

// HandleMsgCreateClient defines the sdk.Handler for MsgCreateClient
func HandleMsgCreateClient(ctx sdk.Context, k Keeper, msg exported.MsgCreateClient) (*sdk.Result, error) {
	clientType := ClientTypeFromString(msg.GetClientType())

	var clientState exported.ClientState

	switch clientType {
	case exported.Tendermint:
		tmMsg, ok := msg.(ibctmtypes.MsgCreateClient)
		if !ok {
			return nil, sdkerrors.Wrap(ErrInvalidClientType, "Msg is not a Tendermint CreateClient msg")
		}
		var err error

		clientState, err = ibctmtypes.InitializeFromMsg(tmMsg)
		if err != nil {
			return nil, err
		}
	case exported.Localhost:
		// msg client id is always "localhost"
		clientState = localhosttypes.NewClientState(ctx.ChainID(), ctx.BlockHeight())
	case fabric.Fabric:
		fabMsg, ok := msg.(fabric.MsgCreateClient)
		if !ok {
			return nil, sdkerrors.Wrap(ErrInvalidClientType, "Msg is not a Fabric CreateClient msg")
		}
		clientState = fabric.NewClientState(fabMsg.GetClientID(), fabMsg.Header)
	default:
		return nil, sdkerrors.Wrap(ErrInvalidClientType, msg.GetClientType())
	}

	_, err := k.CreateClient(
		ctx, clientState, msg.GetConsensusState(),
	)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeCreateClient,
			sdk.NewAttribute(AttributeKeyClientID, msg.GetClientID()),
			sdk.NewAttribute(AttrbuteKeyClientType, msg.GetClientType()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgUpdateClient defines the sdk.Handler for MsgUpdateClient
func HandleMsgUpdateClient(ctx sdk.Context, k Keeper, msg exported.MsgUpdateClient) (*sdk.Result, error) {
	_, err := k.UpdateClient(ctx, msg.GetClientID(), msg.GetHeader())
	if err != nil {
		return nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandlerClientMisbehaviour defines the Evidence module handler for submitting a
// light client misbehaviour.
func HandlerClientMisbehaviour(k Keeper) evidencetypes.Handler {
	return func(ctx sdk.Context, evidence evidenceexported.Evidence) error {
		misbehaviour, ok := evidence.(exported.Misbehaviour)
		if !ok {
			return types.ErrInvalidEvidence
		}

		return k.CheckMisbehaviourAndUpdateState(ctx, misbehaviour)
	}
}

// This code is a copy from https://github.com/cosmos/cosmos-sdk/blob/6973c564572908c780becf74ffaad1ff04ae13e1/x/ibc/02-client/exported/exported.go#L209
// ClientTypeFromString returns a byte that corresponds to the registered client
// type. It returns 0 if the type is not found/registered.
func ClientTypeFromString(clientType string) exported.ClientType {
	switch clientType {
	case exported.ClientTypeTendermint:
		return exported.Tendermint
	case exported.ClientTypeLocalHost:
		return exported.Localhost
	case fabric.ClientTypeFabric:
		return fabric.Fabric
	default:
		return 0
	}
}
