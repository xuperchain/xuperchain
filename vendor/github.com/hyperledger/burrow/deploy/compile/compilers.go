package compile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"golang.org/x/crypto/sha3"

	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/logging"
	hex "github.com/tmthrgd/go-hex"
)

// SolidityInput is a structure for the solidity compiler input json form, see:
// https://solidity.readthedocs.io/en/v0.5.9/using-the-compiler.html#compiler-input-and-output-json-description
type SolidityInput struct {
	Language string                         `json:"language"`
	Sources  map[string]SolidityInputSource `json:"sources"`
	Settings struct {
		Libraries map[string]map[string]string `json:"libraries"`
		Optimizer struct {
			Enabled bool `json:"enabled"`
		} `json:"optimizer"`
		OutputSelection struct {
			File struct {
				OutputType []string `json:"*"`
			} `json:"*"`
		} `json:"outputSelection"`
	} `json:"settings"`
}

// SolidityInputSource should be set for each solidity input source file in SolidityInput
type SolidityInputSource struct {
	Content string   `json:"content,omitempty"`
	Urls    []string `json:"urls,omitempty"`
}

// SolidityOutput is a structure for the output of the solidity json output form
type SolidityOutput struct {
	Contracts map[string]map[string]SolidityContract
	Errors    []struct {
		Component        string
		FormattedMessage string
		Message          string
		Severity         string
		Type             string
	}
}

// SolidityContract is defined for each contract defined in the solidity source code
type SolidityContract struct {
	Abi json.RawMessage
	Evm struct {
		Bytecode         ContractCode
		DeployedBytecode ContractCode
	}
	EWasm struct {
		Wasm string
	}
	Devdoc   json.RawMessage
	Userdoc  json.RawMessage
	Metadata string
	// This is not present in the solidity output, but we add it ourselves
	// This is map from DeployedBytecode to Metadata. A Solidity contract can create any number
	// of contracts, which have distinct metadata. This is a map for the deployed code to metdata,
	// including the first contract itself.
	MetadataMap []MetadataMap `json:",omitempty"`
}

type ContractCode struct {
	Object         string
	LinkReferences json.RawMessage
}

// SolidityMetadata is the json field metadata
type SolidityMetadata struct {
	Version string
	// The solidity compiler needs to tell us it compiles solidity
	Language string
	Compiler struct {
		Version   string
		Keccak256 string
	}
	Sources map[string]struct {
		Keccak256 string
		Content   string
		Urls      []string
	}
	// Other fields elided, see https://solidity.readthedocs.io/en/v0.5.10/metadata.html
}

type Metadata struct {
	ContractName    string
	SourceFile      string
	CompilerVersion string
	Abi             json.RawMessage
}

type MetadataMap struct {
	DeployedBytecode ContractCode
	Metadata         Metadata
}

type Response struct {
	Objects []ResponseItem `json:"objects"`
	Warning string         `json:"warning"`
	Version string         `json:"version"`
	Error   string         `json:"error"`
}

// Compile response object
type ResponseItem struct {
	Filename   string           `json:"filename"`
	Objectname string           `json:"objectname"`
	Contract   SolidityContract `json:"binary"`
}

// LoadSolidityContract is the opposite of the .Save() method. This expects the input file
// to be in the Solidity json output format
func LoadSolidityContract(file string) (*SolidityContract, error) {
	codeB, err := ioutil.ReadFile(file)
	if err != nil {
		return &SolidityContract{}, err
	}
	contract := SolidityContract{}
	err = json.Unmarshal(codeB, &contract)
	if err != nil {
		return &SolidityContract{}, err
	}
	return &contract, nil
}

