package orm

import (
	"math"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoUInt64PrefixScan(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("test")
	cdc := codec.New()
	const (
		testTablePrefix = iota
		testTableSeqPrefix
	)
	tb := NewAutoUInt64TableBuilder(testTablePrefix, testTableSeqPrefix, storeKey, cdc, &GroupMetadata{}).Build()
	ctx := NewMockContext()

	g1 := GroupMetadata{
		Description: "my test 1",
		Admin:       sdk.AccAddress([]byte("admin-address")),
	}
	g2 := GroupMetadata{
		Description: "my test 2",
		Admin:       sdk.AccAddress([]byte("admin-address")),
	}
	g3 := GroupMetadata{
		Description: "my test 3",
		Admin:       sdk.AccAddress([]byte("admin-address")),
	}
	for _, g := range []GroupMetadata{g1, g2, g3} {
		_, err := tb.Create(ctx, &g)
		require.NoError(t, err)
	}

	specs := map[string]struct {
		start, end uint64
		expResult  []GroupMetadata
		expRowIDs  [][]byte
		expError   *errors.Error
		method     func(ctx HasKVStore, start uint64, end uint64) (Iterator, error)
	}{
		"first element": {
			start:     1,
			end:       2,
			method:    tb.PrefixScan,
			expResult: []GroupMetadata{g1},
			expRowIDs: [][]byte{EncodeSequence(1)},
		},
		"first 2 elements": {
			start:     1,
			end:       3,
			method:    tb.PrefixScan,
			expResult: []GroupMetadata{g1, g2},
			expRowIDs: [][]byte{EncodeSequence(1), EncodeSequence(2)},
		},
		"first 3 elements": {
			start:     1,
			end:       4,
			method:    tb.PrefixScan,
			expResult: []GroupMetadata{g1, g2, g3},
			expRowIDs: [][]byte{EncodeSequence(1), EncodeSequence(2), EncodeSequence(3)},
		},
		"search with max end": {
			start:     1,
			end:       math.MaxUint64,
			method:    tb.PrefixScan,
			expResult: []GroupMetadata{g1, g2, g3},
			expRowIDs: [][]byte{EncodeSequence(1), EncodeSequence(2), EncodeSequence(3)},
		},
		"2 to end": {
			start:     2,
			end:       5,
			method:    tb.PrefixScan,
			expResult: []GroupMetadata{g2, g3},
			expRowIDs: [][]byte{EncodeSequence(2), EncodeSequence(3)},
		},
		"start before end should fail": {
			start:    2,
			end:      1,
			method:   tb.PrefixScan,
			expError: ErrArgument,
		},
		"start equals end should fail": {
			start:    1,
			end:      1,
			method:   tb.PrefixScan,
			expError: ErrArgument,
		},
		"reverse first element": {
			start:     1,
			end:       2,
			method:    tb.ReversePrefixScan,
			expResult: []GroupMetadata{g1},
			expRowIDs: [][]byte{EncodeSequence(1)},
		},
		"reverse first 2 elements": {
			start:     1,
			end:       3,
			method:    tb.ReversePrefixScan,
			expResult: []GroupMetadata{g2, g1},
			expRowIDs: [][]byte{EncodeSequence(2), EncodeSequence(1)},
		},
		"reverse first 3 elements": {
			start:     1,
			end:       4,
			method:    tb.ReversePrefixScan,
			expResult: []GroupMetadata{g3, g2, g1},
			expRowIDs: [][]byte{EncodeSequence(3), EncodeSequence(2), EncodeSequence(1)},
		},
		"reverse search with max end": {
			start:     1,
			end:       math.MaxUint64,
			method:    tb.ReversePrefixScan,
			expResult: []GroupMetadata{g3, g2, g1},
			expRowIDs: [][]byte{EncodeSequence(3), EncodeSequence(2), EncodeSequence(1)},
		},
		"reverse 2 to end": {
			start:     2,
			end:       5,
			method:    tb.ReversePrefixScan,
			expResult: []GroupMetadata{g3, g2},
			expRowIDs: [][]byte{EncodeSequence(3), EncodeSequence(2)},
		},
		"reverse start before end should fail": {
			start:    2,
			end:      1,
			method:   tb.ReversePrefixScan,
			expError: ErrArgument,
		},
		"reverse start equals end should fail": {
			start:    1,
			end:      1,
			method:   tb.ReversePrefixScan,
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
