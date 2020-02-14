package group

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/modules/incubator/orm"
)

type keeper struct {
	key sdk.StoreKey

	// Group Table
	groupTable        orm.AutoUInt64Table
	groupByAdminIndex orm.Index

	// Group Member Table
	groupMemberTable         orm.NaturalKeyTable
	groupMemberByGroupIndex  orm.UInt64Index
	groupMemberByMemberIndex orm.Index

	// Group Account Table
	groupAccountTable        orm.NaturalKeyTable
	groupAccountByGroupIndex orm.UInt64Index
	groupAccountByAdminIndex orm.Index

	// ProposalBase Table
	ProposalBaseTable               orm.AutoUInt64Table
	ProposalBaseByGroupAccountIndex orm.Index
	ProposalBaseByProposerIndex     orm.Index

	// Vote Table
	voteTable               orm.NaturalKeyTable
	voteByProposalBaseIndex orm.UInt64Index
	voteByVoterIndex        orm.Index
}

const (
	// Group Table
	GroupTablePrefix        byte = 0x0
	GroupTableSeqPrefix     byte = 0x1
	GroupByAdminIndexPrefix byte = 0x2

	// Group Member Table
	GroupMemberTablePrefix         byte = 0x10
	GroupMemberByGroupIndexPrefix  byte = 0x11
	GroupMemberByMemberIndexPrefix byte = 0x12

	// Group Account Table
	GroupAccountTablePrefix        byte = 0x20
	GroupAccountByGroupIndexPrefix byte = 0x21
	GroupAccountByAdminIndexPrefix byte = 0x22

	// ProposalBase Table
	ProposalBaseTablePrefix               byte = 0x30
	ProposalBaseTableSeqPrefix            byte = 0x31
	ProposalBaseByGroupAccountIndexPrefix byte = 0x32
	ProposalBaseByProposerIndexPrefix     byte = 0x33

	// Vote Table
	VoteTablePrefix               byte = 0x40
	VoteByProposalBaseIndexPrefix byte = 0x41
	VoteByVoterIndexPrefix        byte = 0x42
)

