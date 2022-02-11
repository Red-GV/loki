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
func TestStatefulSetHorizontalPodAutoscalerMatchName(t *testing.T) {
	type test struct {
		StatefulSet *appsv1.StatefulSet
		Autoscaler  *autoscalingv2beta2.HorizontalPodAutoscaler
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

	table := []test{
		{
			StatefulSet: NewIngesterStatefulSet(opt),
			Autoscaler:  NewIngesterHorizontalPodAutoscaler(opt),
		},
		{
			StatefulSet: NewIndexGatewayStatefulSet(opt),
			Autoscaler:  NewIndexGatewayHorizontalPodAutoscaler(opt),
		},
	}

	for _, tst := range table {
		testName := fmt.Sprintf("%s_%s", tst.StatefulSet.GetName(), tst.Autoscaler.GetName())
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tst.StatefulSet.GetName(), tst.Autoscaler.Spec.ScaleTargetRef.Name)
		})
	}
}

// Test that all the autoscalers have the same name as the deployment
func TestDeploymentHorizontalPodAutoscalerMatchName(t *testing.T) {
	type test struct {
		Deployment *appsv1.Deployment
		Autoscaler *autoscalingv2beta2.HorizontalPodAutoscaler
	}

	sha1C := "deadbeef"

	flags := FeatureFlags{
		EnableHorizontalAutoscaling: true,
		EnableGateway:               true,
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
				GatewayAutoscaling: &lokiv1beta1.AutoscalingSpec{
					HorizontalAutoscaling: &lokiv1beta1.HorizontalAutoscalingSpec{
						ScaleUpPercentage:   1,
						ScaleDownPercentage: 1,
					},
				},
			},
		},
	}

	table := []test{
		{
			Deployment: NewQuerierDeployment(opt),
			Autoscaler: NewQuerierHorizontalPodAutoscaler(opt),
		},
		{
			Deployment: NewQueryFrontendDeployment(opt),
			Autoscaler: NewQueryFrontendHorizontalPodAutoscaler(opt),
		},
		{
			Deployment: NewDistributorDeployment(opt),
			Autoscaler: NewDistributorHorizontalPodAutoscaler(opt),
		},
		{
			Deployment: NewGatewayDeployment(opt, sha1C),
			Autoscaler: NewGatewayHorizontalPodAutoscaler(opt),
		},
	}

	for _, tst := range table {
		testName := fmt.Sprintf("%s_%s", tst.Deployment.GetName(), tst.Autoscaler.GetName())
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tst.Deployment.GetName(), tst.Autoscaler.Spec.ScaleTargetRef.Name)
		})
	}
}
