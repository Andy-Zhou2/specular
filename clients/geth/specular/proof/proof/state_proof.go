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

package proof

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/specularl2/specular/clients/geth/specular/proof/state"
)

type IntraStateProof struct {
	Depth                    uint16
	Gas                      uint64
	Refund                   uint64
	LastDepthHash            common.Hash
	ContractAddress          common.Address
	Caller                   common.Address
	Value                    uint256.Int
	CallFlag                 state.CallFlag
	Out                      uint64
	OutSize                  uint64
	Pc                       uint64
	OpCode                   vm.OpCode
	CodeHash                 common.Hash
	StackHash                common.Hash
	StackSize                uint64
	MemorySize               uint64
	MemoryRoot               common.Hash
	InputDataSize            uint64
	InputDataRoot            common.Hash
	ReturnDataSize           uint64
	ReturnDataRoot           common.Hash
	CommittedGlobalStateRoot common.Hash
	GlobalStateRoot          common.Hash
	SelfDestructAcc          common.Hash
	LogAcc                   common.Hash
	BlockHashRoot            common.Hash
	AccesslistRoot           common.Hash
}

func StateProofFromState(s *state.IntraState) *IntraStateProof {
	var lastDepthHash common.Hash
	if s.LastDepthState != nil {
		lastDepthHash = s.LastDepthState.Hash()
	}
	return &IntraStateProof{
		Depth:                    s.Depth,
		Gas:                      s.Gas,
		Refund:                   s.Refund,
		LastDepthHash:            lastDepthHash,
		ContractAddress:          s.ContractAddress,
		Caller:                   s.Caller,
		Value:                    s.Value,
		CallFlag:                 s.CallFlag,
		Out:                      s.Out,
		OutSize:                  s.OutSize,
		Pc:                       s.Pc,
		OpCode:                   s.OpCode,
		CodeHash:                 s.CodeHash,
		StackSize:                uint64(s.Stack.Len()),
		StackHash:                s.Stack.Hash(),
		MemorySize:               s.Memory.Size(),
		MemoryRoot:               s.Memory.Root(),
		InputDataSize:            s.InputData.Size(),
		InputDataRoot:            s.InputData.Root(),
		ReturnDataSize:           s.ReturnData.Size(),
		ReturnDataRoot:           s.ReturnData.Root(),
		CommittedGlobalStateRoot: s.CommittedGlobalState.GetRootForProof(),
		GlobalStateRoot:          s.GlobalState.GetRootForProof(),
		SelfDestructAcc:          s.SelfDestructSet.Hash,
		LogAcc:                   s.LogSeries.Hash(),
		BlockHashRoot:            s.BlockHashTree.Root(),
		AccesslistRoot:           s.AccessListTrie.Root(),
	}
}

