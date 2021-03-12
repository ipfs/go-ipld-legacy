package ipldlegacy

import (
	blocks "github.com/ipfs/go-block-format"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

// UniversalNode satisfies both go-ipld-prime interfaces and legacy interfaces
type UniversalNode interface {
	ipld.Node
	format.Node
}

// NodeConverter converts a go-ipld-prime node + block combination to a UniversalNode that satisfies both current and legacy ipld formats
type NodeConverter func(b blocks.Block, nd ipld.Node) (UniversalNode, error)

type codecConverter struct {
	prototype ipld.NodePrototype
	converter NodeConverter
}

var codecTable = map[uint64]codecConverter{}

// RegisterCodec registers a specialized prototype & converter for a specific codec
func RegisterCodec(codec uint64, prototype ipld.NodePrototype, converter NodeConverter) {
	codecTable[codec] = codecConverter{prototype, converter}
}

// DecodeNode builds a UniversalNode from a block
func DecodeNode(lsys ipld.LinkSystem, b blocks.Block) (UniversalNode, error) {
	c := b.Cid()
	link := cidlink.Link{Cid: c}
	var prototype ipld.NodePrototype = basicnode.Prototype.Any
	converter, hasConverter := codecTable[c.Prefix().Codec]
	if hasConverter {
		prototype = converter.prototype
	}
	nd, err := lsys.Load(ipld.LinkContext{}, link, prototype)
	if err != nil {
		return nil, err
	}

	if hasConverter {
		return converter.converter(b, nd)
	}
	return &LegacyNode{b, nd}, nil
}
