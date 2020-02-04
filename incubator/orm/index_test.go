package orm

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexPrefixScan(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("test")
	cdc := codec.New()
	const (
		testTablePrefix = iota
		testTableSeqPrefix
	)
	tBuilder := NewAutoUInt64TableBuilder(testTablePrefix, testTableSeqPrefix, storeKey, cdc, &GroupMetadata{})
	idx := NewIndex(tBuilder, GroupByAdminIndexPrefix, func(val interface{}) ([][]byte, error) {
		return [][]byte{val.(*GroupMetadata).Admin}, nil
	})
	tb := tBuilder.Build()
	ctx := NewMockContext()

	g1 := GroupMetadata{
		Description: "my test 1",
		Admin:       sdk.AccAddress([]byte("admin-address-a")),
	}
	g2 := GroupMetadata{
		Description: "my test 2",
		Admin:       sdk.AccAddress([]byte("admin-address-b")),
	}
	g3 := GroupMetadata{
		Description: "my test 3",
		Admin:       sdk.AccAddress([]byte("admin-address-b")),
	}
	for _, g := range []GroupMetadata{g1, g2, g3} {
		_, err := tb.Create(ctx, &g)
		require.NoError(t, err)
	}

	specs := map[string]struct {
		start, end []byte
		expResult  []GroupMetadata
		expRowIDs  [][]byte
		expError   *errors.Error
		method     func(ctx HasKVStore, start, end []byte) (Iterator, error)
	}{
		"exact match with a single result": {
			start:     []byte("admin-address-a"),
			end:       []byte("admin-address-b"),
			method:    idx.PrefixScan,
			expResult: []GroupMetadata{g1},
			expRowIDs: [][]byte{EncodeSequence(1)},
		},
		"one result by prefix": {
			start:     []byte("admin-address"),
			end:       []byte("admin-address-b"),
			method:    idx.PrefixScan,
			expResult: []GroupMetadata{g1},
			expRowIDs: [][]byte{EncodeSequence(1)},
		},
		"multi key elements by exact match": {
			start:     []byte("admin-address-b"),
			end:       []byte("admin-address-c"),
			method:    idx.PrefixScan,
			expResult: []GroupMetadata{g2, g3},
			expRowIDs: [][]byte{EncodeSequence(2), EncodeSequence(3)},
		},
		"open end query": {
			start:     []byte("admin-address-b"),
			end:       nil,
			method:    idx.PrefixScan,
			expResult: []GroupMetadata{g2, g3},
			expRowIDs: [][]byte{EncodeSequence(2), EncodeSequence(3)},
		},
		"open start query": {
			start:     nil,
			end:       []byte("admin-address-b"),
			method:    idx.PrefixScan,
			expResult: []GroupMetadata{g1},
			expRowIDs: [][]byte{EncodeSequence(1)},
		},
		"open start and end query": {
			start:     nil,
			end:       nil,
			method:    idx.PrefixScan,
			expResult: []GroupMetadata{g1, g2, g3},
			expRowIDs: [][]byte{EncodeSequence(1), EncodeSequence(2), EncodeSequence(3)},
		},
		"all matching prefix": {
			start:     []byte("admin"),
			end:       nil,
			method:    idx.PrefixScan,
			expResult: []GroupMetadata{g1, g2, g3},
			expRowIDs: [][]byte{EncodeSequence(1), EncodeSequence(2), EncodeSequence(3)},
		},
		"non matching prefix": {
			start:     []byte("nobody"),
			end:       nil,
			method:    idx.PrefixScan,
			expResult: []GroupMetadata{},
		},
		"start equals end": {
			start:    []byte("any"),
			end:      []byte("any"),
			method:   idx.PrefixScan,
			expError: ErrArgument,
		},
		"start after end": {
			start:    []byte("b"),
			end:      []byte("a"),
			method:   idx.PrefixScan,
			expError: ErrArgument,
		},
		"reverse: exact match with a single result": {
			start:     []byte("admin-address-a"),
			end:       []byte("admin-address-b"),
			method:    idx.ReversePrefixScan,
			expResult: []GroupMetadata{g1},
			expRowIDs: [][]byte{EncodeSequence(1)},
		},
		"reverse: one result by prefix": {
			start:     []byte("admin-address"),
			end:       []byte("admin-address-b"),
			method:    idx.ReversePrefixScan,
			expResult: []GroupMetadata{g1},
			expRowIDs: [][]byte{EncodeSequence(1)},
		},
		"reverse: multi key elements by exact match": {
			start:     []byte("admin-address-b"),
			end:       []byte("admin-address-c"),
			method:    idx.ReversePrefixScan,
			expResult: []GroupMetadata{g3, g2},
			expRowIDs: [][]byte{EncodeSequence(3), EncodeSequence(2)},
		},
		"reverse: open end query": {
			start:     []byte("admin-address-b"),
			end:       nil,
			method:    idx.ReversePrefixScan,
			expResult: []GroupMetadata{g3, g2},
			expRowIDs: [][]byte{EncodeSequence(3), EncodeSequence(2)},
		},
		"reverse: open start query": {
			start:     nil,
			end:       []byte("admin-address-b"),
			method:    idx.ReversePrefixScan,
			expResult: []GroupMetadata{g1},
			expRowIDs: [][]byte{EncodeSequence(1)},
		},
		"reverse: open start and end query": {
			start:     nil,
			end:       nil,
			method:    idx.ReversePrefixScan,
			expResult: []GroupMetadata{g3, g2, g1},
			expRowIDs: [][]byte{EncodeSequence(3), EncodeSequence(2), EncodeSequence(1)},
		},
		"reverse: all matching prefix": {
			start:     []byte("admin"),
			end:       nil,
			method:    idx.ReversePrefixScan,
			expResult: []GroupMetadata{g3, g2, g1},
			expRowIDs: [][]byte{EncodeSequence(3), EncodeSequence(2), EncodeSequence(1)},
		},
		"reverse: non matching prefix": {
			start:     []byte("nobody"),
			end:       nil,
			method:    idx.ReversePrefixScan,
			expResult: []GroupMetadata{},
		},
		"reverse: start equals end": {
			start:    []byte("any"),
			end:      []byte("any"),
			method:   idx.ReversePrefixScan,
			expError: ErrArgument,
		},
		"reverse: start after end": {
			start:    []byte("b"),
			end:      []byte("a"),
			method:   idx.ReversePrefixScan,
			expError: ErrArgument,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			it, err := spec.method(ctx, spec.start, spec.end)
			require.True(t, spec.expError.Is(err), "expected #+v but got #+v", spec.expError, err)
			if spec.expError != nil {
				return
			}
			var loaded []GroupMetadata
			rowIDs, err := ReadAll(it, &loaded)
			require.NoError(t, err)
			assert.Equal(t, spec.expResult, loaded)
			assert.Equal(t, spec.expRowIDs, rowIDs)
		})
	}
}

