// Copyright 2015 The shift Authors
// This file is part of the shift library.
//
// The shift library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The shift library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the shift library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"bytes"
	"encoding/json"
	"math/big"

	"fmt"

	"github.com/shiftcurrency/shift/common"
	"github.com/shiftcurrency/shift/common/natspec"
	"github.com/shiftcurrency/shift/shf"
	"github.com/shiftcurrency/shift/rlp"
	"github.com/shiftcurrency/shift/rpc/codec"
	"github.com/shiftcurrency/shift/rpc/shared"
	"github.com/shiftcurrency/shift/xeth"
	"gopkg.in/fatih/set.v0"
)

const (
	EthApiVersion = "1.0"
)

// shf api provider
// See https://github.com/shiftcurrency/wiki/wiki/JSON-RPC
type ethApi struct {
	xeth     *xeth.XEth
	shift *shf.Shift
	methods  map[string]ethhandler
	codec    codec.ApiCoder
}

// shf callback handler
type ethhandler func(*ethApi, *shared.Request) (interface{}, error)

var (
	ethMapping = map[string]ethhandler{
		"shf_accounts":                            (*ethApi).Accounts,
		"shf_blockNumber":                         (*ethApi).BlockNumber,
		"shf_getBalance":                          (*ethApi).GetBalance,
		"shf_protocolVersion":                     (*ethApi).ProtocolVersion,
		"shf_shiftbase":                            (*ethApi).Coinbase,
		"shf_mining":                              (*ethApi).IsMining,
		"shf_syncing":                             (*ethApi).IsSyncing,
		"shf_nrgPrice":                            (*ethApi).NrgPrice,
		"shf_getStorage":                          (*ethApi).GetStorage,
		"shf_storageAt":                           (*ethApi).GetStorage,
		"shf_getStorageAt":                        (*ethApi).GetStorageAt,
		"shf_getTransactionCount":                 (*ethApi).GetTransactionCount,
		"shf_getBlockTransactionCountByHash":      (*ethApi).GetBlockTransactionCountByHash,
		"shf_getBlockTransactionCountByNumber":    (*ethApi).GetBlockTransactionCountByNumber,
		"shf_getUncleCountByBlockHash":            (*ethApi).GetUncleCountByBlockHash,
		"shf_getUncleCountByBlockNumber":          (*ethApi).GetUncleCountByBlockNumber,
		"shf_getData":                             (*ethApi).GetData,
		"shf_getCode":                             (*ethApi).GetData,
		"shf_sign":                                (*ethApi).Sign,
		"shf_sendRawTransaction":                  (*ethApi).SendTransaction,
		"shf_sendTransaction":                     (*ethApi).SendTransaction,
		"shf_transact":                            (*ethApi).SendTransaction,
		"shf_estimateNrg":                         (*ethApi).EstimateNrg,
		"shf_call":                                (*ethApi).Call,
		"shf_flush":                               (*ethApi).Flush,
		"shf_getBlockByHash":                      (*ethApi).GetBlockByHash,
		"shf_getBlockByNumber":                    (*ethApi).GetBlockByNumber,
		"shf_getTransactionByHash":                (*ethApi).GetTransactionByHash,
		"shf_getTransactionByBlockNumberAndIndex": (*ethApi).GetTransactionByBlockNumberAndIndex,
		"shf_getTransactionByBlockHashAndIndex":   (*ethApi).GetTransactionByBlockHashAndIndex,
		"shf_getUncleByBlockHashAndIndex":         (*ethApi).GetUncleByBlockHashAndIndex,
		"shf_getUncleByBlockNumberAndIndex":       (*ethApi).GetUncleByBlockNumberAndIndex,
		"shf_getCompilers":                        (*ethApi).GetCompilers,
		"shf_compileSolidity":                     (*ethApi).CompileSolidity,
		"shf_newFilter":                           (*ethApi).NewFilter,
		"shf_newBlockFilter":                      (*ethApi).NewBlockFilter,
		"shf_newPendingTransactionFilter":         (*ethApi).NewPendingTransactionFilter,
		"shf_uninstallFilter":                     (*ethApi).UninstallFilter,
		"shf_getFilterChanges":                    (*ethApi).GetFilterChanges,
		"shf_getFilterLogs":                       (*ethApi).GetFilterLogs,
		"shf_getLogs":                             (*ethApi).GetLogs,
		"shf_hashrate":                            (*ethApi).Hashrate,
		"shf_getWork":                             (*ethApi).GetWork,
		"shf_submitWork":                          (*ethApi).SubmitWork,
		"shf_submitHashrate":                      (*ethApi).SubmitHashrate,
		"shf_resend":                              (*ethApi).Resend,
		"shf_pendingTransactions":                 (*ethApi).PendingTransactions,
		"shf_getTransactionReceipt":               (*ethApi).GetTransactionReceipt,
	}
)

