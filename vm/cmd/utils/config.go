package utils

import (
	"math/big"
	"os"
	"path/filepath"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/vangjvn/devchain/api"
	"github.com/vangjvn/devchain/vm/ethereum"
	cli "gopkg.in/urfave/cli.v1"
)

const (
	// Client identifier to advertise over the network
	clientIdentifier = "vm"
	// Environment variable for home dir
	emHome = "EMHOME"
)

var (
	// GenesisTargetGasLimit is the target gas limit of the Genesis block.
	// #unstable
	GenesisTargetGasLimit = big.NewInt(100000000)
)

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}

type gethConfig struct {
	Eth      eth.Config
	Node     node.Config
	Ethstats ethstatsConfig
}

// MakeFullNode creates a full go-ethereum node
// #unstable
func MakeFullNode(ctx *cli.Context) *ethereum.Node {
	stack, cfg := makeConfigNode(ctx)

	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return api.NewBackend(ctx, &cfg.Eth)
	}); err != nil {
		ethUtils.Fatalf("Failed to register the ABCI application service: %v", err)
	}

	return stack
}

func makeConfigNode(ctx *cli.Context) (*ethereum.Node, gethConfig) {
	cfg := gethConfig{
		Eth:  DefaultEthConfig(),
		Node: DefaultNodeConfig(),
	}

	ethUtils.SetNodeConfig(ctx, &cfg.Node)
	SetEthermintNodeConfig(&cfg.Node)
	stack, err := ethereum.New(&cfg.Node)
	if err != nil {
		ethUtils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	ethUtils.SetEthConfig(ctx, &stack.Node, &cfg.Eth)
	SetEthermintEthConfig(&cfg.Eth)

	return stack, cfg
}

func DefaultEthConfig() eth.Config {
	cfg := eth.DefaultConfig

	// Get Ewasm interpreter
	if path, ok := os.LookupEnv("EVMC_LIBRARY_PATH"); ok {
		if lib, ok := os.LookupEnv("EVMC_LIBRARY"); ok {
			cfg.EWASMInterpreter = filepath.Join(path, lib)
		} else {
			cfg.EWASMInterpreter = filepath.Join(path, "libssvmEVMC.so")
		}
		vm.InitEVMCEwasm(cfg.EWASMInterpreter)
	}

	return cfg
}

// DefaultNodeConfig returns the default configuration for a go-ethereum node
// #unstable
func DefaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.Version
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	cfg.WSModules = append(cfg.WSModules, "eth")
	cfg.IPCPath = "cybermiles.ipc"
	cfg.InsecureUnlockAllowed = true
	cfg.NoUSB = true

	emHome := os.Getenv(emHome)
	if emHome != "" {
		cfg.DataDir = emHome
	}

	return cfg
}

// SetEthermintNodeConfig takes a node configuration and applies ethermint specific configuration
// #unstable
func SetEthermintNodeConfig(cfg *node.Config) {
	cfg.P2P.MaxPeers = 0
	cfg.P2P.NoDiscovery = true
}

// SetEthermintEthConfig takes a ethereum configuration and applies ethermint specific configuration
// #unstable
func SetEthermintEthConfig(cfg *eth.Config) {
	//cfg.MaxPeers = 0
	cfg.Ethash.PowMode = ethash.ModeFake
}

// MakeDataDir retrieves the currently requested data directory
// #unstable
func MakeDataDir(ctx *cli.Context) string {
	path := node.DefaultDataDir()

	emHome := os.Getenv(emHome)
	if emHome != "" {
		path = emHome
	}

	if ctx.GlobalIsSet(ethUtils.DataDirFlag.Name) {
		path = ctx.GlobalString(ethUtils.DataDirFlag.Name)
	}

	if path == "" {
		ethUtils.Fatalf("Cannot determine default data directory, please set manually (--datadir)")
	}

	return path
}
