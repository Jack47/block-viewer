package block

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Header struct {
	Version             int32
	PrevBlockHeaderHash string
	prevBlockHeaderHash []byte // [32]

	MerkleRootHash string
	merkleRootHash []byte // [32]
	// The block time is a Unix epoch time when the miner started hashing the header (according to the miner).
	time uint32
	Time time.Time
	// Difficulty
	NBits uint32
	Nonce uint32
}

var ToUint16 = binary.LittleEndian.Uint16
var ToUint32 = binary.LittleEndian.Uint32
var ToUint64 = binary.LittleEndian.Uint64

func toBigEndianBytes(data []byte) []byte {
	bytes := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		bytes[i] = data[len(data)-1-i]
	}
	return bytes
}
func compactSizeUint(data []byte) (v uint64, bytesUsed int) {
	switch data[0] {
	case 0xfd:
		v = uint64(ToUint16(data[1:]))
		bytesUsed = 3
	case 0xfe:
		v = uint64(ToUint32(data[1:]))
		bytesUsed = 5
	case 0xff:
		v = ToUint64(data[1:])
		bytesUsed = 9
	default:
		v = uint64(data[0])
		bytesUsed = 1
	}
	return
}

func NewHeader(data []byte) (h *Header, err error) {
	// The hashes are in internal byte order;
	// the other values are all in little-endian order.
	// But what does internal byte order means?
	h = new(Header)
	h.Version = int32(ToUint32(data[0:4]))
	h.prevBlockHeaderHash = data[4:36]
	h.PrevBlockHeaderHash = hex.EncodeToString(toBigEndianBytes(h.prevBlockHeaderHash))
	h.merkleRootHash = data[36:68]
	h.MerkleRootHash = hex.EncodeToString(toBigEndianBytes(h.merkleRootHash))
	h.time = ToUint32(data[68:72])
	h.Time = time.Unix(int64(h.time), 0)
	h.NBits = ToUint32(data[72:76])
	h.Nonce = ToUint32(data[76:80])
	return
}

type Block struct {
	Header *Header
	// The total number of transactions in this block, including the coinbase transaction.
	TxCount uint64
	Txs     []*Tx
}

func NewBlock(data []byte) (bl *Block, err error) {
	if data == nil {
		err = errors.New("nil pointer")
		return
	}
	if len(data) < 80 {
		err = errors.New("len shorter than 80")
		return
	}
	h, err := NewHeader(data[:80])
	bl = new(Block)
	if err != nil {
		err = errors.New(fmt.Sprintf("parse header failed: %s", err))
		return
	}
	bl.Header = h
	body := data[80:]

	offset := 0
	bytesUsed := 0
	bl.TxCount, bytesUsed = compactSizeUint(body)
	offset += bytesUsed
	bl.Txs = make([]*Tx, bl.TxCount)

	// coinbase transaction
	// Always created by a miner, it includes a single coinbase.
	cbTx := new(Tx)
	bytesUsed, err = cbTx.parse(body[offset:], true /*isCoinBase*/)
	if err != nil {
		return nil, err
	}
	bl.Txs[0] = cbTx
	offset += bytesUsed
	print(cbTx)

	for i := uint64(1); i < bl.TxCount; i++ {
		tx := new(Tx)
		// fmt.Printf("parse %dth\n", i)
		bytesUsed, err = tx.parse(body[offset:], false /*isCoinBase*/)
		if err != nil {
			return nil, err
		}
		print(tx)
		fmt.Printf("\n")
		offset += bytesUsed
		bl.Txs[i] = tx
	}
	return
}

type Tx struct {
	Size    int
	Version uint32
	InCount uint64
	// must have exactly one input, called a coinbase.
	Ins      []*TxIn
	OutCount uint64
	Outs     []*TxOut
	// A time (Unix epoch time) or block number
	LockTime uint32
}

