package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime" // #nosec G702
	"runtime/debug"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"time"

	stderrors "github.com/pkg/errors"

	"github.com/cometbft/cometbft/libs/log"
	tmrpctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	evmsupport "github.com/artela-network/artela-rollkit/x/evm/types"

	evmtxs "github.com/artela-network/artela-rollkit/x/evm/txs"
)

type DebugBackend interface {
	TraceTransaction(hash common.Hash, config *evmsupport.TraceConfig) (interface{}, error)
	TraceBlock(height rpc.BlockNumber,
		config *evmsupport.TraceConfig,
		block *tmrpctypes.ResultBlock,
	) ([]*evmtxs.TxTraceResult, error)
	CosmosBlockByHash(blockHash common.Hash) (*tmrpctypes.ResultBlock, error)
	CosmosBlockByNumber(blockNum rpc.BlockNumber) (*tmrpctypes.ResultBlock, error)
	HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*ethtypes.Header, error)
	BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*ethtypes.Block, error)
}

// HandlerT keeps track of the cpu profiler and trace execution
type HandlerT struct {
	cpuFilename   string
	cpuFile       io.WriteCloser
	mu            sync.Mutex
	traceFilename string
	traceFile     io.WriteCloser
}

// DebugAPI is the collection of tracing APIs exposed over the private debugging endpoint.
type DebugAPI struct {
	ctx     *server.Context
	logger  log.Logger
	backend DebugBackend
	handler *HandlerT
}

// NewDebugAPI creates a new DebugAPI definition for the tracing methods of the Ethereum service.
func NewDebugAPI(
	backend DebugBackend,
) *DebugAPI {
	return &DebugAPI{
		backend: backend,
		handler: new(HandlerT),
	}
}

// TraceTransaction returns the structured logs created during the execution of EVM
// and returns them as a JSON object.
func (a *DebugAPI) TraceTransaction(hash common.Hash, config evmsupport.TraceConfig) (interface{}, error) {
	return a.backend.TraceTransaction(hash, &config)
}

// TraceBlockByNumber returns the structured logs created during the execution of
// EVM and returns them as a JSON object.
func (a *DebugAPI) TraceBlockByNumber(height rpc.BlockNumber, config evmsupport.TraceConfig) ([]*evmtxs.TxTraceResult, error) {
	a.logger.Debug("debug_traceBlockByNumber", "height", height)
	if height == 0 {
		return nil, errors.New("genesis is not traceable")
	}
	// Get Tendermint Block
	resBlock, err := a.backend.CosmosBlockByNumber(height)
	if err != nil {
		a.logger.Debug("get block failed", "height", height, "error", err.Error())
		return nil, err
	}

	return a.backend.TraceBlock(rpc.BlockNumber(resBlock.Block.Height), &config, resBlock)
}

// TraceBlockByHash returns the structured logs created during the execution of
// EVM and returns them as a JSON object.
func (a *DebugAPI) TraceBlockByHash(hash common.Hash, config evmsupport.TraceConfig) ([]*evmtxs.TxTraceResult, error) {
	a.logger.Debug("debug_traceBlockByHash", "hash", hash)
	// Get Tendermint Block
	resBlock, err := a.backend.CosmosBlockByHash(hash)
	if err != nil {
		a.logger.Debug("get block failed", "hash", hash.Hex(), "error", err.Error())
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		a.logger.Debug("block not found", "hash", hash.Hex())
		return nil, errors.New("block not found")
	}

	return a.backend.TraceBlock(rpc.BlockNumber(resBlock.Block.Height), &config, resBlock)
}

// BlockProfile turns on goroutine profiling for nsec seconds and writes profile data to
// file. It uses a profile rate of 1 for most accurate information. If a different rate is
// desired, set the rate and write the profile manually.
func (a *DebugAPI) BlockProfile(file string, nsec uint) error {
	a.logger.Debug("debug_blockProfile", "file", file, "nsec", nsec)
	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)

	time.Sleep(time.Duration(nsec) * time.Second)
	return writeProfile("block", file, a.logger)
}

// CpuProfile turns on CPU profiling for nsec seconds and writes
// profile data to file.
func (a *DebugAPI) CpuProfile(file string, nsec uint) error {
	a.logger.Debug("debug_cpuProfile", "file", file, "nsec", nsec)
	if err := a.StartCPUProfile(file); err != nil {
		return err
	}
	time.Sleep(time.Duration(nsec) * time.Second)
	return a.StopCPUProfile()
}

// GcStats returns GC statistics.
func (a *DebugAPI) GcStats() *debug.GCStats {
	a.logger.Debug("debug_gcStats")
	s := new(debug.GCStats)
	debug.ReadGCStats(s)
	return s
}

// GoTrace turns on tracing for nsec seconds and writes
// trace data to file.
func (a *DebugAPI) GoTrace(file string, nsec uint) error {
	a.logger.Debug("debug_goTrace", "file", file, "nsec", nsec)
	if err := a.StartGoTrace(file); err != nil {
		return err
	}
	time.Sleep(time.Duration(nsec) * time.Second)
	return a.StopGoTrace()
}

