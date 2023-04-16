// Copyright 2022, Specular contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prover

import (
	"context"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/specularl2/specular/clients/geth/specular/prover/state"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error)
	GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error)
	GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error)
	RPCGasCap() uint64
	ChainConfig() *params.ChainConfig
	Engine() consensus.Engine
	ChainDb() ethdb.Database
	// StateAtBlock returns the state corresponding to the stateroot of the block.
	// N.B: For executing transactions on block N, the required stateRoot is block N-1,
	// so this method should be called with the parent.
	StateAtBlock(ctx context.Context, block *types.Block, reexec uint64, base state.SpecularState, checkLive, preferDisk bool) (state.SpecularState, error)
	StateAtTransaction(ctx context.Context, block *types.Block, txIndex int, reexec uint64) (core.Message, state.SpecularBlockContext, state.SpecularState, error)

	// functions from package vm:
	NewEVMBlockContext(header *types.Header, chain core.ChainContext, author *common.Address) state.SpecularBlockContext
	NewEVM(blockCtx state.SpecularBlockContext, txCtx vm.TxContext, statedb state.SpecularState, chainConfig *params.ChainConfig, config state.SpecularConfig) state.SpecularEVM

	// functions from package core:
	ApplyMessage(evm state.SpecularEVM, msg core.Message, gp *core.GasPool) (*core.ExecutionResult, error)
}

// ProverAPI is the collection of Specular one-step proof APIs.
type ProverAPI struct {
	backend Backend
}

// NewAPI creates a new API definition for the Specular one-step proof services.
func NewAPI(backend Backend) *ProverAPI {
	return &ProverAPI{backend: backend}
}

type chainContext struct {
	backend Backend
	ctx     context.Context
}

func (context *chainContext) Engine() consensus.Engine {
	return context.backend.Engine()
}

func (context *chainContext) GetHeader(hash common.Hash, number uint64) *types.Header {
	header, err := context.backend.HeaderByNumber(context.ctx, rpc.BlockNumber(number))
	if err != nil {
		return nil
	}
	if header.Hash() == hash {
		return header
	}
	header, err = context.backend.HeaderByHash(context.ctx, hash)
	if err != nil {
		return nil
	}
	return header
}

func createChainContext(backend Backend, ctx context.Context) core.ChainContext {
	return &chainContext{backend: backend, ctx: ctx}
}

func (api *ProverAPI) ProveTransaction(ctx context.Context, hash common.Hash, target common.Hash, config *ProverConfig) (hexutil.Bytes, error) {
	return hexutil.Bytes{}, nil
}

func (api *ProverAPI) ProveBlocksForBenchmark(ctx context.Context, startGasUsed *big.Int, startNum, endNum uint64, config *ProverConfig) ([]hexutil.Bytes, error) {
	states, err := GenerateStates(api.backend, ctx, startGasUsed, startNum, endNum, config)
	if err != nil {
		return nil, err
	}
	var proofs []hexutil.Bytes
	for _, s := range states {
		log.Info("Generate for ", "state", s)
		proof, err := GenerateProof(api.backend, ctx, s, config)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, proof.Encode())
	}
	return proofs, nil
}

func (api *ProverAPI) GenerateStateHashes(ctx context.Context, startGasUsed *big.Int, startNum, endNum uint64, config *ProverConfig) ([]common.Hash, error) {
	states, err := GenerateStates(api.backend, ctx, startGasUsed, startNum, endNum, config)
	if err != nil {
		return nil, err
	}
	hashes := make([]common.Hash, len(states))
	for i, state := range states {
		hashes[i] = state.Hash()
	}
	return hashes, nil
}

// APIs return the collection of RPC services the tracer package offers.
func APIs(backend Backend) []rpc.API {
	// Append all the local APIs and return
	return []rpc.API{
		{
			Namespace: "proof",
			Version:   "1.0",
			Service:   NewAPI(backend),
			Public:    false,
		},
	}
}