func TestUInt64Index(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("test")
	cdc := codec.New()

	groupMemberTableBuilder := NewNaturalKeyTableBuilder(GroupMemberTablePrefix, GroupMemberTableSeqPrefix, GroupMemberTableIndexPrefix, storeKey, cdc, &GroupMember{})
	idx := NewUInt64Index(groupMemberTableBuilder, GroupMemberByMemberIndexPrefix, func(val interface{}) ([]uint64, error) {
		return []uint64{uint64(val.(*GroupMember).Member[0])}, nil
	})
	groupMemberTable := groupMemberTableBuilder.Build()

	ctx := NewMockContext()

	m := GroupMember{
		Group:  sdk.AccAddress(EncodeSequence(1)),
		Member: sdk.AccAddress([]byte("member-address")),
		Weight: sdk.NewInt(10),
	}
	err := groupMemberTable.Create(ctx, &m)
	require.NoError(t, err)

	indexedKey := uint64('m')

	// Has
	assert.True(t, idx.Has(ctx, indexedKey))

	// Get
	it, err := idx.Get(ctx, indexedKey)
	require.NoError(t, err)
	var loaded GroupMember
	rowID, err := it.LoadNext(&loaded)
	require.NoError(t, err)
	require.Equal(t, uint64(1), DecodeSequence(rowID))
	require.Equal(t, m, loaded)

	// PrefixScan match
	it, err = idx.PrefixScan(ctx, 0, 255)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.NoError(t, err)
	require.Equal(t, uint64(1), DecodeSequence(rowID))
	require.Equal(t, m, loaded)

	// PrefixScan no match
	it, err = idx.PrefixScan(ctx, indexedKey+1, 255)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.Error(t, ErrIteratorDone, err)

	// ReversePrefixScan match
	it, err = idx.ReversePrefixScan(ctx, 0, 255)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.NoError(t, err)
	require.Equal(t, uint64(1), DecodeSequence(rowID))
	require.Equal(t, m, loaded)

	// ReversePrefixScan no match
	it, err = idx.ReversePrefixScan(ctx, indexedKey+1, 255)
	require.NoError(t, err)
	rowID, err = it.LoadNext(&loaded)
	require.Error(t, ErrIteratorDone, err)
}