func NewGroupKeeper(storeKey sdk.StoreKey) keeper {
	k := keeper{key: storeKey}

	//
	// Group Table
	//
	groupTableBuilder := orm.NewAutoUInt64TableBuilder(GroupTablePrefix, GroupTableSeqPrefix, storeKey, &GroupMetadata{})
	k.groupByAdminIndex = orm.NewIndex(groupTableBuilder, GroupByAdminIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		return []orm.RowID{val.(*GroupMetadata).Admin.Bytes()}, nil
	})
	k.groupTable = groupTableBuilder.Build()

	//
	// Group Member Table
	//
	groupMemberTableBuilder := orm.NewNaturalKeyTableBuilder(GroupMemberTablePrefix, storeKey, &GroupMember{}, orm.FixLengthIndexKeys(8))
	k.groupMemberByGroupIndex = orm.NewUInt64Index(groupMemberTableBuilder, GroupMemberByGroupIndexPrefix, func(val interface{}) ([]uint64, error) {
		group := val.(*GroupMember).Group
		return []uint64{uint64(group)}, nil
	})
	k.groupMemberByMemberIndex = orm.NewIndex(groupMemberTableBuilder, GroupMemberByMemberIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		member := val.(*GroupMember).Member
		return []orm.RowID{member.Bytes()}, nil
	})
	k.groupMemberTable = groupMemberTableBuilder.Build()

	//
	// Group Account Table
	//
	groupAccountTableBuilder := orm.NewNaturalKeyTableBuilder(GroupAccountTablePrefix, storeKey, &GroupAccountMetadataBase{}, orm.Max255DynamicLengthIndexKeyCodec{})
	k.groupAccountByGroupIndex = orm.NewUInt64Index(groupAccountTableBuilder, GroupAccountByGroupIndexPrefix, func(value interface{}) ([]uint64, error) {
		group := value.(*GroupAccountMetadataBase).Group
		return []uint64{uint64(group)}, nil
	})
	k.groupAccountByAdminIndex = orm.NewIndex(groupAccountTableBuilder, GroupAccountByAdminIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		admin := value.(*GroupAccountMetadataBase).Admin
		return []orm.RowID{admin.Bytes()}, nil
	})
	k.groupAccountTable = groupAccountTableBuilder.Build()

	//
	// ProposalBase Table
	//
	ProposalBaseTableBuilder := orm.NewAutoUInt64TableBuilder(ProposalBaseTablePrefix, ProposalBaseTableSeqPrefix, storeKey, &ProposalBase{})
	k.ProposalBaseByGroupAccountIndex = orm.NewIndex(ProposalBaseTableBuilder, ProposalBaseByGroupAccountIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		account := value.(*ProposalBase).GroupAccount
		return []orm.RowID{account.Bytes()}, nil

	})
	k.ProposalBaseByProposerIndex = orm.NewIndex(ProposalBaseTableBuilder, ProposalBaseByProposerIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		proposers := value.(*ProposalBase).Proposers
		r := make([]orm.RowID, len(proposers))
		for i := range proposers {
			r[i] = proposers[i].Bytes()
		}
		return r, nil
	})
	k.ProposalBaseTable = ProposalBaseTableBuilder.Build()

	//
	// Vote Table
	//
	voteTableBuilder := orm.NewNaturalKeyTableBuilder(VoteTablePrefix, storeKey, &Vote{}, orm.Max255DynamicLengthIndexKeyCodec{})
	k.voteByProposalBaseIndex = orm.NewUInt64Index(voteTableBuilder, VoteByProposalBaseIndexPrefix, func(value interface{}) ([]uint64, error) {
		return []uint64{uint64(value.(*Vote).Proposal)}, nil
	})
	k.voteByVoterIndex = orm.NewIndex(voteTableBuilder, VoteByVoterIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		return []orm.RowID{value.(*Vote).Voter.Bytes()}, nil
	})
	k.voteTable = voteTableBuilder.Build()

	return k
}

type Keeper interface {
	// Groups
	CreateGroup(ctx sdk.Context, admin sdk.AccAddress, members []Member, comment string) (GroupID, error)
	UpdateGroupMembers(ctx sdk.Context, group GroupID, membersUpdates []Member) error
	UpdateGroupAdmin(ctx sdk.Context, group GroupID, newAdmin sdk.AccAddress) error
	UpdateGroupComment(ctx sdk.Context, group GroupID, newComment string) error

	// Group Accounts
	CreateGroupAccount(ctx sdk.Context, admin sdk.AccAddress, group GroupID, policy DecisionPolicy, comment string) (sdk.AccAddress, error)
	UpdateGroupAccountAdmin(ctx sdk.Context, groupAcc sdk.AccAddress, newAdmin sdk.AccAddress) error
	UpdateGroupAccountDecisionPolicy(ctx sdk.Context, groupAcc sdk.AccAddress, newPolicy DecisionPolicy) error
	UpdateGroupAccountComment(ctx sdk.Context, groupAcc sdk.AccAddress, newComment string) error

	// ProposalBases

	// Propose returns a new ProposalBase ID and a populated sdk.Result which could return an error
	// or the result of execution if execNow was set to true
	Propose(ctx sdk.Context, groupAcc sdk.AccAddress, approvers []sdk.AccAddress, msgs []sdk.Msg, comment string, execNow bool) (id ProposalID, execResult sdk.Result)

	Vote(ctx sdk.Context, id ProposalID, voters []sdk.AccAddress, choice Choice) error

	// Exec attempts to execute the specified ProposalBase. If the ProposalBase is in a valid
	// state and has enough approvals, then it will be executed and its result will be
	// returned, otherwise the result will contain an error
	Exec(ctx sdk.Context, id ProposalID) sdk.Result
}
