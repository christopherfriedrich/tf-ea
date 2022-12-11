package hcl

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/christopherfriedrich/tf-ea/internal/log"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func ParseDir(dirPath string) {
	log.Logger.Sugar().Info("Starting to parse input dir ...")

	files, err := translateDirectory(dirPath)
	if err != nil {
		log.Logger.Sugar().Error(err)
	}
}

// wrapper for a parsed hcl.File including its path
type parsedFile struct {
	path    string
	hclFile *hcl.File
}

type parsedFiles []parsedFile

// Takes a path to a directory whos files are then transformed into hcl.File and wrapped into parsedFile
func translateDirectory(dirPath string) (parsedFiles, error) {

	parser := hclparse.NewParser()

	fileInfos, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}

		var parseFunc func(filename string) (*hcl.File, hcl.Diagnostics)
		if strings.HasSuffix(info.Name(), ".tf") {
			parseFunc = parser.ParseHCLFile
		}

		if strings.HasSuffix(info.Name(), ".tf.json") {
			parseFunc = parser.ParseJSONFile
		}

		// this is not a file we can parse:
		if parseFunc == nil {
			continue
		}

		path := filepath.Join(dirPath, info.Name())
		_, diag := parseFunc(path)
		if diag != nil && diag.HasErrors() {
			log.Logger.Sugar().Warnf("skipping file: %s hcl parsing err: %s", path, diag.Error())

			continue
		}
	}

	files := make(parsedFiles, 0, len(parser.Files()))
	for filename, f := range parser.Files() {
		files = append(files, parsedFile{hclFile: f, path: filename})
	}

	return files, nil
}

func translateDirectoryFiles(parsedDirectoryFiles parsedFiles) (Blocks, error) {
	var parsedBlocks Blocks

	for _, file := range parsedDirectoryFiles {
		fileBlocks, err := loadBlocksFromFile(file, nil)
		if err != nil {
			if p.stopOnHCLError {
				return nil, err
			}

			log.Logger.Sugar().Warnf("skipping file could not load blocks err: %s", err)
			continue
		}

		if len(fileBlocks) > 0 {
			log.Logger.Sugar().Debugf("Added %d blocks from %s...", len(fileBlocks), fileBlocks[0].DefRange.Filename)
		}

		for _, fileBlock := range fileBlocks {
			parsedBlocks = append(
				parsedBlocks,
				NewBlock(file.path, fileBlock, nil, nil),
			)
		}
	}

	return parsedBlocks, nil
}

type Context struct {
}

func NewBlock(file string, originalBlock *hcl.Block, ctx *Context, module *Block) *Block {
	if ctx == nil {
		ctx = NewContext()
	}

	var children Blocks
	if body, ok := originalBlock.Body.(*hclsyntax.Body); ok {

	}
}
