package manifests

import (
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BuildHorizontalPodAutoscalers builds the horizontal pod autoscalers
func BuildHorizontalPodAutoscalers(opts Options) []client.Object {
	return []client.Object{
		NewIngesterHorizontalPodAutoscaler(opts),
		NewQuerierHorizontalPodAutoscaler(opts),
	}
}

// NewIngesterHorizontalPodAutoscaler creates a k8s autoscaler for the ingester stateful set
func NewIngesterHorizontalPodAutoscaler(opts Options) *autoscalingv2beta2.HorizontalPodAutoscaler {
	labels := ComponentLabels(LabelIngesterComponent, opts.Name)
	name := horizontalAutoscalerName(LabelIngesterComponent)
	replicas := opts.Stack.Template.Ingester.Replicas

	return newHorizontalPodAutoscaler(name, opts.Namespace, "StatefulSet", IngesterName(opts.Name), labels, replicas)
}

// NewQuerierHorizontalPodAutoscaler creates a k8s autoscaler for the querier deployment set
func NewQuerierHorizontalPodAutoscaler(opts Options) *autoscalingv2beta2.HorizontalPodAutoscaler {
	labels := ComponentLabels(LabelQuerierComponent, opts.Name)
	name := horizontalAutoscalerName(LabelQuerierComponent)
	replicas := opts.Stack.Template.Querier.Replicas

	return newHorizontalPodAutoscaler(name, opts.Namespace, "Deployment", QuerierName(opts.Name), labels, replicas)
}

func newHorizontalPodAutoscaler(name, namespace, targetKind, targetName string, labels labels.Set, replicas int32) *autoscalingv2beta2.HorizontalPodAutoscaler {
	return &autoscalingv2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: autoscalingv2beta2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
				Kind:       targetKind,
				Name:       targetName,
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
			MinReplicas: pointer.Int32Ptr(replicas),
			MaxReplicas: replicas * 4,
			Behavior: &autoscalingv2beta2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2beta2.HPAScalingRules{
					Policies: []autoscalingv2beta2.HPAScalingPolicy{
						{
							Type:          autoscalingv2beta2.PercentScalingPolicy,
							Value:         50,
							PeriodSeconds: 60,
						},
						{
							Type:          autoscalingv2beta2.PodsScalingPolicy,
							Value:         3,
							PeriodSeconds: 60,
						},
					},
					SelectPolicy:               policyPtr(autoscalingv2beta2.MinPolicySelect),
					StabilizationWindowSeconds: pointer.Int32Ptr(300),
				},
				ScaleUp: &autoscalingv2beta2.HPAScalingRules{
					Policies: []autoscalingv2beta2.HPAScalingPolicy{
						{
							Type:          autoscalingv2beta2.PercentScalingPolicy,
							Value:         100,
							PeriodSeconds: 15,
						},
						{
							Type:          autoscalingv2beta2.PodsScalingPolicy,
							Value:         3,
							PeriodSeconds: 15,
						},
					},
					SelectPolicy:               policyPtr(autoscalingv2beta2.MaxPolicySelect),
					StabilizationWindowSeconds: pointer.Int32Ptr(0),
				},
			},
			Metrics: []autoscalingv2beta2.MetricSpec{
				{
					Type: autoscalingv2beta2.ResourceMetricSourceType,
					Resource: &autoscalingv2beta2.ResourceMetricSource{
						Name: corev1.ResourceMemory,
						Target: autoscalingv2beta2.MetricTarget{
							Type:               autoscalingv2beta2.UtilizationMetricType,
							AverageUtilization: pointer.Int32Ptr(80),
						},
					},
				},
			},
		},
	}
}

func policyPtr(p autoscalingv2beta2.ScalingPolicySelect) *autoscalingv2beta2.ScalingPolicySelect {
	return &p
}