func (tx *Tx) parse(data []byte, isCoinBase bool) (bytesUsed int, err error) {
	// TODO check length
	offset := 0
	tx.Version = ToUint32(data[0:])
	offset += 4
	tx.InCount, bytesUsed = compactSizeUint(data[offset:])
	offset += bytesUsed
	// parse tx.Ins
	tx.Ins = make([]*TxIn, tx.InCount)
	if isCoinBase && tx.InCount != 1 {
		return offset, errors.New(fmt.Sprintf("coinbase transaction must have execatly 1 input transaction, called a coinbase, but got: %d", tx.InCount))
	}

	for i := uint64(0); i < tx.InCount; i++ {
		in := new(TxIn)
		bytesUsed, err = in.parse(data[offset:], isCoinBase)
		if err != nil {
			return offset, err
		}
		tx.Ins[i] = in
		offset += bytesUsed
	}
	tx.OutCount, bytesUsed = compactSizeUint(data[offset:])
	offset += bytesUsed
	// parse tx.outs
	tx.Outs = make([]*TxOut, tx.OutCount)
	for i := uint64(0); i < tx.OutCount; i++ {
		out := new(TxOut)
		bytesUsed, err = out.parse(data[offset:])
		if err != nil {
			return offset, err
		}
		tx.Outs[i] = out
		offset += bytesUsed
	}

	tx.LockTime = ToUint32(data[offset:])
	offset += 4
	tx.Size = offset
	return offset, nil
}

type TxIn struct {
	PrevOutput *OutPoint // 36 bytes The previous outpoint being spent.
	// The number of bytes in the signature script. Maximum is 10,000 bytes.
	ScriptBytes     uint64
	signatureScript []byte
	SignatureScript string
	// Sequence number. Default for Bitcoin Core and almost all other programs is 0xffffffff.
	Sequence uint32
}

func (txIn *TxIn) parse(data []byte, isCoinBase bool) (bytesUsed int, err error) {
	minLen := 36 + 1 + 1 + 4 /* minimum len */
	if len(data) < minLen {
		return 0, errors.New(fmt.Sprintf("data len must >= %d, got %d", minLen, len(data)))
	}
	offset := 0
	txIn.PrevOutput = new(OutPoint)
	if bytesUsed, err = txIn.PrevOutput.parse(data); err != nil {
		return offset, err
	}
	offset += bytesUsed
	txIn.ScriptBytes, bytesUsed = compactSizeUint(data[offset:])
	offset += bytesUsed
	// TODO is cast to int acceptable ?
	txIn.signatureScript = data[offset : offset+int(txIn.ScriptBytes)]
	txIn.SignatureScript = hex.EncodeToString(txIn.signatureScript)
	offset += int(txIn.ScriptBytes)
	txIn.Sequence = ToUint32(data)
	offset += 4
	return offset, nil
}

// Each non-coinbase input spends an outpoint from a previous transaction.
type OutPoint struct {
	// The TXID of the transaction holding the output to spend
	hash  []byte // 32 bytes
	Hash  string
	Index uint32
}

func (o *OutPoint) parse(data []byte) (bytesUsed int, err error) {
	o.hash = data[0:32]
	o.Hash = hex.EncodeToString(toBigEndianBytes(o.hash))
	o.Index = ToUint32(data[32:])
	return 32 + 4, nil
}

type TxOut struct {
	// Number of satoshis to spend.
	Value             uint64
	PubKeyScriptBytes uint64
	// Indicate what conditions must be fulfilled for those
	// satoshis to be further spent
	pubKeyScript []byte
	PubKeyScript string
}

func (t *TxOut) parse(data []byte) (bytesUsed int, err error) {
	offset := 0
	bytesUsed = 0
	t.Value = ToUint64(data)
	offset += 8
	t.PubKeyScriptBytes, bytesUsed = compactSizeUint(data[offset:])
	offset += bytesUsed
	t.pubKeyScript = data[offset : offset+int(t.PubKeyScriptBytes)]
	t.PubKeyScript = hex.EncodeToString(t.pubKeyScript)
	offset += int(t.PubKeyScriptBytes)
	return offset, nil
}

// fetch block from internet, dump to temp dir if dumpFile is true
func FetchBlock(blockHash string, dumpFile bool) ([]byte, error) {
	// url := "https://blockchain.info/block/" + blockHash + "?format=hex"
	url := "https://webbtc.com/block/" + blockHash + ".bin"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch block failed: %s", err)
	}
	if resp.StatusCode == 200 {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read body failed: %s", err)
		}
		if dumpFile {
			ioutil.WriteFile(filepath.Join(os.TempDir(), blockHash), bytes, 0400)
		}
		return bytes, nil
	}
	return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
}
