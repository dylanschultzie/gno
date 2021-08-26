package sdk

import (
	"context"
	"time"

	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/log"
	"github.com/gnolang/gno/pkgs/store"
	"github.com/gnolang/gno/pkgs/store/gas"
)

/*
Context is a mostly immutable object contains all information needed to process
a request.

It contains a context.Context object inside if you want to use that, but please
do not over-use it. We try to keep all data structured and standard additions
here would be better just to add to the Context struct
*/
type Context struct {
	ctx           context.Context
	ms            store.MultiStore
	header        abci.Header
	chainID       string
	txBytes       []byte
	logger        log.Logger
	voteInfo      []abci.VoteInfo
	gasMeter      store.GasMeter
	blockGasMeter store.GasMeter
	checkTx       bool
	minGasPrices  []GasPrice
	consParams    *abci.ConsensusParams
	eventLogger   *EventLogger
}

// Proposed rename, not done to avoid API breakage
type Request = Context

// Read-only accessors
func (c Context) Context() context.Context      { return c.ctx }
func (c Context) MultiStore() store.MultiStore  { return c.ms }
func (c Context) BlockHeight() int64            { return c.header.GetHeight() }
func (c Context) BlockTime() time.Time          { return c.header.GetTime() }
func (c Context) ChainID() string               { return c.chainID }
func (c Context) TxBytes() []byte               { return c.txBytes }
func (c Context) Logger() log.Logger            { return c.logger }
func (c Context) VoteInfos() []abci.VoteInfo    { return c.voteInfo }
func (c Context) GasMeter() store.GasMeter      { return c.gasMeter }
func (c Context) BlockGasMeter() store.GasMeter { return c.blockGasMeter }
func (c Context) IsCheckTx() bool               { return c.checkTx }
func (c Context) MinGasPrices() []GasPrice      { return c.minGasPrices }
func (c Context) EventLogger() *EventLogger     { return c.eventLogger }

// clone the header before returning
func (c Context) BlockHeader() abci.Header {
	var msg = amino.DeepCopy(&c.header).(*abci.Header)
	return *msg
}

func (c Context) ConsensusParams() *abci.ConsensusParams {
	return amino.DeepCopy(c.consParams).(*abci.ConsensusParams)
}

// create a new context
func NewContext(ms store.MultiStore, header abci.Header, isCheckTx bool, logger log.Logger) Context {
	return Context{
		ctx:          context.Background(),
		ms:           ms,
		header:       header,
		chainID:      header.GetChainID(),
		checkTx:      isCheckTx,
		logger:       logger,
		gasMeter:     store.NewInfiniteGasMeter(),
		minGasPrices: nil,
		eventLogger:  NewEventLogger(),
	}
}

func (c Context) WithContext(ctx context.Context) Context {
	c.ctx = ctx
	return c
}

func (c Context) WithMultiStore(ms store.MultiStore) Context {
	c.ms = ms
	return c
}

func (c Context) WithBlockHeader(header abci.Header) Context {
	c.header = header
	return c
}

func (c Context) WithChainID(chainID string) Context {
	c.chainID = chainID
	return c
}

func (c Context) WithTxBytes(txBytes []byte) Context {
	c.txBytes = txBytes
	return c
}

func (c Context) WithLogger(logger log.Logger) Context {
	c.logger = logger
	return c
}

func (c Context) WithVoteInfos(voteInfo []abci.VoteInfo) Context {
	c.voteInfo = voteInfo
	return c
}

func (c Context) WithGasMeter(meter store.GasMeter) Context {
	c.gasMeter = meter
	return c
}

func (c Context) WithBlockGasMeter(meter store.GasMeter) Context {
	c.blockGasMeter = meter
	return c
}

func (c Context) WithIsCheckTx(isCheckTx bool) Context {
	c.checkTx = isCheckTx
	return c
}

func (c Context) WithMinGasPrices(gasPrices []GasPrice) Context {
	c.minGasPrices = gasPrices
	return c
}

func (c Context) WithConsensusParams(params *abci.ConsensusParams) Context {
	c.consParams = params
	return c
}

func (c Context) WithEventLogger(em *EventLogger) Context {
	c.eventLogger = em
	return c
}

// WithValue is deprecated, provided for backwards compatibility
// Please use
//     ctx = ctx.WithContext(context.WithValue(ctx.Context(), key, false))
// instead of
//     ctx = ctx.WithValue(key, false)
// NOTE: why?
func (c Context) WithValue(key, value interface{}) Context {
	c.ctx = context.WithValue(c.ctx, key, value)
	return c
}

// Value is deprecated, provided for backwards compatibility
// Please use
//     ctx.Context().Value(key)
// instead of
//     ctx.Value(key)
// NOTE: why?
func (c Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// ----------------------------------------------------------------------------
// Store / Caching
// ----------------------------------------------------------------------------

// Store fetches a Store from the MultiStore.
func (c Context) Store(key store.StoreKey) store.Store {
	return gas.New(c.MultiStore().GetStore(key), c.GasMeter(), store.DefaultGasConfig())
}

// CacheContext returns a new Context with the multi-store cached and a new
// EventLogger . The cached context is written to the context when writeCache
// is called.
func (c Context) CacheContext() (cc Context, writeCache func()) {
	cms := c.MultiStore().MultiCacheWrap()
	cc = c.WithMultiStore(cms).WithEventLogger(NewEventLogger())
	return cc, cms.MultiWrite
}