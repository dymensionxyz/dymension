package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Global map to translate external Protobuf package names to the corresponding Cosmos SDK module names.
// These are used for tagging the OpenAPI specification.
var packageToModule = map[string]string{
	"cosmos.auth.v1beta1":                              "auth",
	"cosmos.base.node.v1beta1":                         "node",
	"cosmos.base.tendermint.v1beta1":                   "tendermint",
	"ibc.applications.transfer.v1":                     "transfer",
	"ibc.applications.interchain_accounts.host.v1":     "host",
	"ethermint.evm.v1":                                 "evm",
	"cosmos.authz.v1beta1":                             "authz",
	"cosmos.bank.v1beta1":                              "bank",
	"cosmos.consensus.v1":                              "consensus",
	"cosmos.distribution.v1beta1":                      "distribution",
	"cosmos.evidence.v1beta1":                          "evidence",
	"cosmos.feegrant.v1beta1":                          "feegrant",
	"cosmos.gov.v1beta1":                               "gov",
	"cosmos.gov.v1":                                    "gov",
	"cosmos.group.v1":                                  "group",
	"cosmos.mint.v1beta1":                              "mint",
	"cosmos.nft.v1beta1":                               "nft",
	"cosmos.params.v1beta1":                            "params",
	"cosmos.slashing.v1beta1":                          "slashing",
	"cosmos.staking.v1beta1":                           "staking",
	"cosmos.upgrade.v1beta1":                           "upgrade",
	"ethermint.feemarket.v1":                           "feemarket",
	"ibc.applications.fee.v1":                          "fee",
	"ibc.applications.interchain_accounts.controller.v1": "controller",
	"ibc.core.channel.v1":                              "channel",
	"ibc.core.client.v1":                               "client",
	"ibc.core.connection.v1":                           "connection",
	"cosmos.autocli.v1":                                "autocli",
	"cosmos.app.v1alpha1":                              "app",
	"cosmos.tx.v1beta1":                                "tx",
}

// Regex definitions are compiled once globally for performance.
var (
	// Regex to capture the Protobuf package name: `package X;`
	pkgRegex = regexp.MustCompile(`^\s*package\s+([\w\.]+)\s*;\s*$`)

	// Detects the start of an `rpc` line for processing.
	startRpcRegex = regexp.MustCompile(`(?i)^\s*rpc\s+\S`)

	// Detects an empty RPC block: `{} or {};`
	emptyBlockRegex = regexp.MustCompile(`\{\s*\}\s*;?\s*$`)

	// Detects if the OpenAPI annotation is already present in a block.
	hasOpenapiv2Regex = regexp.MustCompile(`openapiv2_operation`)
)