// Save persists the contract in its json form to disk
func (contract *SolidityContract) Save(dir, file string) error {
	str, err := json.Marshal(*contract)
	if err != nil {
		return err
	}
	// This will make the contract file appear atomically
	// This is important since if we run concurrent jobs, one job could be compiling a solidity
	// file while another reads the bin file. If write is incomplete, it will result in failures
	f, err := ioutil.TempFile(dir, "bin.*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(str)
	if err != nil {
		return err
	}
	f.Close()
	return os.Rename(f.Name(), filepath.Join(dir, file))
}

func link(bytecode string, linkReferences json.RawMessage, libraries map[string]string) (string, error) {
	var links map[string]map[string][]struct{ Start, Length int }
	err := json.Unmarshal(linkReferences, &links)
	if err != nil {
		return "", err
	}
	for _, f := range links {
		for name, relos := range f {
			addr, ok := libraries[name]
			if !ok {
				return "", fmt.Errorf("library %s is not defined", name)
			}
			for _, relo := range relos {
				if relo.Length != crypto.AddressLength {
					return "", fmt.Errorf("linkReference should be %d bytes long, not %d", crypto.AddressLength, relo.Length)
				}
				if len(addr) != crypto.AddressHexLength {
					return "", fmt.Errorf("address %s should be %d character long, not %d", addr, crypto.AddressHexLength, len(addr))
				}
				start := relo.Start * 2
				end := relo.Start*2 + crypto.AddressHexLength
				if bytecode[start+1] != '_' || bytecode[end-1] != '_' {
					return "", fmt.Errorf("relocation dummy not found at %d in %s ", relo.Start, bytecode)
				}
				bytecode = bytecode[:start] + addr + bytecode[end:]
			}
		}
	}

	return bytecode, nil
}

// Link will replace the unresolved references with the libraries provided
func (contract *SolidityContract) Link(libraries map[string]string) error {
	bin := contract.Evm.Bytecode.Object
	if strings.Contains(bin, "_") {
		bin, err := link(bin, contract.Evm.Bytecode.LinkReferences, libraries)
		if err != nil {
			return err
		}
		contract.Evm.Bytecode.Object = bin
	}

	// When compiling a solidity file with many contracts contained it, some of those contracts might
	// never be created by the contract we're current linking. However, Solidity does not tell us
	// which contracts can be created by a contract.
	// See: https://github.com/ethereum/solidity/issues/7111
	// Some of these contracts might have unresolved libraries. We can safely skip those contracts.
	if contract.MetadataMap != nil {
		for i, m := range contract.MetadataMap {
			bin := m.DeployedBytecode.Object
			if strings.Contains(bin, "_") {
				bin, err := link(bin, m.DeployedBytecode.LinkReferences, libraries)
				if err != nil {
					continue
				}
				contract.MetadataMap[i].DeployedBytecode.Object = bin
			}
		}
	}

	return nil
}

func (contract *SolidityContract) Code() (code string) {
	code = contract.Evm.Bytecode.Object
	if code == "" {
		code = contract.EWasm.Wasm
	}
	return
}

func EVM(file string, optimize bool, workDir string, libraries map[string]string, logger *logging.Logger) (*Response, error) {
	input := SolidityInput{Language: "Solidity", Sources: make(map[string]SolidityInputSource)}

	input.Sources[file] = SolidityInputSource{Urls: []string{file}}
	input.Settings.Optimizer.Enabled = optimize
	input.Settings.OutputSelection.File.OutputType = []string{"abi", "evm.deployedBytecode.object", "evm.bytecode.linkReferences", "metadata", "bin", "devdoc"}
	input.Settings.Libraries = make(map[string]map[string]string)
	input.Settings.Libraries[""] = make(map[string]string)

	for l, a := range libraries {
		input.Settings.Libraries[""][l] = "0x" + a
	}

	command, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	logger.TraceMsg("Command Input", "command", string(command))
	result, err := runSolidity(string(command), workDir)
	if err != nil {
		return nil, err
	}
	logger.TraceMsg("Command Output", "result", result)

	output := SolidityOutput{}
	err = json.Unmarshal([]byte(result), &output)
	if err != nil {
		return nil, err
	}

	// Collect our ABIs
	metamap := make([]MetadataMap, 0)
	for filename, src := range output.Contracts {
		for contractname, item := range src {
			var meta SolidityMetadata
			_ = json.Unmarshal([]byte(item.Metadata), &meta)
			if item.Evm.DeployedBytecode.Object != "" {
				metamap = append(metamap, MetadataMap{
					DeployedBytecode: item.Evm.DeployedBytecode,
					Metadata: Metadata{
						ContractName:    contractname,
						SourceFile:      filename,
						CompilerVersion: meta.Compiler.Version,
						Abi:             item.Abi,
					},
				})
			}
		}
	}

	respItemArray := make([]ResponseItem, 0)

	for f, s := range output.Contracts {
		for contract, item := range s {
			item.MetadataMap = metamap
			respItem := ResponseItem{
				Filename:   f,
				Objectname: objectName(contract),
				Contract:   item,
			}
			respItemArray = append(respItemArray, respItem)
		}
	}

	warnings := ""
	errors := ""
	for _, msg := range output.Errors {
		if msg.Type == "Warning" {
			warnings += msg.FormattedMessage
		} else {
			errors += msg.FormattedMessage
		}
	}

	for _, re := range respItemArray {
		logger.TraceMsg("Response formulated",
			"name", re.Objectname,
			"bin", re.Contract.Code(),
			"abi", string(re.Contract.Abi))
	}

	resp := Response{
		Objects: respItemArray,
		Warning: warnings,
		Error:   errors,
	}

	return &resp, nil
}

func WASM(file string, workDir string, logger *logging.Logger) (*Response, error) {
	shellCmd := exec.Command("solang", "--target", "ewasm", "--standard-json", file)
	if workDir != "" {
		shellCmd.Dir = workDir
	}
	output, err := shellCmd.CombinedOutput()
	if err != nil {
		logger.InfoMsg("solang failed", "output", string(output))
		return nil, err
	}
	logger.TraceMsg("Command Output", "result", string(output))

	wasmoutput := SolidityOutput{}
	err = json.Unmarshal(output, &wasmoutput)
	if err != nil {
		return nil, err
	}

	respItemArray := make([]ResponseItem, 0)

	for f, s := range wasmoutput.Contracts {
		for contract, item := range s {
			respItem := ResponseItem{
				Filename:   f,
				Objectname: objectName(contract),
				Contract:   item,
			}
			respItemArray = append(respItemArray, respItem)
		}
	}

	warnings := ""
	errors := ""
	for _, msg := range wasmoutput.Errors {
		if msg.Type == "Warning" {
			warnings += msg.FormattedMessage
		} else {
			errors += msg.FormattedMessage
		}
	}

	for _, re := range respItemArray {
		logger.TraceMsg("Response formulated",
			"name", re.Objectname,
			"bin", re.Contract.Evm.Bytecode.Object,
			"abi", string(re.Contract.Abi))
	}

	resp := Response{
		Objects: respItemArray,
		Warning: warnings,
		Error:   errors,
	}

	return &resp, nil
}

func objectName(contract string) string {
	if contract == "" {
		return ""
	}
	parts := strings.Split(strings.TrimSpace(contract), ":")
	return parts[len(parts)-1]
}

func runSolidity(jsonCmd string, workDir string) (string, error) {
	buf := bytes.NewBufferString(jsonCmd)
	shellCmd := exec.Command("solc", "--standard-json", "--allow-paths", "/")
	if workDir != "" {
		shellCmd.Dir = workDir
	}
	shellCmd.Stdin = buf
	output, err := shellCmd.CombinedOutput()
	s := string(output)
	return s, err
}

func PrintResponse(resp Response, cli bool, logger *logging.Logger) {
	if resp.Error != "" {
		logger.InfoMsg("solidity error", "errors", resp.Error)
	} else {
		for _, r := range resp.Objects {
			logger.InfoMsg("Response",
				"name", r.Objectname,
				"bin", r.Contract.Code(),
				"abi", string(r.Contract.Abi[:]),
				"link", string(r.Contract.Evm.Bytecode.LinkReferences[:]),
			)
		}
	}
}

// GetMetadata get the CodeHashes + Abis for the generated Code. So, we have a map for all the possible contracts codes hashes to abis
func (contract *SolidityContract) GetMetadata(logger *logging.Logger) (map[acmstate.CodeHash]string, error) {
	res := make(map[acmstate.CodeHash]string)
	if contract.Evm.DeployedBytecode.Object == "" {
		return nil, nil
	}

	for _, m := range contract.MetadataMap {
		if strings.Contains(m.DeployedBytecode.Object, "_") {
			continue
		}
		runtime, err := hex.DecodeString(m.DeployedBytecode.Object)
		if err != nil {
			return nil, err
		}

		bs, err := json.Marshal(m.Metadata)
		if err != nil {
			return nil, err
		}

		hash := sha3.NewLegacyKeccak256()
		hash.Write(runtime)
		var codehash acmstate.CodeHash
		copy(codehash[:], hash.Sum(nil))
		logger.TraceMsg("Found metadata",
			"code", fmt.Sprintf("%X", runtime),
			"code hash", fmt.Sprintf("%X", codehash),
			"meta", string(bs))
		res[codehash] = string(bs)
	}
	return res, nil
}

// GetDeployCodeHash deals with the issue described in https://github.com/ethereum/solidity/issues/7101
// When a library contract (one declared with "libary { }" rather than "contract { }"), the deployed code
// will not match what the solidity compiler said it would be. This is done to implement "call protection";
// library contracts are only supposed to be called from our solidity contracts, not directly. To prevent
// this, the library deployed code compares the callee address with the contract address itself. If it equal,
// it calls revert.
// The library contract address is only known post-deploy so this issue can only be handled post-deploy. This
// is why this is not dealt with during deploy time.
func GetDeployCodeHash(code []byte, address crypto.Address) []byte {
	if bytes.HasPrefix(code, append([]byte{byte(asm.PUSH20)}, address.Bytes()...)) {
		code = append([]byte{byte(asm.PUSH20)}, append(make([]byte, crypto.AddressLength), code[crypto.AddressLength+1:]...)...)
	}

	hash := sha3.NewLegacyKeccak256()
	hash.Write(code)
	return hash.Sum(nil)
}
