//go:build unit || !integration

package scheduler

import (
	"testing"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestExecSet_FilterNonTerminal(t *testing.T) {
	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", State: model.ExecutionStateBidAccepted},
		{ComputeReference: "exec2", State: model.ExecutionStateCompleted},
		{ComputeReference: "exec3", State: model.ExecutionStateFailed},
	}

	set := execSetFromSlice(executions)
	filtered := set.filterNonTerminal()
	assert.ElementsMatch(t, filtered.keys(), []string{"exec1"})
}

func TestExecSet_FilterByState(t *testing.T) {
	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", State: model.ExecutionStateBidAccepted},
		{ComputeReference: "exec2", State: model.ExecutionStateFailed},
		{ComputeReference: "exec3", State: model.ExecutionStateBidAccepted},
		{ComputeReference: "exec4", State: model.ExecutionStateCompleted},
	}

	set := execSetFromSlice(executions)

	filtered1 := set.filterByState(model.ExecutionStateBidAccepted)
	assert.Len(t, filtered1, 2)
	assert.ElementsMatch(t, filtered1.keys(), []string{"exec1", "exec3"})

	filtered2 := set.filterByState(model.ExecutionStateFailed)
	assert.Len(t, filtered2, 1)
	assert.ElementsMatch(t, filtered2.keys(), []string{"exec2"})

	filtered3 := set.filterByState(model.ExecutionStateCompleted)
	assert.Len(t, filtered3, 1)
	assert.ElementsMatch(t, filtered3.keys(), []string{"exec4"})
}

func TestExecSet_FilterRunning(t *testing.T) {
	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", State: model.ExecutionStateBidAccepted},
		{ComputeReference: "exec2", State: model.ExecutionStateAskForBidAccepted},
		{ComputeReference: "exec3", State: model.ExecutionStateCompleted},
	}

	set := execSetFromSlice(executions)
	filtered := set.filterRunning()

	assert.Len(t, filtered, 1)
	assert.ElementsMatch(t, filtered.keys(), []string{"exec1"})
}

func TestExecSet_FilterFailed(t *testing.T) {
	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", State: model.ExecutionStateFailed},
		{ComputeReference: "exec2", State: model.ExecutionStateCompleted},
		{ComputeReference: "exec3", State: model.ExecutionStateFailed},
	}

	set := execSetFromSlice(executions)
	filtered := set.filterFailed()

	assert.Len(t, filtered, 2)
	assert.ElementsMatch(t, filtered.keys(), []string{"exec1", "exec3"})
}

func TestExecSet_Union(t *testing.T) {
	set1 := execSet{
		"exec1": {ComputeReference: "exec1", State: model.ExecutionStateAskForBid},
		"exec2": {ComputeReference: "exec2", State: model.ExecutionStateBidAccepted},
	}

	set2 := execSet{
		"exec2": {ComputeReference: "exec2", State: model.ExecutionStateCompleted},
		"exec3": {ComputeReference: "exec3", State: model.ExecutionStateBidAccepted},
	}

	union := set1.union(set2)

	assert.Len(t, union, 3)
	assert.ElementsMatch(t, union.keys(), []string{"exec1", "exec2", "exec3"})

	// verify exec2 of the second set is the one that is kept
	assert.Equal(t, model.ExecutionStateCompleted, union["exec2"].State)
}

func TestExecSet_Latest(t *testing.T) {
	now := time.Now()
	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", UpdateTime: now},
		{ComputeReference: "exec2", UpdateTime: now.Add(+1 * time.Second)},
		{ComputeReference: "exec3", UpdateTime: now.Add(-1 * time.Second)},
	}

	set := execSetFromSlice(executions)
	latest := set.latest()

	assert.NotNil(t, latest)
	assert.Equal(t, "exec2", latest.ComputeReference)
}

func TestExecSet_CountByState(t *testing.T) {
	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", State: model.ExecutionStateBidAccepted},
		{ComputeReference: "exec2", State: model.ExecutionStateFailed},
		{ComputeReference: "exec3", State: model.ExecutionStateBidAccepted},
		{ComputeReference: "exec4", State: model.ExecutionStateCompleted},
	}

	set := execSetFromSlice(executions)
	counts := set.countByState()

	assert.Equal(t, 2, counts[model.ExecutionStateBidAccepted])
	assert.Equal(t, 1, counts[model.ExecutionStateFailed])
	assert.Equal(t, 1, counts[model.ExecutionStateCompleted])
}

