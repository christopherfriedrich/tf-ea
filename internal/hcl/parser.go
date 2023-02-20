package hcl

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/christopherfriedrich/tf-ea/internal/log"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

func ParseDir(dirPath string) {
	log.Logger.Sugar().Info("Starting to parse input dir ...")

	files, err := translateDirectory(dirPath)
	if err != nil {
		log.Logger.Sugar().Error(err)
	}
	// https://github.com/infracost/infracost/blob/master/internal/hcl/parser.go#L303
	blocks, err := translateDirectoryFiles(files)
	if err != nil {
		log.Logger.Sugar().Error(err)
	}

	// load vars from tfvars file
	// https://github.com/infracost/infracost/blob/master/internal/hcl/parser.go#L313
	// introduce variable tfvarspath for 2nd arg
	inputVars, err := loadVars(blocks, []string{})
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
			log.Logger.Sugar().Warnf("skipping file could not load blocks err: %s", err)
			continue
		}

		if len(fileBlocks) > 0 {
			log.Logger.Sugar().Debugf("Added %d blocks from %s...", len(fileBlocks), fileBlocks[0].DefRange.Filename)
		}

		for _, fileBlock := range fileBlocks {
			parsedBlocks = append(
				parsedBlocks,
				NewBlock(file.path, fileBlock, nil),
			)
		}
	}

	return parsedBlocks, nil
}

func loadVars(blocks Blocks, filenames []string) (map[string]cty.Value, error) {
	// TODO: read vars from environment
	// combinedVars := p.tfEnvVars
	combinedVars := make(map[string]cty.Value)
	inputVars := make(map[string]cty.Value)
	if combinedVars == nil {
		combinedVars = make(map[string]cty.Value)
	}

	// handle variables from Terraform Cloud

	for _, name := range []string{} {
		err := loadAndCombineVars(name, combinedVars)
		if err != nil {
			log.Logger.Sugar().Errorf("could not load vars from auto var file %s", name)
			continue
		}
	}

	for _, filename := range filenames {
		err := loadAndCombineVars(filename, combinedVars)
		if err != nil {
			return combinedVars, err
		}
	}

	for k, v := range inputVars {
		combinedVars[k] = v
	}

	return combinedVars, nil
}

func loadAndCombineVars(filename string, combinedVars map[string]cty.Value) error {
	vars, err := loadVarFile(filename)
	if err != nil {
		return err
	}

	for k, v := range vars {
		combinedVars[k] = v
	}

	return nil
}

func loadVarFile(filename string) (map[string]cty.Value, error) {
	inputVars := make(map[string]cty.Value)

	if filename == "" {
		return inputVars, nil
	}

	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("Passed var file does not exist: %s. Make sure you are passing the var file path relative to the --path flag.", filename)
		}

		return nil, fmt.Errorf("could not stat provided var file: %w", err)
	}

	var parseFunc func(filename string) (*hcl.File, hcl.Diagnostics)

	hclParser := hclparse.NewParser()

	parseFunc = hclParser.ParseHCLFile
	if strings.HasSuffix(filename, ".json") {
		parseFunc = hclParser.ParseJSONFile
	}

	variableFile, diags := parseFunc(filename)
	if diags.HasErrors() {
		log.Logger.Sugar().Debugf("could not parse supplied var file %s", filename)

		return inputVars, nil
	}

	attrs, _ := variableFile.Body.JustAttributes()

	for _, attr := range attrs {
		value, diag := attr.Expr.Value(&hcl.EvalContext{})
		if diag.HasErrors() {
			log.Logger.Sugar().Debugf("problem evaluating input var %s", attr.Name)
		}

		inputVars[attr.Name] = value
	}

	return inputVars, nil
}

type Context struct {
}
