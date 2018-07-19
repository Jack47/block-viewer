package block

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// TODO: add more test cases
var testBlocks = []struct {
	blockHash string
	expected  Block
}{
	{
		blockHash: "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
		expected: Block{
			Header: &Header{
				Version:             1,
				PrevBlockHeaderHash: "0000000000000000000000000000000000000000000000000000000000000000",
				MerkleRootHash:      "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
				Time:                time.Unix(1231006505, 0),
				NBits:               486604799,
				Nonce:               2083236893,
			},
			TxCount: 1,
			Txs: []*Tx{
				&Tx{
					Size:     204,
					Version:  1,
					InCount:  1,
					OutCount: 1,
					Ins: []*TxIn{
						&TxIn{
							PrevOutput: &OutPoint{
								Hash:  "0000000000000000000000000000000000000000000000000000000000000000",
								Index: 4294967295,
							},
							ScriptBytes:     77,
							SignatureScript: "04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73",
							Sequence:        0,
						},
					},
					Outs: []*TxOut{
						&TxOut{
							Value:             5000000000,
							PubKeyScriptBytes: 67,
							PubKeyScript:      "4104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac",
						},
					},
					LockTime: 0,
				},
			},
		},
	},
}

func TestFetchAndParseBlocks(b *testing.T) {
	for _, testBlock := range testBlocks {
		bytes, err := fetchBlock(testBlock.blockHash, true /*dumpFile*/)
		if err != nil {
			b.Fatal(err)
		}
		block, err := NewBlock(bytes)
		if err != nil {
			b.Fatal(fmt.Sprintf("parse genesis block failed: %s", err))
		}
		expected, _ := json.MarshalIndent(testBlock.expected, "" /*prefix*/, " " /*indent*/)
		real, err := json.MarshalIndent(block, "" /*prefix*/, " " /*indent*/)
		if err != nil {
			b.Fatal(fmt.Sprintf("marshal parsed block failed: %s", err))
		}
		if string(expected) != string(real) {
			b.Fatal(fmt.Sprintf("expected block: \n%s\n, but got: \n%s\n", expected, real))
		}
	}
}
