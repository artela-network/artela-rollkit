package keeper

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"sync"

	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/artela-network/artela-evm/vm"
	inherent "github.com/artela-network/aspect-core/chaincoreext/jit_inherent"
	"github.com/artela-network/aspect-core/djpm"
	aspcoretype "github.com/artela-network/aspect-core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethereum "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	artelatypes "github.com/artela-network/artela-rollkit/x/evm/artela/types"

	"github.com/artela-network/artela-rollkit/common"
	artela "github.com/artela-network/artela-rollkit/ethereum/types"
	"github.com/artela-network/artela-rollkit/x/aspect/provider"
	"github.com/artela-network/artela-rollkit/x/evm/artela/api"
	artvmtype "github.com/artela-network/artela-rollkit/x/evm/artela/types"
	"github.com/artela-network/artela-rollkit/x/evm/states"
	"github.com/artela-network/artela-rollkit/x/evm/txs"
	"github.com/artela-network/artela-rollkit/x/evm/types"
)

type (
	Keeper struct {
		cdc                   codec.BinaryCodec
		storeService          store.KVStoreService
		transientStoreService store.TransientStoreService
		accountKeeper         types.AccountKeeper
		bankKeeper            types.BankKeeper
		stakingKeeper         types.StakingKeeper
		feeKeeper             types.FeeKeeper
		aspectKeeper          types.AspectKeeper
		subSpace              types.ParamSubspace
		blockGetter           types.BlockGetter
		logger                log.Logger
		ChainIDGetter         types.ChainIDGetter

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		// miscellaneous
		// chain ID number obtained from the context's chain id
		eip155ChainID *big.Int

		// tracer used to collect execution traces from the EVM txs execution
		tracer string

		// keep the evm and matched stateDB instance just finished running
		aspectRuntimeContext *artvmtype.AspectRuntimeContext

		aspect *provider.ArtelaProvider

		clientContext client.Context

		// store the block context, this will be fresh every block.
		BlockContext *artvmtype.EthBlockContext

		// cache of aspect sig
		VerifySigCache *sync.Map
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	transientStoreService store.TransientStoreService,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	feeKeeper types.FeeKeeper,
	aspectKeeper types.AspectKeeper,
	blockGetter types.BlockGetter,
	chainIDGetter types.ChainIDGetter,
	logger log.Logger,
	authority string,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	// ensure evm module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the EVM module account has not been set")
	}

	// init aspect
	aspect := provider.NewArtelaProvider(storeService, aspectKeeper.GetStoreService(), artvmtype.GetLastBlockHeight(blockGetter))
	// new Aspect Runtime Context
	aspectRuntimeContext := artvmtype.NewAspectRuntimeContext()
	aspectRuntimeContext.Init(storeService, aspectKeeper.GetStoreService())

	// pass in the parameter space to the CommitStateDB in order to use custom denominations for the EVM operations
	k := Keeper{
		logger:                logger.With("module", fmt.Sprintf("x/%s", types.ModuleName)),
		cdc:                   cdc,
		authority:             authority,
		accountKeeper:         accountKeeper,
		bankKeeper:            bankKeeper,
		stakingKeeper:         stakingKeeper,
		feeKeeper:             feeKeeper,
		aspectKeeper:          aspectKeeper,
		storeService:          storeService,
		transientStoreService: transientStoreService,
		tracer:                "1",
		aspectRuntimeContext:  aspectRuntimeContext,
		aspect:                aspect,
		VerifySigCache:        new(sync.Map),
		ChainIDGetter:         chainIDGetter,
	}

	djpm.NewAspect(aspect, common.WrapLogger(k.logger.With("module", "aspect")))
	api.InitAspectGlobals(&k)

	// init aspect host api factory
	aspcoretype.GetEvmHostHook = api.GetEvmHostInstance
	aspcoretype.GetStateDbHook = api.GetStateDBHostInstance
	aspcoretype.GetAspectRuntimeContextHostHook = api.GetAspectRuntimeContextHostInstance
	aspcoretype.GetAspectStateHostHook = api.GetAspectStateHostInstance
	aspcoretype.GetAspectPropertyHostHook = api.GetAspectPropertyHostInstance
	aspcoretype.GetAspectTransientStorageHostHook = api.GetAspectTransientStorageHostInstance
	aspcoretype.GetAspectTraceHostHook = api.GetAspectTraceHostInstance

	aspcoretype.GetAspectContext = k.GetAspectContext
	aspcoretype.SetAspectContext = k.SetAspectContext

	aspcoretype.JITSenderAspectByContext = k.JITSenderAspectByContext
	aspcoretype.IsCommit = k.IsCommit
	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) SetClientContext(ctx client.Context) {
	k.clientContext = ctx
}