// create new ethApi instance
func NewEthApi(xeth *xeth.XEth, shf *shf.Shift, codec codec.Codec) *ethApi {
	return &ethApi{xeth, shf, ethMapping, codec.New(nil)}
}

// collection with supported methods
func (self *ethApi) Methods() []string {
	methods := make([]string, len(self.methods))
	i := 0
	for k := range self.methods {
		methods[i] = k
		i++
	}
	return methods
}

// Execute given request
func (self *ethApi) Execute(req *shared.Request) (interface{}, error) {
	if callback, ok := self.methods[req.Method]; ok {
		return callback(self, req)
	}

	return nil, shared.NewNotImplementedError(req.Method)
}

func (self *ethApi) Name() string {
	return shared.EthApiName
}

func (self *ethApi) ApiVersion() string {
	return EthApiVersion
}

func (self *ethApi) Accounts(req *shared.Request) (interface{}, error) {
	return self.xeth.Accounts(), nil
}

func (self *ethApi) Hashrate(req *shared.Request) (interface{}, error) {
	return newHexNum(self.xeth.HashRate()), nil
}

func (self *ethApi) BlockNumber(req *shared.Request) (interface{}, error) {
	num := self.xeth.CurrentBlock().Number()
	return newHexNum(num.Bytes()), nil
}