func TestExecSet_CountCompleted(t *testing.T) {
	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", State: model.ExecutionStateBidAccepted},
		{ComputeReference: "exec2", State: model.ExecutionStateFailed},
		{ComputeReference: "exec3", State: model.ExecutionStateCompleted},
		{ComputeReference: "exec4", State: model.ExecutionStateCompleted},
	}

	set := execSetFromSlice(executions)
	count := set.countCompleted()

	assert.Equal(t, 2, count)
}

func TestExecSet_String(t *testing.T) {
	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", State: model.ExecutionStateBidAccepted},
		{ComputeReference: "exec2", State: model.ExecutionStateFailed},
		{ComputeReference: "exec3", State: model.ExecutionStateCompleted},
	}

	set := execSetFromSlice(executions)
	str := set.String()

	assert.Contains(t, str, "len(3)")
	assert.Contains(t, str, `"exec1":`)
	assert.Contains(t, str, `"exec2":`)
	assert.Contains(t, str, `"exec3":`)
}

func TestExecSet_has(t *testing.T) {
	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", State: model.ExecutionStateBidAccepted},
		{ComputeReference: "exec2", State: model.ExecutionStateFailed},
	}

	set := execSetFromSlice(executions)

	assert.True(t, set.has("exec1"))
	assert.True(t, set.has("exec2"))
	assert.False(t, set.has("exec3"))
}

func TestExecSet_FilterByNodeHealth(t *testing.T) {
	nodeInfos := map[string]*model.NodeInfo{
		"node1": {},
		"node2": {},
	}

	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", NodeID: "node1"},
		{ComputeReference: "exec2", NodeID: "node2"},
		{ComputeReference: "exec3", NodeID: "node3"},
	}

	set := execSetFromSlice(executions)
	healthy, lost := set.filterByNodeHealth(nodeInfos)

	assert.Len(t, healthy, 2)
	assert.Len(t, lost, 1)
	assert.ElementsMatch(t, healthy.keys(), []string{"exec1", "exec2"})
	assert.ElementsMatch(t, lost.keys(), []string{"exec3"})
}

func TestExecSet_FilterByOverSubscription(t *testing.T) {
	desiredCount := 3
	now := time.Now()

	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", UpdateTime: now},
		{ComputeReference: "exec2", UpdateTime: now.Add(time.Second)},
		{ComputeReference: "exec3", UpdateTime: now.Add(2 * time.Second)},
		{ComputeReference: "exec4", UpdateTime: now.Add(3 * time.Second)},
		{ComputeReference: "exec5", UpdateTime: now.Add(4 * time.Second)},
	}

	set := execSetFromSlice(executions)
	remaining, overSubscriptions := set.filterByOverSubscriptions(desiredCount)

	assert.ElementsMatch(t, remaining.keys(), []string{"exec1", "exec2", "exec3"})
	assert.ElementsMatch(t, overSubscriptions.keys(), []string{"exec4", "exec5"})
}

func TestExecSet_FilterByApprovalStatus(t *testing.T) {
	desiredCount := 3
	now := time.Now()

	executions := []*model.ExecutionState{
		{ComputeReference: "exec1", State: model.ExecutionStateAskForBidAccepted, UpdateTime: now},
		{ComputeReference: "exec2", State: model.ExecutionStateAskForBidAccepted, UpdateTime: now.Add(time.Second)},
		{ComputeReference: "exec3", State: model.ExecutionStateBidAccepted, UpdateTime: now.Add(2 * time.Second)},
		{ComputeReference: "exec4", State: model.ExecutionStateBidAccepted, UpdateTime: now.Add(3 * time.Second)},
		{ComputeReference: "exec5", State: model.ExecutionStateCompleted, UpdateTime: now.Add(4 * time.Second)},
	}

	set := execSetFromSlice(executions)
	approvalStatus := set.filterByApprovalStatus(desiredCount)

	assert.ElementsMatch(t, approvalStatus.running.keys(), []string{"exec3", "exec4"})
	assert.ElementsMatch(t, approvalStatus.toApprove.keys(), []string{"exec1"})
	assert.ElementsMatch(t, approvalStatus.toReject.keys(), []string{"exec2"})
	assert.ElementsMatch(t, approvalStatus.pending.keys(), []string{})
	assert.Equal(t, 3, approvalStatus.activeCount())
}
