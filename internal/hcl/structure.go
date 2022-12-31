package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var (
	tf_schema = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "terraform",
			},
			{
				Type:       "provider",
				LabelNames: []string{"name"},
			},
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
			{
				Type: "locals",
			},
			{
				Type:       "output",
				LabelNames: []string{"name"},
			},
			{
				Type:       "module",
				LabelNames: []string{"name"},
			},
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
			},
			{
				Type:       "data",
				LabelNames: []string{"type", "name"},
			},
		},
	}
	tf_provider_schema = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "provider",
				LabelNames: []string{"name"},
			},
		},
	}
	tf_module_schema = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "module",
				LabelNames: []string{"name"},
			},
		},
	}
)

// A block is a representation of a hcl.Block enriched with informations which allow an easy evaluation of the block itself.
type Block struct {
	// each block has a parent block, representing the module the block belongs to
	parentBlock *Block
	// each block may have a number of blocks inside itself
	childBlocks Blocks
	// the original hcl.Block this Block was created from
	originalBlock *hcl.Block
}

func NewBlock(file string, hclBlock *hcl.Block, parentBlock *Block) *Block {
	var children Blocks

	if body, ok := hclBlock.Body.(*hclsyntax.Body); ok {
		for _, bodyblock := range body.Blocks {
			children = append(children, NewBlock(file, bodyblock.AsHCLBlock(), parentBlock))
		}

		return &Block{
			parentBlock:   parentBlock,
			childBlocks:   children,
			originalBlock: hclBlock,
		}
	}

	content, _, diag := hclBlock.Body.PartialContent(tf_schema)
	if diag != nil && diag.HasErrors() {
		block := &Block{
			parentBlock:   parentBlock,
			childBlocks:   children,
			originalBlock: hclBlock,
		}

		return block
	}

	for _, hb := range content.Blocks {
		children = append(children, NewBlock(file, hb, parentBlock))
	}

	block := &Block{
		parentBlock:   parentBlock,
		childBlocks:   children,
		originalBlock: hclBlock,
	}

	return block

}

func (block Block) Type() string {
	return block.originalBlock.Type
}

// Helper to represent a block slice
type Blocks []*Block

func (blocks Blocks) WithType(requestedType string) Blocks {
	var blocksWithRequestedType Blocks

	for _, block := range blocks {
		if block.Type() == requestedType {
			blocksWithRequestedType = append(blocksWithRequestedType, block)
		}
	}

	return blocksWithRequestedType
}

func loadBlocksFromFile(file parsedFile) (hcl.Blocks, error) {
	contents, diags := file.hclFile.Body.Content(tf_schema)
	if diags != nil && diags.HasErrors() {
		return nil, diags
	}

	if contents == nil {
		return nil, fmt.Errorf("empty file, can not load blocks")
	}

	return contents.Blocks, nil
}
