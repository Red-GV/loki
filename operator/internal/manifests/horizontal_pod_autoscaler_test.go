package manifests

import (
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"

	lokiv1beta1 "github.com/grafana/loki/operator/api/v1beta1"
	"github.com/stretchr/testify/assert"
)

// Test that all the autoscalers have the same name as the statefulsets
func TestHorizontalPodAutoscalerObjectMatchName(t *testing.T) {
	type stsTest struct {
		StatefulSet *appsv1.StatefulSet
		Autoscaler  *autoscalingv2beta2.HorizontalPodAutoscaler
	}

	type dplTest struct {
		Deployment *appsv1.Deployment
		Autoscaler *autoscalingv2beta2.HorizontalPodAutoscaler
	}

	flags := FeatureFlags{
		EnableHorizontalAutoscaling: true,
	}

	opt := Options{
		Name:      "test",
		Namespace: "test",
		Image:     "test",
		Flags:     flags,
		Stack: lokiv1beta1.LokiStackSpec{
			Size: lokiv1beta1.SizeOneXExtraSmall,
			Template: &lokiv1beta1.LokiTemplateSpec{
				Compactor: &lokiv1beta1.LokiComponentSpec{
					Replicas: 1,
				},
				Distributor: &lokiv1beta1.LokiComponentSpec{
					Replicas: 1,
				},
				Ingester: &lokiv1beta1.LokiComponentSpec{
					Replicas: 1,
				},
				Querier: &lokiv1beta1.LokiComponentSpec{
					Replicas: 1,
				},
				QueryFrontend: &lokiv1beta1.LokiComponentSpec{
					Replicas: 1,
				},
				Gateway: &lokiv1beta1.LokiComponentSpec{
					Replicas: 1,
				},
				IndexGateway: &lokiv1beta1.LokiComponentSpec{
					Replicas: 1,
				},
			},
			Autoscaling: &lokiv1beta1.AutoscalingTemplateSpec{
				IngestionAutoscaling: &lokiv1beta1.AutoscalingSpec{
					HorizontalAutoscaling: &lokiv1beta1.HorizontalAutoscalingSpec{
						ScaleUpPercentage:   1,
						ScaleDownPercentage: 1,
					},
				},
				QueryAutoscaling: &lokiv1beta1.AutoscalingSpec{
					HorizontalAutoscaling: &lokiv1beta1.HorizontalAutoscalingSpec{
						ScaleUpPercentage:   1,
						ScaleDownPercentage: 1,
					},
				},
			},
		},
	}

	stsTable := []stsTest{
		{
			StatefulSet: NewIngesterStatefulSet(opt),
			Autoscaler:  NewIngesterHorizontalPodAutoscaler(opt),
		},
	}

	dplTable := []dplTest{
		{
			Deployment: NewQuerierDeployment(opt),
			Autoscaler: NewQuerierHorizontalPodAutoscaler(opt),
		},
	}

	for _, tst := range stsTable {
		testName := fmt.Sprintf("%s_%s", tst.StatefulSet.GetName(), tst.Autoscaler.GetName())
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tst.StatefulSet.GetName(), tst.Autoscaler.Spec.ScaleTargetRef.Name)
		})
	}

	for _, tst := range dplTable {
		testName := fmt.Sprintf("%s_%s", tst.Deployment.GetName(), tst.Autoscaler.GetName())
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tst.Deployment.GetName(), tst.Autoscaler.Spec.ScaleTargetRef.Name)
		})
	}
}