// MemStats returns detailed runtime memory statistics.
func (a *DebugAPI) MemStats() *runtime.MemStats {
	a.logger.Debug("debug_memStats")
	s := new(runtime.MemStats)
	runtime.ReadMemStats(s)
	return s
}

// SetBlockProfileRate sets the rate of goroutine block profile data collection.
// rate 0 disables block profiling.
func (a *DebugAPI) SetBlockProfileRate(rate int) {
	a.logger.Debug("debug_setBlockProfileRate", "rate", rate)
	runtime.SetBlockProfileRate(rate)
}

// Stacks returns a printed representation of the stacks of all goroutines.
func (a *DebugAPI) Stacks() string {
	a.logger.Debug("debug_stacks")
	buf := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buf, 2)
	if err != nil {
		a.logger.Error("Failed to create stacks", "error", err.Error())
	}
	return buf.String()
}

// StartCPUProfile turns on CPU profiling, writing to the given file.
func (a *DebugAPI) StartCPUProfile(file string) error {
	a.logger.Debug("debug_startCPUProfile", "file", file)
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	switch {
	case isCPUProfileConfigurationActivated(a.ctx):
		a.logger.Debug("CPU profiling already in progress using the configuration file")
		return errors.New("CPU profiling already in progress using the configuration file")
	case a.handler.cpuFile != nil:
		a.logger.Debug("CPU profiling already in progress")
		return errors.New("CPU profiling already in progress")
	default:
		fp, err := ExpandHome(file)
		if err != nil {
			a.logger.Debug("failed to get filepath for the CPU profile file", "error", err.Error())
			return err
		}
		f, err := os.Create(fp)
		if err != nil {
			a.logger.Debug("failed to create CPU profile file", "error", err.Error())
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			a.logger.Debug("cpu profiling already in use", "error", err.Error())
			if err := f.Close(); err != nil {
				a.logger.Debug("failed to close cpu profile file")
				return stderrors.Wrap(err, "failed to close cpu profile file")
			}
			return err
		}

		a.logger.Info("CPU profiling started", "profile", file)
		a.handler.cpuFile = f
		a.handler.cpuFilename = file
		return nil
	}
}

// StopCPUProfile stops an ongoing CPU profile.
func (a *DebugAPI) StopCPUProfile() error {
	a.logger.Debug("debug_stopCPUProfile")
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	switch {
	case isCPUProfileConfigurationActivated(a.ctx):
		a.logger.Debug("CPU profiling already in progress using the configuration file")
		return errors.New("CPU profiling already in progress using the configuration file")
	case a.handler.cpuFile != nil:
		a.logger.Info("Done writing CPU profile", "profile", a.handler.cpuFilename)
		pprof.StopCPUProfile()
		if err := a.handler.cpuFile.Close(); err != nil {
			a.logger.Debug("failed to close cpu file")
			return stderrors.Wrap(err, "failed to close cpu file")
		}
		a.handler.cpuFile = nil
		a.handler.cpuFilename = ""
		return nil
	default:
		a.logger.Debug("CPU profiling not in progress")
		return errors.New("CPU profiling not in progress")
	}
}

// WriteBlockProfile writes a goroutine blocking profile to the given file.
func (a *DebugAPI) WriteBlockProfile(file string) error {
	a.logger.Debug("debug_writeBlockProfile", "file", file)
	return writeProfile("block", file, a.logger)
}

// WriteMemProfile writes an allocation profile to the given file.
// Note that the profiling rate cannot be set through the DebugAPI,
// it must be set on the command line.
func (a *DebugAPI) WriteMemProfile(file string) error {
	a.logger.Debug("debug_writeMemProfile", "file", file)
	return writeProfile("heap", file, a.logger)
}

// MutexProfile turns on mutex profiling for nsec seconds and writes profile data to file.
// It uses a profile rate of 1 for most accurate information. If a different rate is
// desired, set the rate and write the profile manually.
func (a *DebugAPI) MutexProfile(file string, nsec uint) error {
	a.logger.Debug("debug_mutexProfile", "file", file, "nsec", nsec)
	runtime.SetMutexProfileFraction(1)
	time.Sleep(time.Duration(nsec) * time.Second)
	defer runtime.SetMutexProfileFraction(0)
	return writeProfile("mutex", file, a.logger)
}

// SetMutexProfileFraction sets the rate of mutex profiling.
func (a *DebugAPI) SetMutexProfileFraction(rate int) {
	a.logger.Debug("debug_setMutexProfileFraction", "rate", rate)
	runtime.SetMutexProfileFraction(rate)
}

// WriteMutexProfile writes a goroutine blocking profile to the given file.
func (a *DebugAPI) WriteMutexProfile(file string) error {
	a.logger.Debug("debug_writeMutexProfile", "file", file)
	return writeProfile("mutex", file, a.logger)
}