// The required import statement for OpenAPI v2 options.
const importLine = `import "protoc-gen-openapiv2/options/annotations.proto";`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <proto root dir>\n", os.Args[0])
		os.Exit(1)
	}
	root := os.Args[1]

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		
		// Only process files defining query or service definitions.
		if base == "query.proto" || base == "service.proto" {
			fmt.Printf("Processing file: %s\n", path)
			e := fixFile(path)
			if e != nil {
				return fmt.Errorf("fixFile error in %s: %w", path, e)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// fixFile reads, processes, and overwrites the target Protobuf file.
func fixFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	pkgName := findPackageName(lines)

	// Skip specific external packages that do not require this annotation logic.
	if pkgName == "cosmos.orm.query.v1alpha1" {
		return nil
	}

	moduleName := determineModuleName(pkgName)

	// Ensure the required OpenAPI options import is present.
	lines = ensureImport(lines)

	// Inject the openapiv2 tags into RPC definitions.
	lines = fixRPC(lines, moduleName)

	final := strings.Join(lines, "\n")
	// Note: Using 0o644 for standard read/write permissions.
	return os.WriteFile(filePath, []byte(final), 0o644)
}

// findPackageName extracts the Protobuf package name from the file lines.
func findPackageName(lines []string) string {
	for _, ln := range lines {
		trim := strings.TrimSpace(ln)
		match := pkgRegex.FindStringSubmatch(trim)
		if match != nil {
			return match[1]
		}
	}
	return ""
}

// determineModuleName translates the Protobuf package name into the corresponding module name (tag).
func determineModuleName(pkgName string) string {
	if pkgName == "" {
		return "unknown"
	}
	const prefix = "dymensionxyz.dymension."
	
	// Handle local Dymension packages.
	if strings.HasPrefix(pkgName, prefix) {
		parts := strings.Split(pkgName, ".")
		// The module name is expected to be the third part (e.g., dymensionxyz.dymension.rollapp -> rollapp).
		if len(parts) >= 3 {
			return parts[2]
		}
		return "unknown"
	}
	
	// Handle well-known external Cosmos SDK packages.
	if mod, ok := packageToModule[pkgName]; ok {
		return mod
	}
	return "unknown"
}

// ensureImport guarantees the OpenAPI options import is present, adding it after the last 'import' or 'package' line.
func ensureImport(lines []string) []string {
	hasImport := false
	lastImport := -1
	lastPackage := -1
	for i, ln := range lines {
		if strings.Contains(ln, importLine) {
			hasImport = true
		}
		trim := strings.TrimSpace(ln)
		if strings.HasPrefix(trim, "import ") {
			lastImport = i
		} else if strings.HasPrefix(trim, "package ") {
			lastPackage = i
		}
	}
	
	if !hasImport {
		// Determine injection point: after the last import, or after the package definition.
		insertPos := 0
		if lastImport != -1 {
			insertPos = lastImport + 1
		} else if lastPackage != -1 {
			insertPos = lastPackage + 1
		}
		
		// In-place insertion idiom: append a new element and shift the slice.
		lines = append(lines, "")
		copy(lines[insertPos+1:], lines[insertPos:])
		lines[insertPos] = importLine
	}
	return lines
}

// fixRPC iterates through lines and collects RPC blocks to inject the OpenAPI annotation.
func fixRPC(in []string, moduleName string) []string {
	out := []string{}
	i := 0

	for i < len(in) {
		ln := in[i]
		trim := strings.TrimSpace(ln)

		if startRpcRegex.MatchString(trim) {
			blockLines := []string{ln}

			// Check if RPC is defined on a single line with empty braces (e.g., `rpc ... {}`)
			isEmptyBlock := emptyBlockRegex.MatchString(trim)
			
			bracketDepth := 0
			if strings.Contains(trim, "{") {
				bracketDepth++
			}
			i++
			done := false

			// If it's a single-line empty block, process immediately.
			if isEmptyBlock {
				if !alreadyHasOpenapiv2(blockLines) {
					// Use a safer function to inject annotation into an empty block.
					blockLines = injectOpenapiv2IntoEmptyBlock(blockLines, moduleName)
				}
				out = append(out, blockLines...)
				continue
			}
			
			// Process multi-line RPC block until the closing brace is found.
			for i < len(in) && !done {
				nextLine := in[i]
				nextTrim := strings.TrimSpace(nextLine)

				bracketDelta := countBracketChanges(nextTrim)
				oldDepth := bracketDepth
				bracketDepth += bracketDelta

				blockLines = append(blockLines, nextLine)
				i++

				if oldDepth > 0 && bracketDepth == 0 {
					done = true
				}
			}
			
			// Inject annotation if not already present, ensuring the block was fully captured (bracketDepth == 0).
			if !alreadyHasOpenapiv2(blockLines) && bracketDepth == 0 {
				blockLines = injectOpenapiv2(blockLines, moduleName)
			}
			out = append(out, blockLines...)
			continue
		}

		out = append(out, ln)
		i++
	}
	return out
}

// countBracketChanges calculates the net change in brace depth on a single line.
func countBracketChanges(line string) int {
	count := 0
	for _, r := range line {
		if r == '{' {
			count++
		} else if r == '}' {
			count--
		}
	}
	return count
}

// alreadyHasOpenapiv2 checks if the annotation is present in the block.
func alreadyHasOpenapiv2(blockLines []string) bool {
	for _, ln := range blockLines {
		if hasOpenapiv2Regex.MatchString(ln) {
			return true
		}
	}
	return false
}

// injectOpenapiv2IntoEmptyBlock safely replaces a single-line empty RPC block with the annotated version.
func injectOpenapiv2IntoEmptyBlock(blockLines []string, moduleName string) []string {
	if len(blockLines) == 0 {
		return blockLines
	}
	rpcLine := blockLines[0]
	
	// Remove trailing braces (with optional semicolon) to get the clean RPC definition start.
	rpcLine = emptyBlockRegex.ReplaceAllStringFunc(rpcLine, func(match string) string {
		return strings.TrimRight(match, " \t{};") // Remove braces and spaces
	})
	
	// Add the annotation and re-add braces.
	annotation := fmt.Sprintf(" {\n		option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = { tags: [\"%s\"] };\n}", moduleName)
	return []string{rpcLine + annotation}
}


// injectOpenapiv2 inserts the annotation right before the closing brace of a multi-line RPC block.
func injectOpenapiv2(blockLines []string, moduleName string) []string {
	lastIdx := len(blockLines) - 1
	if lastIdx < 0 {
		return blockLines
	}

	// Find the line containing the closing brace '}' from the end backwards.
	endIdx := -1
	for i := lastIdx; i >= 0; i-- {
		trim := strings.TrimSpace(blockLines[i])
		if strings.HasSuffix(trim, "}") || strings.HasSuffix(trim, "};") {
			endIdx = i
			break
		}
	}
	if endIdx < 0 {
		// Could not find the closing brace, return the original block.
		return blockLines
	}

	annotation := fmt.Sprintf("		option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = { tags: [\"%s\"] };", moduleName)

	// Insert the annotation just before the line containing the closing brace.
	newArr := make([]string, 0, len(blockLines)+1)
	newArr = append(newArr, blockLines[:endIdx]...)
	newArr = append(newArr, annotation)
	newArr = append(newArr, blockLines[endIdx:]...)
	return newArr
}