func (k Keeper) GetClientContext() client.Context {
	return k.clientContext
}

// InitChainID sets the chain id to the local variable in the keeper
func (k *Keeper) InitChainID() {
	chainId := k.ChainIDGetter()
	if k.eip155ChainID != nil {
		return
	}

	chainID, err := artela.ParseChainID(chainId)
	if err != nil {
		panic(err)
	}

	if k.eip155ChainID != nil && k.eip155ChainID.Cmp(chainID) != 0 {
		panic("chain id already set")
	}

	k.eip155ChainID = chainID
}

// ChainID returns the EIP155 chain ID for the EVM context
func (k Keeper) ChainID() *big.Int {
	if k.eip155ChainID == nil {
		k.InitChainID()
	}
	return k.eip155ChainID
}

// ----------------------------------------------------------------------------
// 								Block Bloom
// 							Required by Web3 API
// ----------------------------------------------------------------------------

// EmitBlockBloomEvent emit block bloom events
func (k Keeper) EmitBlockBloomEvent(ctx sdk.Context, bloom ethereum.Bloom) {
	encodedBloom := base64.StdEncoding.EncodeToString(bloom.Bytes())

	sprintf := fmt.Sprintf("emit block event %d bloom %s header %d, ", len(bloom.Bytes()), encodedBloom, ctx.BlockHeight())
	k.Logger().Debug(sprintf)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBlockBloom,
			sdk.NewAttribute(types.AttributeKeyEthereumBloom, encodedBloom),
		),
	)
}

// GetBlockBloomTransient returns bloom bytes for the current block height
func (k Keeper) GetBlockBloomTransient(ctx sdk.Context) *big.Int {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.transientStoreService.OpenTransientStore(ctx)), types.KeyPrefixTransientBloom)
	heightBz := sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight()))
	bz := store.Get(heightBz)
	if len(bz) == 0 {
		return big.NewInt(0)
	}

	return new(big.Int).SetBytes(bz)
}

// SetBlockBloomTransient sets the given bloom bytes to the transient store. This value is reset on
// every block.
func (k Keeper) SetBlockBloomTransient(ctx sdk.Context, bloom *big.Int) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.transientStoreService.OpenTransientStore(ctx)), types.KeyPrefixTransientBloom)
	heightBz := sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight()))
	store.Set(heightBz, bloom.Bytes())

	k.Logger().Debug(
		"setState: SetBlockBloomTransient",
		"block-height", ctx.BlockHeight(),
		"bloom", bloom.String(),
	)
}

// ----------------------------------------------------------------------------
// 								  Tx Index
// ----------------------------------------------------------------------------

// SetTxIndexTransient set the index of processing txs
func (k Keeper) SetTxIndexTransient(ctx sdk.Context, index uint64) {
	store := k.transientStoreService.OpenTransientStore(ctx)
	_ = store.Set(types.KeyPrefixTransientTxIndex, sdk.Uint64ToBigEndian(index))

	k.Logger().Debug(
		"setState: SetTxIndexTransient",
		"key", "KeyPrefixTransientTxIndex",
		"index", index,
	)
}