// FreeOSMemory forces a garbage collection.
func (a *DebugAPI) FreeOSMemory() {
	a.logger.Debug("debug_freeOSMemory")
	debug.FreeOSMemory()
}

// SetGCPercent sets the garbage collection target percentage. It returns the previous
// setting. A negative value disables GC.
func (a *DebugAPI) SetGCPercent(v int) int {
	a.logger.Debug("debug_setGCPercent", "percent", v)
	return debug.SetGCPercent(v)
}

// GetHeaderRlp retrieves the RLP encoded for of a single header.
func (a *DebugAPI) GetHeaderRlp(number uint64) (hexutil.Bytes, error) {
	header, err := a.backend.HeaderByNumber(context.TODO(), rpc.BlockNumber(number))
	if err != nil {
		return nil, err
	}

	return rlp.EncodeToBytes(header)
}

// GetBlockRlp retrieves the RLP encoded for of a single block.
func (a *DebugAPI) GetBlockRlp(number uint64) (hexutil.Bytes, error) {
	block, err := a.backend.BlockByNumber(context.TODO(), rpc.BlockNumber(number))
	if err != nil {
		return nil, err
	}

	return rlp.EncodeToBytes(block)
}

// PrintBlock retrieves a block and returns its pretty printed form.
func (a *DebugAPI) PrintBlock(number uint64) (string, error) {
	block, err := a.backend.BlockByNumber(context.TODO(), rpc.BlockNumber(number))
	if err != nil {
		return "", err
	}

	return spew.Sdump(block), nil
}

// SeedHash retrieves the seed hash of a block.
func (a *DebugAPI) SeedHash(_ uint64) (string, error) {
	return "", errors.New("SeedHash is not implemented")
}

// IntermediateRoots executes a block, and returns a list
// of intermediate roots: the stateroot after each transaction.
func (a *DebugAPI) IntermediateRoots(hash common.Hash, _ *evmsupport.TraceConfig) ([]common.Hash, error) {
	a.logger.Debug("debug_intermediateRoots", "hash", hash)
	return ([]common.Hash)(nil), nil
}

// StartGoTrace turns on tracing, writing to the given file.
func (a *DebugAPI) StartGoTrace(file string) error {
	a.logger.Debug("debug_startGoTrace", "file", file)
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	if a.handler.traceFile != nil {
		a.logger.Debug("trace already in progress")
		return errors.New("trace already in progress")
	}
	fp, err := ExpandHome(file)
	if err != nil {
		a.logger.Debug("failed to get filepath for the CPU profile file", "error", err.Error())
		return err
	}
	f, err := os.Create(fp)
	if err != nil {
		a.logger.Debug("failed to create go trace file", "error", err.Error())
		return err
	}
	if err := trace.Start(f); err != nil {
		a.logger.Debug("Go tracing already started", "error", err.Error())
		if err := f.Close(); err != nil {
			a.logger.Debug("failed to close trace file")
			return stderrors.Wrap(err, "failed to close trace file")
		}

		return err
	}
	a.handler.traceFile = f
	a.handler.traceFilename = file
	a.logger.Info("Go tracing started", "dump", a.handler.traceFilename)
	return nil
}

// StopGoTrace stops an ongoing trace.
func (a *DebugAPI) StopGoTrace() error {
	a.logger.Debug("debug_stopGoTrace")
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	trace.Stop()
	if a.handler.traceFile == nil {
		a.logger.Debug("trace not in progress")
		return errors.New("trace not in progress")
	}
	a.logger.Info("Done writing Go trace", "dump", a.handler.traceFilename)
	if err := a.handler.traceFile.Close(); err != nil {
		a.logger.Debug("failed to close trace file")
		return stderrors.Wrap(err, "failed to close trace file")
	}
	a.handler.traceFile = nil
	a.handler.traceFilename = ""
	return nil
}

// isCPUProfileConfigurationActivated checks if cpuprofile was configured via flag
func isCPUProfileConfigurationActivated(ctx *server.Context) bool {
	// TODO: use same constants as server/start.go
	// constant declared in start.go cannot be imported (cyclical dependency)
	const flagCPUProfile = "cpu-profile"
	if cpuProfile := ctx.Viper.GetString(flagCPUProfile); cpuProfile != "" {
		return true
	}
	return false
}

// ExpandHome expands home directory in file paths.
// ~someuser/tmp will not be expanded.
func ExpandHome(p string) (string, error) {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		usr, err := user.Current()
		if err != nil {
			return p, err
		}
		home := usr.HomeDir
		p = home + p[1:]
	}
	return filepath.Clean(p), nil
}

// writeProfile writes the data to a file
func writeProfile(name, file string, log log.Logger) error {
	p := pprof.Lookup(name)
	log.Info("Writing profile records", "count", p.Count(), "type", name, "dump", file)
	fp, err := ExpandHome(file)
	if err != nil {
		return err
	}
	f, err := os.Create(fp)
	if err != nil {
		return err
	}

	if err := p.WriteTo(f, 0); err != nil {
		if err := f.Close(); err != nil {
			return err
		}
		return err
	}

	return f.Close()
}