func (self *ethApi) GetBalance(req *shared.Request) (interface{}, error) {
	args := new(GetBalanceArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	return self.xeth.AtStateNum(args.BlockNumber).BalanceAt(args.Address), nil
}

func (self *ethApi) ProtocolVersion(req *shared.Request) (interface{}, error) {
	return self.xeth.EthVersion(), nil
}

func (self *ethApi) Coinbase(req *shared.Request) (interface{}, error) {
	return newHexData(self.xeth.Coinbase()), nil
}

func (self *ethApi) IsMining(req *shared.Request) (interface{}, error) {
	return self.xeth.IsMining(), nil
}

func (self *ethApi) IsSyncing(req *shared.Request) (interface{}, error) {
	origin, current, height := self.shift.Downloader().Progress()
	if current < height {
		return map[string]interface{}{
			"startingBlock": newHexNum(big.NewInt(int64(origin)).Bytes()),
			"currentBlock":  newHexNum(big.NewInt(int64(current)).Bytes()),
			"highestBlock":  newHexNum(big.NewInt(int64(height)).Bytes()),
		}, nil
	}
	return false, nil
}

func (self *ethApi) NrgPrice(req *shared.Request) (interface{}, error) {
	return newHexNum(self.xeth.DefaultNrgPrice().Bytes()), nil
}

func (self *ethApi) GetStorage(req *shared.Request) (interface{}, error) {
	args := new(GetStorageArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	return self.xeth.AtStateNum(args.BlockNumber).State().SafeGet(args.Address).Storage(), nil
}

func (self *ethApi) GetStorageAt(req *shared.Request) (interface{}, error) {
	args := new(GetStorageAtArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	return self.xeth.AtStateNum(args.BlockNumber).StorageAt(args.Address, args.Key), nil
}

func (self *ethApi) GetTransactionCount(req *shared.Request) (interface{}, error) {
	args := new(GetTxCountArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	count := self.xeth.AtStateNum(args.BlockNumber).TxCountAt(args.Address)
	return fmt.Sprintf("%#x", count), nil
}

func (self *ethApi) GetBlockTransactionCountByHash(req *shared.Request) (interface{}, error) {
	args := new(HashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	block := self.xeth.EthBlockByHash(args.Hash)
	if block == nil {
		return nil, nil
	}
	return fmt.Sprintf("%#x", len(block.Transactions())), nil
}

func (self *ethApi) GetBlockTransactionCountByNumber(req *shared.Request) (interface{}, error) {
	args := new(BlockNumArg)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	block := self.xeth.EthBlockByNumber(args.BlockNumber)
	if block == nil {
		return nil, nil
	}
	return fmt.Sprintf("%#x", len(block.Transactions())), nil
}

func (self *ethApi) GetUncleCountByBlockHash(req *shared.Request) (interface{}, error) {
	args := new(HashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	block := self.xeth.EthBlockByHash(args.Hash)
	if block == nil {
		return nil, nil
	}
	return fmt.Sprintf("%#x", len(block.Uncles())), nil
}

func (self *ethApi) GetUncleCountByBlockNumber(req *shared.Request) (interface{}, error) {
	args := new(BlockNumArg)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	block := self.xeth.EthBlockByNumber(args.BlockNumber)
	if block == nil {
		return nil, nil
	}
	return fmt.Sprintf("%#x", len(block.Uncles())), nil
}

func (self *ethApi) GetData(req *shared.Request) (interface{}, error) {
	args := new(GetDataArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	v := self.xeth.AtStateNum(args.BlockNumber).CodeAtBytes(args.Address)
	return newHexData(v), nil
}

func (self *ethApi) Sign(req *shared.Request) (interface{}, error) {
	args := new(NewSigArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	v, err := self.xeth.Sign(args.From, args.Data, false)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (self *ethApi) SubmitTransaction(req *shared.Request) (interface{}, error) {
	args := new(NewDataArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	v, err := self.xeth.PushTx(args.Data)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// JsonTransaction is returned as response by the JSON RPC. It contains the
// signed RLP encoded transaction as Raw and the signed transaction object as Tx.
type JsonTransaction struct {
	Raw string `json:"raw"`
	Tx  *tx    `json:"tx"`
}

func (self *ethApi) SignTransaction(req *shared.Request) (interface{}, error) {
	args := new(NewTxArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	// nonce may be nil ("guess" mode)
	var nonce string
	if args.Nonce != nil {
		nonce = args.Nonce.String()
	}

	var nrg, price string
	if args.Nrg != nil {
		nrg = args.Nrg.String()
	}
	if args.NrgPrice != nil {
		price = args.NrgPrice.String()
	}
	tx, err := self.xeth.SignTransaction(args.From, args.To, nonce, args.Value.String(), nrg, price, args.Data)
	if err != nil {
		return nil, err
	}

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}

	return JsonTransaction{"0x" + common.Bytes2Hex(data), newTx(tx)}, nil
}

func (self *ethApi) SendTransaction(req *shared.Request) (interface{}, error) {
	args := new(NewTxArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	// nonce may be nil ("guess" mode)
	var nonce string
	if args.Nonce != nil {
		nonce = args.Nonce.String()
	}

	var nrg, price string
	if args.Nrg != nil {
		nrg = args.Nrg.String()
	}
	if args.NrgPrice != nil {
		price = args.NrgPrice.String()
	}
	v, err := self.xeth.Transact(args.From, args.To, nonce, args.Value.String(), nrg, price, args.Data)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (self *ethApi) GetNatSpec(req *shared.Request) (interface{}, error) {
	args := new(NewTxArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	var jsontx = fmt.Sprintf(`{"params":[{"to":"%s","data": "%s"}]}`, args.To, args.Data)
	notice := natspec.GetNotice(self.xeth, jsontx, self.shift.HTTPClient())

	return notice, nil
}

func (self *ethApi) EstimateNrg(req *shared.Request) (interface{}, error) {
	_, nrg, err := self.doCall(req.Params)
	if err != nil {
		return nil, err
	}

	// TODO unwrap the parent method's ToHex call
	if len(nrg) == 0 {
		return newHexNum(0), nil
	} else {
		return newHexNum(common.String2Big(nrg)), err
	}
}

func (self *ethApi) Call(req *shared.Request) (interface{}, error) {
	v, _, err := self.doCall(req.Params)
	if err != nil {
		return nil, err
	}

	// TODO unwrap the parent method's ToHex call
	if v == "0x0" {
		return newHexData([]byte{}), nil
	} else {
		return newHexData(common.FromHex(v)), nil
	}
}

func (self *ethApi) Flush(req *shared.Request) (interface{}, error) {
	return nil, shared.NewNotImplementedError(req.Method)
}

func (self *ethApi) doCall(params json.RawMessage) (string, string, error) {
	args := new(CallArgs)
	if err := self.codec.Decode(params, &args); err != nil {
		return "", "", err
	}

	return self.xeth.AtStateNum(args.BlockNumber).Call(args.From, args.To, args.Value.String(), args.Nrg.String(), args.NrgPrice.String(), args.Data)
}

func (self *ethApi) GetBlockByHash(req *shared.Request) (interface{}, error) {
	args := new(GetBlockByHashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	block := self.xeth.EthBlockByHash(args.BlockHash)
	if block == nil {
		return nil, nil
	}
	return NewBlockRes(block, self.xeth.Td(block.Hash()), args.IncludeTxs), nil
}

func (self *ethApi) GetBlockByNumber(req *shared.Request) (interface{}, error) {
	args := new(GetBlockByNumberArgs)
	if err := json.Unmarshal(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	block := self.xeth.EthBlockByNumber(args.BlockNumber)
	if block == nil {
		return nil, nil
	}
	return NewBlockRes(block, self.xeth.Td(block.Hash()), args.IncludeTxs), nil
}

func (self *ethApi) GetTransactionByHash(req *shared.Request) (interface{}, error) {
	args := new(HashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	tx, bhash, bnum, txi := self.xeth.EthTransactionByHash(args.Hash)
	if tx != nil {
		v := NewTransactionRes(tx)
		// if the blockhash is 0, assume this is a pending transaction
		if bytes.Compare(bhash.Bytes(), bytes.Repeat([]byte{0}, 32)) != 0 {
			v.BlockHash = newHexData(bhash)
			v.BlockNumber = newHexNum(bnum)
			v.TxIndex = newHexNum(txi)
		}
		return v, nil
	}
	return nil, nil
}

func (self *ethApi) GetTransactionByBlockHashAndIndex(req *shared.Request) (interface{}, error) {
	args := new(HashIndexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	raw := self.xeth.EthBlockByHash(args.Hash)
	if raw == nil {
		return nil, nil
	}
	block := NewBlockRes(raw, self.xeth.Td(raw.Hash()), true)
	if args.Index >= int64(len(block.Transactions)) || args.Index < 0 {
		return nil, nil
	} else {
		return block.Transactions[args.Index], nil
	}
}

func (self *ethApi) GetTransactionByBlockNumberAndIndex(req *shared.Request) (interface{}, error) {
	args := new(BlockNumIndexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	raw := self.xeth.EthBlockByNumber(args.BlockNumber)
	if raw == nil {
		return nil, nil
	}
	block := NewBlockRes(raw, self.xeth.Td(raw.Hash()), true)
	if args.Index >= int64(len(block.Transactions)) || args.Index < 0 {
		// return NewValidationError("Index", "does not exist")
		return nil, nil
	}
	return block.Transactions[args.Index], nil
}

func (self *ethApi) GetUncleByBlockHashAndIndex(req *shared.Request) (interface{}, error) {
	args := new(HashIndexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	raw := self.xeth.EthBlockByHash(args.Hash)
	if raw == nil {
		return nil, nil
	}
	block := NewBlockRes(raw, self.xeth.Td(raw.Hash()), false)
	if args.Index >= int64(len(block.Uncles)) || args.Index < 0 {
		// return NewValidationError("Index", "does not exist")
		return nil, nil
	}
	return block.Uncles[args.Index], nil
}

func (self *ethApi) GetUncleByBlockNumberAndIndex(req *shared.Request) (interface{}, error) {
	args := new(BlockNumIndexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	raw := self.xeth.EthBlockByNumber(args.BlockNumber)
	if raw == nil {
		return nil, nil
	}
	block := NewBlockRes(raw, self.xeth.Td(raw.Hash()), true)
	if args.Index >= int64(len(block.Uncles)) || args.Index < 0 {
		return nil, nil
	} else {
		return block.Uncles[args.Index], nil
	}
}

func (self *ethApi) GetCompilers(req *shared.Request) (interface{}, error) {
	var lang string
	if solc, _ := self.xeth.Solc(); solc != nil {
		lang = "Solidity"
	}
	c := []string{lang}
	return c, nil
}

func (self *ethApi) CompileSolidity(req *shared.Request) (interface{}, error) {
	solc, _ := self.xeth.Solc()
	if solc == nil {
		return nil, shared.NewNotAvailableError(req.Method, "solc (solidity compiler) not found")
	}

	args := new(SourceArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	contracts, err := solc.Compile(args.Source)
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

func (self *ethApi) NewFilter(req *shared.Request) (interface{}, error) {
	args := new(BlockFilterArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	id := self.xeth.NewLogFilter(args.Earliest, args.Latest, args.Skip, args.Max, args.Address, args.Topics)
	return newHexNum(big.NewInt(int64(id)).Bytes()), nil
}

func (self *ethApi) NewBlockFilter(req *shared.Request) (interface{}, error) {
	return newHexNum(self.xeth.NewBlockFilter()), nil
}

func (self *ethApi) NewPendingTransactionFilter(req *shared.Request) (interface{}, error) {
	return newHexNum(self.xeth.NewTransactionFilter()), nil
}

func (self *ethApi) UninstallFilter(req *shared.Request) (interface{}, error) {
	args := new(FilterIdArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	return self.xeth.UninstallFilter(args.Id), nil
}

func (self *ethApi) GetFilterChanges(req *shared.Request) (interface{}, error) {
	args := new(FilterIdArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	switch self.xeth.GetFilterType(args.Id) {
	case xeth.BlockFilterTy:
		return NewHashesRes(self.xeth.BlockFilterChanged(args.Id)), nil
	case xeth.TransactionFilterTy:
		return NewHashesRes(self.xeth.TransactionFilterChanged(args.Id)), nil
	case xeth.LogFilterTy:
		return NewLogsRes(self.xeth.LogFilterChanged(args.Id)), nil
	default:
		return []string{}, nil // reply empty string slice
	}
}

func (self *ethApi) GetFilterLogs(req *shared.Request) (interface{}, error) {
	args := new(FilterIdArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	return NewLogsRes(self.xeth.Logs(args.Id)), nil
}

func (self *ethApi) GetLogs(req *shared.Request) (interface{}, error) {
	args := new(BlockFilterArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	return NewLogsRes(self.xeth.AllLogs(args.Earliest, args.Latest, args.Skip, args.Max, args.Address, args.Topics)), nil
}

func (self *ethApi) GetWork(req *shared.Request) (interface{}, error) {
	self.xeth.SetMining(true, 0)
	ret, err := self.xeth.RemoteMining().GetWork()
	if err != nil {
		return nil, shared.NewNotReadyError("mining work")
	} else {
		return ret, nil
	}
}

func (self *ethApi) SubmitWork(req *shared.Request) (interface{}, error) {
	args := new(SubmitWorkArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	return self.xeth.RemoteMining().SubmitWork(args.Nonce, common.HexToHash(args.Digest), common.HexToHash(args.Header)), nil
}

func (self *ethApi) SubmitHashrate(req *shared.Request) (interface{}, error) {
	args := new(SubmitHashRateArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return false, shared.NewDecodeParamError(err.Error())
	}
	self.xeth.RemoteMining().SubmitHashrate(common.HexToHash(args.Id), args.Rate)
	return true, nil
}

func (self *ethApi) Resend(req *shared.Request) (interface{}, error) {
	args := new(ResendArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	from := common.HexToAddress(args.Tx.From)

	pending := self.shift.TxPool().GetTransactions()
	for _, p := range pending {
		if pFrom, err := p.FromFrontier(); err == nil && pFrom == from && p.SigHash() == args.Tx.tx.SigHash() {
			self.shift.TxPool().RemoveTx(common.HexToHash(args.Tx.Hash))
			return self.xeth.Transact(args.Tx.From, args.Tx.To, args.Tx.Nonce, args.Tx.Value, args.NrgLimit, args.NrgPrice, args.Tx.Data)
		}
	}

	return nil, fmt.Errorf("Transaction %s not found", args.Tx.Hash)
}

func (self *ethApi) PendingTransactions(req *shared.Request) (interface{}, error) {
	txs := self.shift.TxPool().GetTransactions()

	// grab the accounts from the account manager. This will help with determining which
	// transactions should be returned.
	accounts, err := self.shift.AccountManager().Accounts()
	if err != nil {
		return nil, err
	}

	// Add the accouns to a new set
	accountSet := set.New()
	for _, account := range accounts {
		accountSet.Add(account.Address)
	}

	var ltxs []*tx
	for _, tx := range txs {
		if from, _ := tx.FromFrontier(); accountSet.Has(from) {
			ltxs = append(ltxs, newTx(tx))
		}
	}

	return ltxs, nil
}

func (self *ethApi) GetTransactionReceipt(req *shared.Request) (interface{}, error) {
	args := new(HashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	txhash := common.BytesToHash(common.FromHex(args.Hash))
	tx, bhash, bnum, txi := self.xeth.EthTransactionByHash(args.Hash)
	rec := self.xeth.GetTxReceipt(txhash)
	// We could have an error of "not found". Should disambiguate
	// if err != nil {
	// 	return err, nil
	// }
	if rec != nil && tx != nil {
		v := NewReceiptRes(rec)
		v.BlockHash = newHexData(bhash)
		v.BlockNumber = newHexNum(bnum)
		v.TransactionIndex = newHexNum(txi)
		return v, nil
	}

	return nil, nil
}