func (s *IntraStateProof) Encode() []byte {
	proofLen := 2 + 8 + 8 + 8 + 1 + 32 + 8 + 32 + 8 + 8 + 32 + 32 + 32 + 32 + 32 + 32 // Depth, Gas, Refund, Pc, OpCode, CodeMerkle, StackSize, StackHash, MemorySize, ReturnDataSize, CommittedGlobalStateRoot, GlobalStateRoot, SelfDestructAcc, LogAcc, BlockHashRoot, AccesslistRoot
	if s.Depth != 1 {
		proofLen += 32 + 20 + 20 + 32 + 8 + 1 + 8 + 8 // LastDepthHash, ContractAddress, Caller, Value, CallFlag, Out, OutSize, InputDataSize
		if s.InputDataSize != 0 {
			proofLen += 32 // InputDataRoot
		}
	}
	if s.MemorySize != 0 {
		proofLen += 32 // MemoryRoot
	}
	if s.ReturnDataSize != 0 {
		proofLen += 32 // ReturnDataRoot
	}
	encoded := make([]byte, proofLen)
	depth := make([]byte, 2)
	binary.BigEndian.PutUint16(depth, s.Depth)
	gas := make([]byte, 8)
	binary.BigEndian.PutUint64(gas, s.Gas)
	refund := make([]byte, 8)
	binary.BigEndian.PutUint64(refund, s.Refund)
	pc := make([]byte, 8)
	binary.BigEndian.PutUint64(pc, s.Pc)
	stackSize := make([]byte, 8)
	binary.BigEndian.PutUint64(stackSize, s.StackSize)
	memSize := make([]byte, 8)
	binary.BigEndian.PutUint64(memSize, s.MemorySize)
	var inputDataSize []byte
	if s.Depth != 1 {
		inputDataSize = make([]byte, 8)
		binary.BigEndian.PutUint64(inputDataSize, s.InputDataSize)
	}
	returnDataSize := make([]byte, 8)
	binary.BigEndian.PutUint64(returnDataSize, s.ReturnDataSize)
	copy(encoded, depth)
	copy(encoded[2:], gas)
	copy(encoded[10:], refund)
	offset := 18
	if s.Depth != 1 {
		copy(encoded[offset:], s.LastDepthHash.Bytes())
		copy(encoded[offset+32:], s.ContractAddress.Bytes())
		copy(encoded[offset+32+20:], s.Caller.Bytes())
		valueBytes := s.Value.Bytes32()
		copy(encoded[offset+32+20+20:], valueBytes[:])
		out := make([]byte, 8)
		binary.BigEndian.PutUint64(out, s.Out)
		outSize := make([]byte, 8)
		binary.BigEndian.PutUint64(outSize, s.OutSize)
		encoded[offset+32+20+20+32] = byte(s.CallFlag)
		copy(encoded[offset+32+20+20+32+1:], out)
		copy(encoded[offset+32+20+20+32+1+8:], outSize)
		offset += 32 + 20 + 20 + 32 + 1 + 8 + 8
	}
	copy(encoded[offset:], pc)
	encoded[offset+8] = byte(s.OpCode)
	copy(encoded[offset+8+1:], s.CodeHash.Bytes())
	copy(encoded[offset+8+1+32:], stackSize)
	copy(encoded[offset+8+1+32+8:], s.StackHash.Bytes())
	copy(encoded[offset+8+1+32+8+32:], memSize)
	offset += 8 + 1 + 32 + 8 + 32 + 8
	if s.MemorySize != 0 {
		copy(encoded[offset:], s.MemoryRoot.Bytes())
		offset += 32
	}
	if s.Depth != 1 {
		copy(encoded[offset:], inputDataSize)
		offset += 8
		if s.InputDataSize != 0 {
			copy(encoded[offset:], s.InputDataRoot.Bytes())
			offset += 32
		}
	}
	copy(encoded[offset:], returnDataSize)
	offset += 8
	if s.ReturnDataSize != 0 {
		copy(encoded[offset:], s.ReturnDataRoot.Bytes())
		offset += 32
	}
	copy(encoded[offset:], s.CommittedGlobalStateRoot.Bytes())
	copy(encoded[offset+32:], s.GlobalStateRoot.Bytes())
	copy(encoded[offset+64:], s.SelfDestructAcc.Bytes())
	copy(encoded[offset+96:], s.LogAcc.Bytes())
	copy(encoded[offset+128:], s.BlockHashRoot.Bytes())
	copy(encoded[offset+160:], s.AccesslistRoot.Bytes())
	return encoded
}

func (s *IntraStateProof) Hash() common.Hash {
	return crypto.Keccak256Hash(s.Encode())
}

type InterStateProof struct {
	GlobalStateRoot     common.Hash
	CumulativeGasUsed   uint256.Int
	TransactionTireRoot common.Hash
	ReceiptTrieRoot     common.Hash
	BlockHashRoot       common.Hash
}

func InterStateProofFromInterState(s *state.InterState) *InterStateProof {
	return &InterStateProof{
		GlobalStateRoot:     s.GlobalState.GetRootForProof(),
		CumulativeGasUsed:   *s.CumulativeGasUsed,
		TransactionTireRoot: s.TransactionTrie.Root(),
		ReceiptTrieRoot:     s.ReceiptTrie.Root(),
		BlockHashRoot:       s.BlockHashTree.Root(),
	}
}

func (s *InterStateProof) Encode() []byte {
	encoded := make([]byte, 160)
	copy(encoded, s.GlobalStateRoot.Bytes())
	gasBytes := s.CumulativeGasUsed.Bytes32()
	copy(encoded[32:], gasBytes[:])
	copy(encoded[64:], s.TransactionTireRoot.Bytes())
	copy(encoded[96:], s.ReceiptTrieRoot.Bytes())
	copy(encoded[128:], s.BlockHashRoot.Bytes())
	return encoded
}

type BlockStateProof struct {
	GlobalStateRoot   common.Hash
	CumulativeGasUsed uint256.Int
	BlockHashRoot     common.Hash
}

func BlockStateProofFromBlockState(s *state.BlockState) *BlockStateProof {
	return &BlockStateProof{
		GlobalStateRoot:   s.GlobalState.GetRootForProof(),
		CumulativeGasUsed: *s.CumulativeGasUsed,
		BlockHashRoot:     s.BlockHashTree.Root(),
	}
}

func (s *BlockStateProof) Encode() []byte {
	encoded := make([]byte, 96)
	copy(encoded, s.GlobalStateRoot.Bytes())
	gasBytes := s.CumulativeGasUsed.Bytes32()
	copy(encoded[32:], gasBytes[:])
	copy(encoded[64:], s.BlockHashRoot.Bytes())
	return encoded
}
