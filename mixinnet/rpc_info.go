package mixinnet

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	Mint struct {
		Pool   decimal.Decimal `json:"pool"`
		Pledge decimal.Decimal `json:"pledge"`
		Batch  uint64          `json:"batch"`
	}

	Queue struct {
		Finals uint64 `json:"finals"`
		Caches uint64 `json:"caches"`
	}

	ConsensusNode struct {
		Node        Hash      `json:"node"`
		Signer      Address   `json:"signer"`
		Payee       Address   `json:"payee"`
		State       string    `json:"state"`
		Timestamp   int64     `json:"timestamp"`
		Transaction Hash      `json:"transaction"`
		Aggregator  uint64    `json:"aggregator"`
		Works       [2]uint64 `json:"works"`
	}

	GraphReferences struct {
		External Hash `json:"external"`
		Self     Hash `json:"self"`
	}

	GraphSnapshot struct {
		Node        Hash            `json:"node"`
		Hash        Hash            `json:"hash"`
		References  GraphReferences `json:"references"`
		Round       uint64          `json:"round"`
		Timestamp   int64           `json:"timestamp"`
		Transaction Hash            `json:"transaction"`
		Signature   string          `json:"signature"` // CosiSignature
		Version     int             `json:"version"`
	}

	GraphCache struct {
		Node       Hash             `json:"node"`
		References GraphReferences  `json:"references"`
		Timestamp  int64            `json:"timestamp"`
		Round      uint64           `json:"round"`
		Snapshots  []*GraphSnapshot `json:"snapshots"`
	}

	GraphFinal struct {
		Node  Hash   `json:"node"`
		Hash  Hash   `json:"hash"`
		Start int64  `json:"start"`
		End   int64  `json:"end"`
		Round uint64 `json:"round"`
	}

	Graph struct {
		SPS       float64                `json:"sps"`
		Topology  uint64                 `json:"topology"`
		Consensus []*ConsensusNode       `json:"consensus"`
		Final     map[string]*GraphFinal `json:"final"`
		Cache     map[string]*GraphCache `json:"cache"`
	}

	ConsensusInfo struct {
		Network   Hash      `json:"network"`
		Node      Hash      `json:"node"`
		Version   string    `json:"version"`
		Uptime    string    `json:"uptime"`
		Epoch     time.Time `json:"epoch"`
		Timestamp time.Time `json:"timestamp"`
		Mint      Mint      `json:"mint"`
		Queue     Queue     `json:"queue"`
		Graph     Graph     `json:"graph"`
	}
)
