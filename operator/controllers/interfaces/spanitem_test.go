package interfaces

import (
	"testing"

	"github.com/keptn/lifecycle-toolkit/operator/api/v1alpha2"
	apicommon "github.com/keptn/lifecycle-toolkit/operator/api/v1alpha2/common"
	"github.com/keptn/lifecycle-toolkit/operator/controllers/interfaces/fake"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSpanItemWrapper(t *testing.T) {
	evaluation := &v1alpha2.KeptnEvaluation{
		ObjectMeta: v1.ObjectMeta{
			Name: "evaluation",
		},
		Spec: v1alpha2.KeptnEvaluationSpec{
			AppName:    "app",
			AppVersion: "appversion",
			Type:       apicommon.PostDeploymentCheckType,
		},
		Status: v1alpha2.KeptnEvaluationStatus{
			OverallStatus: apicommon.StateFailed,
		},
	}

	object, err := NewSpanItemWrapperFromClientObject(evaluation)
	require.Nil(t, err)

	require.Equal(t, "evaluation", object.GetSpanKey(""))
}

func TestSpanItem(t *testing.T) {
	spanItemMock := fake.SpanItemMock{
		SetPhaseTraceIDFunc: func(phase string, carrier propagation.MapCarrier) {
		},
		SetSpanAttributesFunc: func(span trace.Span) {
		},
		GetSpanKeyFunc: func(phase string) string {
			return "key"
		},
		GetSpanNameFunc: func(phase string) string {
			return "name"
		},
	}

	wrapper := SpanItemWrapper{Obj: &spanItemMock}

	wrapper.SetPhaseTraceID("", nil)
	require.Len(t, spanItemMock.SetPhaseTraceIDCalls(), 1)

	wrapper.SetSpanAttributes(nil)
	require.Len(t, spanItemMock.SetSpanAttributesCalls(), 1)

	_ = wrapper.GetSpanKey("")
	require.Len(t, spanItemMock.GetSpanKeyCalls(), 1)

	wrapper.GetSpanName("")
	require.Len(t, spanItemMock.GetSpanNameCalls(), 1)

}
