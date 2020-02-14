# Group Module

## Group

A group is simply an aggregation of accounts with associated weights. It is not
an account and doesn't have a balance. It doesn't in and of itself have any
sort of voting or decision power. It does have an "administrator" which has
the power to add, remove and update members in the group. Note that a
group account could be an administrator of a group.

## Group Account

A group account is an account associated with a group and a decision policy.

## Decision Policy

A decision policy is the mechanism by which members of a group can vote on 
proposals.

All decision policies generally would have a minimum and maximum voting window.
The minimum voting window is the minimum amount of time that must pass in order
for a proposal to potentially pass, and it may be set to 0. The maximum voting
window is the maximum time that a proposal may be voted on before it is closed.
Both of these values must be less than a chain-wide max voting window parameter.

### Threshold decision policy

A threshold decision policy defines a threshold of yes votes that must be achieved
in order to pass a proposal. For this decision policy, abstain and veto are
simply treated as no's.

## Proposal

Any account can submit a proposal for a group account to decide upon. A proposal
consists of a set of `sdk.Msg`s that will be executed if the proposal passes
as well as any comment associated with the proposal.

## Voting

There are four choices to choose while voting - yes, no, abstain and veto. Not
all decision policies will support them. Votes can contain an optional comment.

## Executing Proposals

Proposals will not be automatically executed by the chain in this current design,
but rather a user must submit a `MsgExec` transaction to attempt to execute the
proposal based on the current votes and decision policy. A future upgrade could
automate this propose and have the group account (or a fee granter) pay.

## Changing Group Membership

In the current implementation, changing a group's membership (adding or removing members or changing their power)
will cause all existing proposals for group accounts linked to this group
to be invalidated. They will simply fail if someone calls `MsgExec` and will
eventually be garbage collected.
