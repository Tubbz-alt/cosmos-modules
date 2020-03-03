package group

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/modules/incubator/orm"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"time"
)

type GroupID uint64

func (p GroupID) Uint64() uint64 {
	return uint64(p)
}

func (g GroupID) Byte() []byte {
	return orm.EncodeSequence(uint64(g))
}

type ProposalID uint64

func (p ProposalID) Byte() []byte {
	return orm.EncodeSequence(uint64(p))
}

func (p ProposalID) Uint64() uint64 {
	return uint64(p)
}

type DecisionPolicyResult struct {
	Allow bool
	Final bool
}

type DecisionPolicy interface {
	Allow(tally Tally, totalPower sdk.Dec, votingDuration time.Duration) (DecisionPolicyResult, error)
}

func (p ThresholdDecisionPolicy) Allow(tally Tally, totalPower sdk.Dec, votingDuration time.Duration) (DecisionPolicyResult, error) {
	timeout, err := types.DurationFromProto(&p.Timout)
	if err != nil {
		return DecisionPolicyResult{}, err
	}
	if timeout < votingDuration {
		return DecisionPolicyResult{Allow: false, Final: true}, nil
	}
	if tally.YesCount.GT(p.Threshold) {
		return DecisionPolicyResult{Allow: true, Final: true}, nil
	}
	undecided := totalPower.Sub(tally.TotalCounts())
	if tally.YesCount.Add(undecided).LTE(p.Threshold) {
		return DecisionPolicyResult{Allow: false, Final: true}, nil
	}
	return DecisionPolicyResult{Allow: false, Final: false}, nil
}

func (g GroupMember) NaturalKey() []byte {
	result := make([]byte, 8, 8+len(g.Member))
	copy(result[0:8], g.Group.Byte())
	result = append(result, g.Member...)
	return result
}

func (g GroupAccountMetadataBase) NaturalKey() []byte {
	return g.GroupAccount
}

func (g StdGroupAccountMetadata) NaturalKey() []byte {
	return g.Base.NaturalKey()
}

func (v Vote) NaturalKey() []byte {
	result := make([]byte, 8, 8+len(v.Voter))
	copy(result[0:8], v.Proposal.Byte())
	result = append(result, v.Voter...)
	return result
}

func (m Member) ValidateBasic() error {
	if m.Address.Empty() {
		return sdkerrors.Wrap(ErrEmpty, "address")
	}
	return nil
}

const defaultMaxCommentLength = 255

// Parameter keys
var (
	ParamMaxCommentLength = []byte("MaxCommentLength")
)

// DefaultParams returns the default parameters for the group module.
func DefaultParams() Params {
	return Params{
		MaxCommentLength: defaultMaxCommentLength,
	}
}

func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(ParamMaxCommentLength, &p.MaxCommentLength, noopValidator()),
	}
}
func (p Params) Validate() error {
	return nil
}

func noopValidator() subspace.ValueValidatorFn {
	return func(value interface{}) error { return nil }
}

func (m *Tally) Sub(vote Vote, weight sdk.Dec) error {
	if weight.LTE(sdk.ZeroDec()) {
		return errors.Wrap(ErrInvalid, "weight must be greater than 0")
	}
	switch vote.Choice {
	case Choice_YES:
		m.YesCount = m.YesCount.Sub(weight)
	case Choice_NO:
		m.NoCount = m.NoCount.Sub(weight)
	case Choice_ABSTAIN:
		m.AbstainCount = m.AbstainCount.Sub(weight)
	case Choice_VETO:
		m.VetoCount = m.VetoCount.Sub(weight)
	default:
		return errors.Wrapf(ErrInvalid, "unknown choice %s", vote.Choice.String())
	}
	return nil
}

func (m *Tally) Add(vote Vote, weight sdk.Dec) error {
	if weight.LTE(sdk.ZeroDec()) {
		return errors.Wrap(ErrInvalid, "weight must be greater than 0")
	}
	switch vote.Choice {
	case Choice_YES:
		m.YesCount = m.YesCount.Add(weight)
	case Choice_NO:
		m.NoCount = m.NoCount.Add(weight)
	case Choice_ABSTAIN:
		m.AbstainCount = m.AbstainCount.Add(weight)
	case Choice_VETO:
		m.VetoCount = m.VetoCount.Add(weight)
	default:
		return errors.Wrapf(ErrInvalid, "unknown choice %s", vote.Choice.String())
	}
	return nil
}

// TotalCounts is the sum of all weights.
func (m Tally) TotalCounts() sdk.Dec {
	return m.YesCount.Add(m.NoCount).Add(m.AbstainCount).Add(m.VetoCount)
}
