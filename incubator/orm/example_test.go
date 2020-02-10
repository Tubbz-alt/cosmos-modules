package orm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GroupKeeper struct {
	key                      sdk.StoreKey
	groupTable               AutoUInt64Table
	groupByAdminIndex        Index
	groupMemberTable         NaturalKeyTable
	groupMemberByGroupIndex  Index
	groupMemberByMemberIndex Index
}

func (g GroupMember) NaturalKey() RowID {
	result := make([]byte, 0, len(g.Group)+len(g.Member))
	result = append(result, g.Group...)
	result = append(result, g.Member...)
	return result
}

var (
	GroupTablePrefix               byte = 0x0
	GroupTableSeqPrefix            byte = 0x1
	GroupByAdminIndexPrefix        byte = 0x2
	GroupMemberTablePrefix         byte = 0x3
	GroupMemberTableSeqPrefix      byte = 0x4
	GroupMemberTableIndexPrefix    byte = 0x5
	GroupMemberByGroupIndexPrefix  byte = 0x6
	GroupMemberByMemberIndexPrefix byte = 0x7
)

func NewGroupKeeper(storeKey sdk.StoreKey) GroupKeeper {
	k := GroupKeeper{key: storeKey}

	groupTableBuilder := NewAutoUInt64TableBuilder(GroupTablePrefix, GroupTableSeqPrefix, storeKey, &GroupMetadata{})
	// note: quite easy to mess with Index prefixes when managed outside. no fail fast on duplicates
	k.groupByAdminIndex = NewIndex(groupTableBuilder, GroupByAdminIndexPrefix, func(val interface{}) ([]RowID, error) {
		return []RowID{[]byte(val.(*GroupMetadata).Admin)}, nil
	})
	k.groupTable = groupTableBuilder.Build()

	groupMemberTableBuilder := NewNaturalKeyTableBuilder(GroupMemberTablePrefix, storeKey, &GroupMember{})

	k.groupMemberByGroupIndex = NewIndex(groupMemberTableBuilder, GroupMemberByGroupIndexPrefix, func(val interface{}) ([]RowID, error) {
		group := val.(*GroupMember).Group
		return []RowID{[]byte(group)}, nil
	})
	k.groupMemberByMemberIndex = NewIndex(groupMemberTableBuilder, GroupMemberByMemberIndexPrefix, func(val interface{}) ([]RowID, error) {
		return []RowID{[]byte(val.(*GroupMember).Member)}, nil
	})
	k.groupMemberTable = groupMemberTableBuilder.Build()

	return k
}
