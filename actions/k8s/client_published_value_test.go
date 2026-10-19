package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// rootActionID addresses the "run1-a0" TaskAction newConditionClient preloads.
var rootActionID = &common.ActionIdentifier{
	Run:  &common.RunIdentifier{Project: "proj", Domain: "dev", Name: "run1"},
	Name: "run1",
}

func getPublished(t *testing.T, c *ActionsClient) *executorv1.TaskAction {
	t.Helper()
	ta := &executorv1.TaskAction{}
	require.NoError(t, c.k8sClient.Get(context.Background(), client.ObjectKey{Name: "run1-a0", Namespace: "flyte"}, ta))
	return ta
}

func running(attempts uint32) *workflow.ActionStatus {
	return &workflow.ActionStatus{Phase: common.ActionPhase_ACTION_PHASE_RUNNING, Attempts: attempts}
}

func TestPutStatus_PublishedValue(t *testing.T) {
	ctx := context.Background()
	c := newConditionClient(t)
	addr := strLiteral("http://agent:8080/s/run1")

	require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), addr))
	ta := getPublished(t, c)
	require.True(t, proto.Equal(addr, PublishedValueFromStatus(ctx, ta)), "value stored on the CR status")

	// No value in the same attempt keeps it.
	require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), nil))
	require.True(t, proto.Equal(addr, PublishedValueFromStatus(ctx, getPublished(t, c))))

	// A new value replaces it.
	other := strLiteral("http://agent:8080/s/run1-b")
	require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), other))
	require.True(t, proto.Equal(other, PublishedValueFromStatus(ctx, getPublished(t, c))))

	// A new attempt clears it.
	require.NoError(t, c.PutStatus(ctx, rootActionID, 2, running(2), nil))
	assert.Empty(t, getPublished(t, c).Status.PublishedValue)
}

func TestPutStatus_PublishedValueDoesNotTouchSignalValue(t *testing.T) {
	ctx := context.Background()
	c := newConditionClient(t)
	require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), strLiteral("x")))
	assert.Empty(t, getPublished(t, c).Status.SignalValue, "a published value must never complete the action")
}

func TestBuildActionUpdate_PublishedValue(t *testing.T) {
	ctx := context.Background()
	ta, _ := newTestActionUpdate("session1")
	assert.Nil(t, buildActionUpdate(ctx, ta, watch.Modified).Value())

	raw, err := proto.Marshal(strLiteral("http://agent:8080/s/session1"))
	require.NoError(t, err)
	ta.Status.PublishedValue = raw
	ta.Status.Conditions = []metav1.Condition{{
		Type: string(executorv1.ConditionTypeProgressing), Status: "True", Reason: string(executorv1.ConditionReasonExecuting),
	}}
	update := buildActionUpdate(ctx, ta, watch.Modified)
	require.NotNil(t, update.Value(), "a running task's published value reaches watchers")
	assert.Equal(t, "http://agent:8080/s/session1", update.Value().GetScalar().GetPrimitive().GetStringValue())
	assert.Nil(t, update.SignalValue)

	// Once terminal, the task's published value is no longer delivered.
	ta.Status.Conditions = []metav1.Condition{{Type: string(executorv1.ConditionTypeSucceeded), Status: "True"}}
	assert.Nil(t, buildActionUpdate(ctx, ta, watch.Modified).Value())
}

func TestNextPublishedValue(t *testing.T) {
	a, err := proto.MarshalOptions{Deterministic: true}.Marshal(strLiteral("a"))
	require.NoError(t, err)

	got, err := nextPublishedValue(nil, 1, 1, strLiteral("a"))
	require.NoError(t, err)
	assert.Equal(t, a, got, "set")

	got, err = nextPublishedValue(a, 1, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, a, got, "kept within the attempt")

	got, err = nextPublishedValue(a, 1, 2, nil)
	require.NoError(t, err)
	assert.Nil(t, got, "cleared by a new attempt")
}
