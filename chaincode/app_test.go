package chaincode_test

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/datachainlab/fabric-ibc/app"
	"github.com/datachainlab/fabric-ibc/example"
	"github.com/datachainlab/fabric-ibc/x/compat"
	fabrictests "github.com/datachainlab/fabric-ibc/x/ibc/xx-fabric/tests"
	fabrictypes "github.com/datachainlab/fabric-ibc/x/ibc/xx-fabric/types"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/common"
	msppb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/common/policydsl"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtime "github.com/tendermint/tendermint/types/time"
	dbm "github.com/tendermint/tm-db"
)

func TestApp(t *testing.T) {
	const (
		clientID0     = "ibcclient0"
		connectionID0 = "connection0"
		portID0       = "transfer"
		channelID0    = "channelid0"
		channelOrder0 = channel.ORDERED

		clientID1     = "ibcclient1"
		connectionID1 = "connection1"
		portID1       = "transfer"
		channelID1    = "channelid1"
		channelOrder1 = channel.ORDERED
	)

	const (
		fabchannelID = "dummyChannel"
	)

	var ccid = fabrictypes.ChaincodeID{
		Name:    "dummyCC",
		Version: "dummyVer",
	}

	require := require.New(t)

	conf, err := fabrictypes.DefaultConfig()
	require.NoError(err)
	// setup the MSP manager so that we can sign/verify
	lcMSP, err := fabrictests.GetLocalMsp(conf.MSPsDir, "SampleOrgMSP")
	require.NoError(err)
	endorser, err := lcMSP.GetDefaultSigningIdentity()
	require.NoError(err)

	app0Tk := newTimeKeeper()
	app1Tk := newTimeKeeper()
	stub0 := compat.MakeFakeStub()
	stub0.GetTxTimestampStub = func() (*timestamp.Timestamp, error) {
		return &timestamp.Timestamp{Seconds: app0Tk.Now().Unix()}, nil
	}
	ctx0 := mockContext{stub: stub0}
	stub1 := compat.MakeFakeStub()
	stub1.GetTxTimestampStub = func() (*timestamp.Timestamp, error) {
		return &timestamp.Timestamp{Seconds: app1Tk.Now().Unix()}, nil
	}
	ctx1 := mockContext{stub: stub1}

	prv0 := secp256k1.GenPrivKey()
	prv1 := secp256k1.GenPrivKey()

	app0 := MakeTestChaincodeApp(prv0, fabchannelID, ccid, endorser, clientID0, connectionID0, portID0, channelID0, channelOrder0)
	require.NoError(app0.init(ctx0))
	app1 := MakeTestChaincodeApp(prv1, fabchannelID, ccid, endorser, clientID1, connectionID1, portID1, channelID1, channelOrder1)
	require.NoError(app1.init(ctx1))

	// Create Clients
	require.NoError(app0.runMsg(stub0, app0.createMsgCreateClient(t, ctx0)))
	require.NoError(app1.runMsg(stub1, app1.createMsgCreateClient(t, ctx1)))

	// Update Clients
	app0Tk.Add(5 * time.Second)
	_, err = app0.updateSequence(ctx0)
	require.NoError(err)
	require.NoError(app0.runMsg(stub0, app0.createMsgUpdateClient(t)))

	app1Tk.Add(5 * time.Second)
	_, err = app1.updateSequence(ctx1)
	require.NoError(err)
	require.NoError(app1.runMsg(stub1, app1.createMsgUpdateClient(t)))

	// Create connection
	require.NoError(app0.runMsg(stub0, app0.createMsgConnectionOpenInit(t, app1)))
	require.NoError(app1.runMsg(stub1, app1.createMsgConnectionOpenTry(t, ctx0, app0)))
	require.NoError(app0.runMsg(stub0, app0.createMsgConnectionOpenAck(t, ctx1, app1)))
	require.NoError(app1.runMsg(stub1, app1.createMsgConnectionOpenConfirm(t, ctx0, app0)))

	// Create channel
	require.NoError(app0.runMsg(stub0, app0.createMsgChannelOpenInit(t, app1)))
	require.NoError(app1.runMsg(stub1, app1.createMsgChannelOpenTry(t, ctx0, app0)))
	require.NoError(app0.runMsg(stub0, app0.createMsgChannelOpenAck(t, ctx1, app1)))
	require.NoError(app1.runMsg(stub1, app1.createMsgChannelOpenConfirm(t, ctx0, app0)))

	{
		bz := channeltypes.SubModuleCdc.MustMarshalJSON(channeltypes.QueryAllChannelsParams{Limit: 100, Page: 1})
		res, err := app0.query(ctx0, app.RequestQuery{Data: string(bz), Path: "/custom/ibc/channel/channels"})
		require.NoError(err)
		var channels []channeltypes.IdentifiedChannel
		channeltypes.SubModuleCdc.MustUnmarshalJSON([]byte(res.Value), &channels)
		require.Equal(1, len(channels))
		require.Equal(channels[0].ID, channelID0)
	}
	{
		bz := channeltypes.SubModuleCdc.MustMarshalJSON(channeltypes.QueryAllChannelsParams{Limit: 100, Page: 1})
		res, err := app1.query(ctx1, app.RequestQuery{Data: string(bz), Path: "/custom/ibc/channel/channels"})
		require.NoError(err)
		var channels []channeltypes.IdentifiedChannel
		channeltypes.SubModuleCdc.MustUnmarshalJSON([]byte(res.Value), &channels)
		require.Equal(1, len(channels))
		require.Equal(channels[0].ID, channelID1)
	}

	var createPacket = func(src, dst TestChaincodeApp, coins sdk.Coins, seq, timeoutHeight, timeoutTimestamp uint64) channel.Packet {
		data := ibctransfertypes.NewFungibleTokenPacketData(coins, src.signer.String(), src.signer.String())
		return channel.NewPacket(data.GetBytes(), seq, src.portID, src.channelID, dst.portID, dst.channelID, timeoutHeight, timeoutTimestamp)
	}

	// Setup transfer
	// https://github.com/cosmos/cosmos-sdk/blob/24b9be0ef841303a2e2b6f60042b5da3b74af2ef/x/ibc-transfer/keeper/relay_test.go#L21
	addr := sdk.AccAddress(example.MasterAccount.PubKey().Address())
	denom := fmt.Sprintf("%v/%v/ftk", app1.portID, app1.channelID)
	coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))
	app0.signer = addr

	// check if nextSequence equals an expected value
	// app0
	res, err := app0.query(ctx0, app.RequestQuery{Data: string(host.KeyNextSequenceSend(app0.portID, app0.channelID)), Path: "/store/ibc/key"})
	require.NoError(err)
	require.Equal(uint64(1), binary.BigEndian.Uint64([]byte(res.Value)))

	// Success
	require.NoError(app0.runMsg(stub0, app0.createMsgTransfer(t, app1, coins, addr, 1000, 0)))
	packet0 := createPacket(app0, app1, coins, 1, 1000, 0)
	require.NoError(app1.runMsg(stub1, app1.createMsgPacketForTransfer(t, ctx0, app0, packet0)))
	require.NoError(app0.runMsg(stub0, app0.createMsgAcknowledgement(t, ctx1, app1, packet0)))

	// check if nextSequence equals an expected value
	// app0
	res, err = app0.query(ctx0, app.RequestQuery{Data: string(host.KeyNextSequenceSend(app0.portID, app0.channelID)), Path: "/store/ibc/key"})
	require.NoError(err)
	require.Equal(uint64(2), binary.BigEndian.Uint64([]byte(res.Value)))
	// app1
	res, err = app1.query(ctx1, app.RequestQuery{Data: string(host.KeyNextSequenceRecv(app1.portID, app1.channelID)), Path: "/store/ibc/key"})
	require.NoError(err)
	require.Equal(uint64(2), binary.BigEndian.Uint64([]byte(res.Value)))

	// Timeout
	var timeoutHeight uint64 = 3
	require.NoError(app0.runMsg(stub0, app0.createMsgTransfer(t, app1, coins, addr, timeoutHeight, 0)))
	packet1 := createPacket(app0, app1, coins, 2, timeoutHeight, 0)

	// Update Clients
	{
		app0Tk.Add(5 * time.Second)
		_, err = app0.updateSequence(ctx0)
		require.NoError(err)
		require.NoError(app0.runMsg(stub0, app0.createMsgUpdateClient(t)))

		app1Tk.Add(5 * time.Second)
		_, err = app1.updateSequence(ctx1)
		require.NoError(err)
		require.NoError(app1.runMsg(stub1, app1.createMsgUpdateClient(t)))
	}

	require.NoError(app0.runMsg(stub0, app0.createMsgTimeoutPacket(t, ctx1, app1, 2, channel.ORDERED, packet1)))
}