// GetTxIndexTransient returns EVM txs index on the current block.
func (k Keeper) GetTxIndexTransient(ctx sdk.Context) uint64 {
	store := k.transientStoreService.OpenTransientStore(ctx)
	bz, _ := store.Get(types.KeyPrefixTransientTxIndex)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// ----------------------------------------------------------------------------
// 									Log
// ----------------------------------------------------------------------------

// GetLogSizeTransient returns EVM log index on the current block.
func (k Keeper) GetLogSizeTransient(ctx sdk.Context) uint64 {
	store := k.transientStoreService.OpenTransientStore(ctx)
	bz, _ := store.Get(types.KeyPrefixTransientLogSize)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLogSizeTransient fetches the current EVM log index from the transient store, increases its
// value by one and then sets the new index back to the transient store.
func (k Keeper) SetLogSizeTransient(ctx sdk.Context, logSize uint64) {
	store := k.transientStoreService.OpenTransientStore(ctx)
	_ = store.Set(types.KeyPrefixTransientLogSize, sdk.Uint64ToBigEndian(logSize))

	k.Logger().Debug(
		"setState: SetLogSizeTransient",
		"key", "KeyPrefixTransientLogSize",
		"logSize", logSize,
	)
}

// ----------------------------------------------------------------------------
// 									Storage
// ----------------------------------------------------------------------------

// GetAccountStorage return states storage associated with an account
func (k Keeper) GetAccountStorage(ctx sdk.Context, address eth.Address) types.Storage {
	storage := types.Storage{}

	k.ForEachStorage(ctx, address, func(key, value eth.Hash) bool {
		storage = append(storage, types.NewState(key, value))
		return true
	})

	return storage
}

// ----------------------------------------------------------------------------
//									Account
// ----------------------------------------------------------------------------

// GetAccountWithoutBalance load nonce and codeHash without balance,
// more efficient in cases where balance is not needed.
func (k *Keeper) GetAccountWithoutBalance(ctx sdk.Context, addr eth.Address) *states.StateAccount {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	acct := k.accountKeeper.GetAccount(ctx, cosmosAddr)
	if acct == nil {
		return nil
	}

	codeHash := types.EmptyCodeHash
	ethAcct, ok := acct.(artela.EthAccountI)
	if ok {
		codeHash = ethAcct.GetCodeHash().Bytes()
	}

	return &states.StateAccount{
		Nonce:    acct.GetSequence(),
		CodeHash: codeHash,
	}
}

// GetAccountOrEmpty returns empty account if not exist, returns error if it's not `EthAccount`
func (k *Keeper) GetAccountOrEmpty(ctx sdk.Context, addr eth.Address) states.StateAccount {
	acct := k.GetAccount(ctx, addr)
	if acct != nil {
		return *acct
	}

	// empty account
	return states.StateAccount{
		Balance:  new(big.Int),
		CodeHash: types.EmptyCodeHash,
	}
}

// GetNonce returns the sequence number of an account, returns 0 if not exists.
func (k *Keeper) GetNonce(ctx sdk.Context, addr eth.Address) uint64 {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	acct := k.accountKeeper.GetAccount(ctx, cosmosAddr)
	if acct == nil {
		return 0
	}

	return acct.GetSequence()
}

// GetBalance load account's balance of gas token
func (k *Keeper) GetBalance(ctx sdk.Context, addr eth.Address) *big.Int {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	evmParams := k.GetParams(ctx)
	evmDenom := evmParams.GetEvmDenom()
	// if node is pruned, params is empty. Return invalid value
	if evmDenom == "" {
		return big.NewInt(-1)
	}
	coin := k.bankKeeper.GetBalance(ctx, cosmosAddr, evmDenom)
	return coin.Amount.BigInt()
}

// ----------------------------------------------------------------------------
// 								Gas and Fee
// ----------------------------------------------------------------------------

// Tracer return a default vm.Tracer based on current keeper states
func (k Keeper) Tracer(ctx sdk.Context, msg *core.Message, ethCfg *params.ChainConfig) vm.EVMLogger {
	return txs.NewTracer(k.tracer, msg, ethCfg, ctx.BlockHeight())
}

// GetBaseFee returns current base fee, return values:
// - `nil`: london hardfork not enabled.
// - `0`: london hardfork enabled but fee is not enabled.
// - `n`: both london hardfork and fee are enabled.
func (k Keeper) GetBaseFee(ctx sdk.Context, ethCfg *params.ChainConfig) *big.Int {
	return k.getBaseFee(ctx, types.IsLondon(ethCfg, ctx.BlockHeight()))
}

func (k Keeper) getBaseFee(ctx sdk.Context, london bool) *big.Int {
	if !london {
		return nil
	}
	baseFee := k.feeKeeper.GetBaseFee(ctx)
	if baseFee == nil {
		// return 0 if fee not enabled.
		baseFee = big.NewInt(0)
	}
	return baseFee
}

// GetMinGasMultiplier returns the MinGasMultiplier param from the fee market module
func (k Keeper) GetMinGasMultiplier(ctx sdk.Context) sdkmath.LegacyDec {
	feeParams := k.feeKeeper.GetParams(ctx)
	if feeParams.MinGasMultiplier.IsNil() {
		// in case we are executing eth_call on a legacy block, returns a zero value.
		return sdkmath.LegacyZeroDec()
	}
	return feeParams.MinGasMultiplier
}

// ResetTransientGasUsed reset gas used to prepare for execution of current cosmos txs, called in ante handler.
func (k Keeper) ResetTransientGasUsed(ctx sdk.Context) {
	store := k.transientStoreService.OpenTransientStore(ctx)
	_ = store.Delete(types.KeyPrefixTransientGasUsed)

	k.Logger().Debug("setState: ResetTransientGasUsed, delete", "key", "KeyPrefixTransientGasUsed")
}

// GetTransientGasUsed returns the gas used by current cosmos txs.
func (k Keeper) GetTransientGasUsed(ctx sdk.Context) uint64 {
	store := k.transientStoreService.OpenTransientStore(ctx)
	bz, _ := store.Get(types.KeyPrefixTransientGasUsed)
	if len(bz) == 0 {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// SetTransientGasUsed sets the gas used by current cosmos txs.
func (k Keeper) SetTransientGasUsed(ctx sdk.Context, gasUsed uint64) {
	store := k.transientStoreService.OpenTransientStore(ctx)
	bz := sdk.Uint64ToBigEndian(gasUsed)
	_ = store.Set(types.KeyPrefixTransientGasUsed, bz)

	k.Logger().Debug(
		"setState: SetTransientGasUsed, set",
		"key", "KeyPrefixTransientGasUsed",
		"gasUsed", fmt.Sprintf("%d", gasUsed),
	)
}

// AddTransientGasUsed accumulate gas used by each eth msg included in current cosmos txs.
func (k Keeper) AddTransientGasUsed(ctx sdk.Context, gasUsed uint64) (uint64, error) {
	result := k.GetTransientGasUsed(ctx) + gasUsed
	if result < gasUsed {
		return 0, errorsmod.Wrap(types.ErrGasOverflow, "transient gas used")
	}

	k.SetTransientGasUsed(ctx, result)
	return result, nil
}

// WithAspectContext creates the Aspect Context and establishes the link to the SDK context.
func (k Keeper) WithAspectContext(ctx sdk.Context, tx *ethereum.Transaction,
	evmConf *states.EVMConfig, block *artelatypes.EthBlockContext) (sdk.Context, *artelatypes.AspectRuntimeContext) {
	ethTxContext := artelatypes.NewEthTxContext(tx)
	ethTxContext.WithEVMConfig(evmConf)

	aspectCtx := artelatypes.NewAspectRuntimeContext()
	protocol := provider.NewAspectProtocolProvider(aspectCtx.EthTxContext)
	jitManager := inherent.NewManager(protocol)

	aspectCtx.SetEthTxContext(ethTxContext, jitManager)
	aspectCtx.WithCosmosContext(ctx)
	aspectCtx.SetEthBlockContext(block)
	aspectCtx.CreateStateObject()
	return ctx.WithValue(artelatypes.AspectContextKey, aspectCtx), aspectCtx
}
