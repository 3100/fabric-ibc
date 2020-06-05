package chaincode

import (
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/datachainlab/fabric-ibc/x/compat"
	fabric "github.com/datachainlab/fabric-ibc/x/ibc/xx-fabric"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/common"
	msppb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/policydsl"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	tmtime "github.com/tendermint/tendermint/types/time"
	dbm "github.com/tendermint/tm-db"
)

func TestApp(t *testing.T) {
	assert := assert.New(t)

	logger := log.NewTMLogger(os.Stdout)
	// var dbProvider = staticTMDBProvider{db: &traceDB{dbm.NewMemDB()}}.Provider
	var dbProvider = DefaultDBProvider
	runner := NewAppRunner(logger, dbProvider)
	stub := compat.MakeFakeStub()

	{
		msg := makeMsgCreateClient()
		assert.NoError(runner.RunMsg(stub, msg))
	}

	{
		msg := makeMsgUpdateClient()
		assert.NoError(runner.RunMsg(stub, msg))
	}
}

const (
	channelID = "dummyChannel"
	clientID  = "fabricclient"
)

var ccid = peer.ChaincodeID{
	Name:    "dummyCC",
	Version: "dummyVer",
}

func makeMsgCreateClient() string {
	cdc, _ := MakeCodecs()

	/// Build Msg
	ch := fabric.NewChaincodeHeader(1, tmtime.Now(), fabric.Proof{})
	var sigs [][]byte
	var pcBytes []byte
	ci := fabric.NewChaincodeInfo(channelID, ccid, pcBytes, sigs)

	h := fabric.NewHeader(ch, ci)
	prv := secp256k1.GenPrivKey()
	addr := prv.PubKey().Address()
	signer := sdk.AccAddress(addr)
	msg := fabric.NewMsgCreateClient(clientID, h, signer)
	if err := msg.ValidateBasic(); err != nil {
		panic(err)
	}
	/// END

	tx := auth.StdTx{
		Msgs: []sdk.Msg{msg},
		Signatures: []auth.StdSignature{
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
	return string(bz)
}

func makePolicy(mspids []string) []byte {
	return protoutil.MarshalOrPanic(&common.ApplicationPolicy{
		Type: &common.ApplicationPolicy_SignaturePolicy{
			SignaturePolicy: policydsl.SignedByNOutOfGivenRole(int32(len(mspids)/2+1), msppb.MSPRole_MEMBER, mspids),
		},
	})
}

func makeMsgUpdateClient() string {
	cdc, _ := MakeCodecs()

	/// Build Msg
	ch := fabric.NewChaincodeHeader(2, tmtime.Now(), fabric.Proof{})
	var sigs [][]byte
	var pcBytes []byte = makePolicy([]string{"Org1"})
	ci := fabric.NewChaincodeInfo(channelID, ccid, pcBytes, sigs)

	h := fabric.NewHeader(ch, ci)
	prv := secp256k1.GenPrivKey()
	addr := prv.PubKey().Address()
	signer := sdk.AccAddress(addr)
	msg := fabric.NewMsgUpdateClient(clientID, h, signer)
	if err := msg.ValidateBasic(); err != nil {
		panic(err)
	}
	/// END

	tx := auth.StdTx{
		Msgs: []sdk.Msg{msg},
		Signatures: []auth.StdSignature{
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
	return string(bz)
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