type mockContext struct {
	stub shim.ChaincodeStubInterface
}

func (c mockContext) GetStub() shim.ChaincodeStubInterface {
	return c.stub
}

func (c mockContext) GetClientIdentity() cid.ClientIdentity {
	panic("failed to get client identity")
}

func makePolicy(mspids []string) []byte {
	return protoutil.MarshalOrPanic(&common.ApplicationPolicy{
		Type: &common.ApplicationPolicy_SignaturePolicy{
			SignaturePolicy: policydsl.SignedByNOutOfGivenRole(int32(len(mspids)/2+1), msppb.MSPRole_MEMBER, mspids),
		},
	})
}

func makeStdTxBytes(cdc codec.Marshaler, prv crypto.PrivKey, msgs ...sdk.Msg) []byte {
	tx := authtypes.StdTx{
		Msgs: msgs,
		Signatures: []authtypes.StdSignature{
			{PubKey: prv.PubKey().Bytes(), Signature: make([]byte, 64)}, // FIXME set valid signature
		},
	}
	if err := tx.ValidateBasic(); err != nil {
		panic(err)
	}

	bz, err := cdc.MarshalJSON(tx)
	if err != nil {
		panic(err)
	}
	return bz
}

func tmDBProvider(_ shim.ChaincodeStubInterface) dbm.DB {
	return dbm.NewMemDB()
}

type staticTMDBProvider struct {
	db dbm.DB
}

func (p staticTMDBProvider) Provider(_ shim.ChaincodeStubInterface) dbm.DB {
	return p.db
}

type timeKeeper struct {
	tm time.Time
}

func newTimeKeeper() *timeKeeper {
	return &timeKeeper{tm: tmtime.Now()}
}

func (tk timeKeeper) Now() time.Time {
	return tk.tm
}

func (tk *timeKeeper) Add(d time.Duration) {
	tk.tm = tk.tm.Add(d)
}
