package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// For packages that do NOT start with "dymensionxyz.dymension."
var packageToModule = map[string]string{
	"cosmos.auth.v1beta1":                                "auth",
	"cosmos.base.node.v1beta1":                           "node",
	"cosmos.base.tendermint.v1beta1":                     "tendermint",
	"ibc.applications.transfer.v1":                       "transfer",
	"ibc.applications.interchain_accounts.host.v1":       "host",
	"ethermint.evm.v1":                                   "evm",
	"cosmos.authz.v1beta1":                               "authz",
	"cosmos.bank.v1beta1":                                "bank",
	"cosmos.consensus.v1":                                "consensus",
	"cosmos.distribution.v1beta1":                        "distribution",
	"cosmos.evidence.v1beta1":                            "evidence",
	"cosmos.feegrant.v1beta1":                            "feegrant",
	"cosmos.gov.v1beta1":                                 "gov",
	"cosmos.gov.v1":                                      "gov",
	"cosmos.group.v1":                                    "group",
	"cosmos.mint.v1beta1":                                "mint",
	"cosmos.nft.v1beta1":                                 "nft",
	"cosmos.params.v1beta1":                              "params",
	"cosmos.slashing.v1beta1":                            "slashing",
	"cosmos.staking.v1beta1":                             "staking",
	"cosmos.upgrade.v1beta1":                             "upgrade",
	"ethermint.feemarket.v1":                             "feemarket",
	"ibc.applications.fee.v1":                            "fee",
	"ibc.applications.interchain_accounts.controller.v1": "controller",
	"ibc.core.channel.v1":                                "channel",
	"ibc.core.client.v1":                                 "client",
	"ibc.core.connection.v1":                             "connection",
	"cosmos.autocli.v1":                                  "autocli",
	"cosmos.app.v1alpha1":                                "app",
	"cosmos.tx.v1beta1":                                  "tx",
}

// Regex to find `package X;`
var pkgRegex = regexp.MustCompile(`^\s*package\s+([\w\.]+)\s*;\s*$`)

// Start of an `rpc` line
var startRpcRegex = regexp.MustCompile(`(?i)^\s*rpc\s+\S`)

// Detect empty block lines: `{} or {};`
var emptyBlockRegex = regexp.MustCompile(`\{\s*\}\s*;?\s*$`)

// If a line has `openapiv2_operation`, we've already annotated
var hasOpenapiv2Regex = regexp.MustCompile(`openapiv2_operation`)

// The import line
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
		if base == "query.proto" || base == "service.proto" {
			fmt.Printf("Processing file: %s\n", path)
			e := fixFile(path)
			if e != nil {
				return fmt.Errorf("fixFile: %w", e)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func fixFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	pkgName := findPackageName(lines)

	// not needed
	if pkgName == "cosmos.orm.query.v1alpha1" {
		return nil
	}

	moduleName := determineModuleName(pkgName)

	// ensure the import line
	lines = ensureImport(lines)

	// fix RPC lines
	lines = fixRPC(lines, moduleName)

	final := strings.Join(lines, "\n")
	return os.WriteFile(filePath, []byte(final), 0o644)
}

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

func determineModuleName(pkgName string) string {
	if pkgName == "" {
		return "unknown"
	}
	const prefix = "dymensionxyz.dymension."
	if strings.HasPrefix(pkgName, prefix) {
		parts := strings.Split(pkgName, ".")
		if len(parts) >= 3 {
			return parts[2]
		}
		return "unknown"
	}
	if mod, ok := packageToModule[pkgName]; ok {
		return mod
	}
	return "unknown"
}

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
		insertPos := 0
		if lastImport != -1 {
			insertPos = lastImport + 1
		} else if lastPackage != -1 {
			insertPos = lastPackage + 1
		}
		lines = append(lines, "")
		copy(lines[insertPos+1:], lines[insertPos:])
		lines[insertPos] = importLine
	}
	return lines
}

func fixRPC(in []string, moduleName string) []string {
	out := []string{}
	i := 0

	for i < len(in) {
		ln := in[i]
		trim := strings.TrimSpace(ln)

		if startRpcRegex.MatchString(trim) {
			blockLines := []string{ln}

			isEmptyBlock := emptyBlockRegex.MatchString(trim) // e.g. ends in {}
			bracketDepth := 0
			if strings.Contains(trim, "{") {
				bracketDepth++
			}
			i++
			done := false

			if isEmptyBlock {
				alreadyHas := alreadyHasOpenapiv2(blockLines)
				if !alreadyHas {
					blockLines = injectOpenapiv2IfNeeded(blockLines, moduleName, true)
				}
				out = append(out, blockLines...)
				continue
			}

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

			if !alreadyHasOpenapiv2(blockLines) {
				if bracketDepth == 0 {
					blockLines = injectOpenapiv2IfNeeded(blockLines, moduleName, false)
				}
			}
			out = append(out, blockLines...)
			continue
		}

		out = append(out, ln)
		i++
	}
	return out
}

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

func injectOpenapiv2IfNeeded(blockLines []string, moduleName string, isEmptyBlock bool) []string {
	if isEmptyBlock {
		return unifyEmptyBlockForceAnnotation(blockLines, moduleName)
	}
	return injectOpenapiv2(blockLines, moduleName)
}

func unifyEmptyBlockForceAnnotation(blockLines []string, moduleName string) []string {
	if len(blockLines) == 0 {
		return blockLines
	}
	rpcLine := blockLines[0]

	rpcLine = strings.TrimSuffix(rpcLine, "{}")
	rpcLine = strings.TrimSuffix(rpcLine, "{};")
	rpcLine = strings.TrimRight(rpcLine, " \t")

	rpcLine += " {\n"
	rpcLine += fmt.Sprintf("		option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = { tags: [\"%s\"] };\n", moduleName)
	rpcLine += "}"

	newBlock := []string{rpcLine}
	return newBlock
}

func alreadyHasOpenapiv2(blockLines []string) bool {
	for _, ln := range blockLines {
		if hasOpenapiv2Regex.MatchString(ln) {
			return true
		}
	}
	return false
}

func injectOpenapiv2(blockLines []string, moduleName string) []string {
	lastIdx := len(blockLines) - 1
	if lastIdx < 0 {
		return blockLines
	}

	endIdx := -1
	for i := lastIdx; i >= 0; i-- {
		trim := strings.TrimSpace(blockLines[i])
		if strings.HasSuffix(trim, "};") || strings.HasSuffix(trim, "}") {
			endIdx = i
			break
		}
	}
	if endIdx < 0 {
		return blockLines
	}

	annotation := fmt.Sprintf("		option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = { tags: [\"%s\"] };", moduleName)

	finalLine := strings.TrimSpace(blockLines[endIdx])
	if strings.HasSuffix(finalLine, "}") || strings.HasSuffix(finalLine, "};") {
		newArr := append([]string{}, blockLines[:endIdx]...)
		newArr = append(newArr, annotation)
		newArr = append(newArr, blockLines[endIdx:]...)
		return newArr
	}
	return blockLines
}
