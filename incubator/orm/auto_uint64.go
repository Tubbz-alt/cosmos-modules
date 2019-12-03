package orm

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ TableBuilder = &autoUInt64TableBuilder{}

func NewAutoUInt64TableBuilder(prefix []byte, key sdk.StoreKey, cdc *codec.Codec) autoUInt64TableBuilder {
	return autoUInt64TableBuilder{prefix: prefix, storeKey: key, cdc: cdc}
}

type autoUInt64TableBuilder struct {
	prefix      []byte
	storeKey    sdk.StoreKey
	cdc         *codec.Codec
	indexerRefs []indexRef
}

func (a autoUInt64TableBuilder) RowGetter() RowGetter {
	panic("implement me")
}

func (a autoUInt64TableBuilder) RegisterIndexer(prefix []byte, indexer Indexer) {
	// todo: fail on duplicates
	a.indexerRefs = append(a.indexerRefs, indexRef{prefix: prefix, indexer: indexer})
}

func (a autoUInt64TableBuilder) Build() autoUInt64Table {
	seq := NewSequence(a.storeKey, a.prefix)
	return autoUInt64Table{
		sequence:    seq,
		prefix:      a.prefix,
		storeKey:    a.storeKey,
		cdc:         a.cdc,
		indexerRefs: a.indexerRefs, // todo: clone slice
	}
}

func (a autoUInt64TableBuilder) AddAfterSaveInterceptor(interceptor AfterSaveInterceptor) {
	panic("implement me")
}

func (a autoUInt64TableBuilder) AddAfterDeleteInterceptor(interceptor AfterDeleteInterceptor) {
	panic("implement me")
}

var _ AutoUInt64Table = autoUInt64Table{}

type autoUInt64Table struct {
	prefix      []byte
	storeKey    sdk.StoreKey
	cdc         *codec.Codec
	indexerRefs []indexRef
	sequence    Sequence
}

func (a autoUInt64Table) Create(ctx HasKVStore, value interface{}) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(a.storeKey), a.prefix)
	v, err := a.cdc.MarshalBinaryBare(value)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to serialize %T", value)
	}
	id, err := a.sequence.NextVal(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "can not fetch next sequence value")
	}

	// todo: store does not return an error that we can handle or return
	store.Set(encodeSequence(id), v)
	return id, nil
}

func (a autoUInt64Table) Save(ctx HasKVStore, id uint64, value interface{}) error {
	store := prefix.NewStore(ctx.KVStore(a.storeKey), a.prefix)
	v, err := a.cdc.MarshalBinaryBare(value)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize %T", value)
	}
	// todo: store does not return an error that we can handle or return
	store.Set(encodeSequence(id), v)
	return nil
}


func (a autoUInt64Table) Has(ctx HasKVStore, key uint64) (bool, error) {
	panic("implement me")
}

func (a autoUInt64Table) Get(ctx HasKVStore, key uint64) (Iterator, error) {
	panic("implement me")
}

func (a autoUInt64Table) PrefixScan(ctx HasKVStore, start uint64, end uint64) (Iterator, error) {
	panic("implement me")
}

func (a autoUInt64Table) ReversePrefixScan(ctx HasKVStore, start uint64, end uint64) (Iterator, error) {
	panic("implement me")
}
