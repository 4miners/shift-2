// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package runtime

import (
	"math/big"

	"github.com/shiftcurrency/shift/common"
	"github.com/shiftcurrency/shift/core"
	"github.com/shiftcurrency/shift/core/state"
	"github.com/shiftcurrency/shift/core/vm"
)

// Env is a basic runtime environment required for running the EVM.
type Env struct {
	depth int
	state *state.StateDB

	origin   common.Address
	shiftbase common.Address

	number     *big.Int
	time       *big.Int
	difficulty *big.Int
	nrgLimit   *big.Int

	logs []vm.StructLog

	getHashFn func(uint64) common.Hash
}

// NewEnv returns a new vm.Environment
func NewEnv(cfg *Config, state *state.StateDB) vm.Environment {
	return &Env{
		state:      state,
		origin:     cfg.Origin,
		shiftbase:   cfg.Coinbase,
		number:     cfg.BlockNumber,
		time:       cfg.Time,
		difficulty: cfg.Difficulty,
		nrgLimit:   cfg.NrgLimit,
	}
}

func (self *Env) StructLogs() []vm.StructLog {
	return self.logs
}

func (self *Env) AddStructLog(log vm.StructLog) {
	self.logs = append(self.logs, log)
}

func (self *Env) Origin() common.Address   { return self.origin }
func (self *Env) BlockNumber() *big.Int    { return self.number }
func (self *Env) Coinbase() common.Address { return self.shiftbase }
func (self *Env) Time() *big.Int           { return self.time }
func (self *Env) Difficulty() *big.Int     { return self.difficulty }
func (self *Env) Db() vm.Database          { return self.state }
func (self *Env) NrgLimit() *big.Int       { return self.nrgLimit }
func (self *Env) VmType() vm.Type          { return vm.StdVmTy }
func (self *Env) GetHash(n uint64) common.Hash {
	return self.getHashFn(n)
}
func (self *Env) AddLog(log *vm.Log) {
	self.state.AddLog(log)
}
func (self *Env) Depth() int     { return self.depth }
func (self *Env) SetDepth(i int) { self.depth = i }
func (self *Env) CanTransfer(from common.Address, balance *big.Int) bool {
	return self.state.GetBalance(from).Cmp(balance) >= 0
}
func (self *Env) MakeSnapshot() vm.Database {
	return self.state.Copy()
}
func (self *Env) SetSnapshot(copy vm.Database) {
	self.state.Set(copy.(*state.StateDB))
}

func (self *Env) Transfer(from, to vm.Account, amount *big.Int) {
	core.Transfer(from, to, amount)
}

func (self *Env) Call(caller vm.ContractRef, addr common.Address, data []byte, nrg, price, value *big.Int) ([]byte, error) {
	return core.Call(self, caller, addr, data, nrg, price, value)
}
func (self *Env) CallCode(caller vm.ContractRef, addr common.Address, data []byte, nrg, price, value *big.Int) ([]byte, error) {
	return core.CallCode(self, caller, addr, data, nrg, price, value)
}

func (self *Env) DelegateCall(me vm.ContractRef, addr common.Address, data []byte, nrg, price *big.Int) ([]byte, error) {
	return core.DelegateCall(self, me, addr, data, nrg, price)
}

func (self *Env) Create(caller vm.ContractRef, data []byte, nrg, price, value *big.Int) ([]byte, common.Address, error) {
	return core.Create(self, caller, data, nrg, price, value)
}
