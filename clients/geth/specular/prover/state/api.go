package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type SpecularState interface {
	Prepare(thash common.Hash, ti int) SpecularState
	Copy() SpecularState
	GetRootForProof() common.Hash
	GetRefund() uint64
	CommitForProof()
	GetCurrentLogs() []*types.Log
	GetCode(address common.Address) []byte
	GetProof(common.Address) ([][]byte, error)
	GetStorageProof(common.Address, common.Hash) ([][]byte, error)
	SubBalance(common.Address, *big.Int)
	SetNonce(common.Address, uint64)
	GetNonce(common.Address) uint64
	AddBalance(common.Address, *big.Int)
	DeleteSuicidedAccountForProof(addr common.Address)
	SetCode(common.Address, []byte)
	GetBalance(common.Address) *big.Int
	GetCodeHash(common.Address) common.Hash
}

type (
	// CanTransferFunc is the signature of a transfer guard function
	CanTransferFunc func(SpecularState, common.Address, *big.Int) bool
	// GetHashFunc returns the n'th block hash in the blockchain
	// and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) common.Hash
)

type SpecularBlockContext struct { // The functions are modified, using SpecularState instead of StateDB
	// CanTransfer returns whether the account contains
	// sufficient ether to transfer the value
	CanTransfer CanTransferFunc
	// GetHash returns the hash corresponding to n
	GetHash GetHashFunc

	// Block information
	Coinbase    common.Address // Provides information for COINBASE
	GasLimit    uint64         // Provides information for GASLIMIT
	BlockNumber *big.Int       // Provides information for NUMBER
	Time        *big.Int       // Provides information for TIME
	Difficulty  *big.Int       // Provides information for DIFFICULTY
	BaseFee     *big.Int       // Provides information for BASEFEE
	Random      *common.Hash   // Provides information for RANDOM
}

type SpecularEVMLoggerInterface interface {
}

// Config are the configuration options for the Interpreter
type SpecularConfig struct {
	Debug  bool                       // Enables debugging
	Tracer SpecularEVMLoggerInterface // Opcode logger
}

type SpecularEVM struct {
	// Context provides auxiliary blockchain related information
	Context SpecularBlockContext
	// StateDB gives access to the underlying state
	StateDB SpecularState
}
