package manifests

import (
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	l := ComponentLabels(LabelIngesterComponent, opts.Name)
	policy := autoscalingv2beta2.MinPolicySelect

	return &autoscalingv2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: autoscalingv2beta2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   horizontalAutoscalerName(LabelIngesterComponent),
			Labels: l,
		},
		Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
				Kind:       "StatefulSet",
				Name:       IngesterName(opts.Name),
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
			MinReplicas: pointer.Int32Ptr(opts.Stack.Template.Ingester.Replicas),
			MaxReplicas: opts.Stack.Template.Ingester.Replicas * 4,
			Behavior: &autoscalingv2beta2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2beta2.HPAScalingRules{
					Policies: []autoscalingv2beta2.HPAScalingPolicy{
						{
							Type:          "Percent",
							Value:         20,
							PeriodSeconds: 60,
						},
					},
					StabilizationWindowSeconds: pointer.Int32Ptr(300),
				},
				ScaleUp: &autoscalingv2beta2.HPAScalingRules{
					Policies: []autoscalingv2beta2.HPAScalingPolicy{
						{
							Type:          "Percent",
							Value:         100,
							PeriodSeconds: 15,
						},
						{
							Type:          "Pods",
							Value:         4,
							PeriodSeconds: 15,
						},
					},
					SelectPolicy:               &policy,
					StabilizationWindowSeconds: pointer.Int32Ptr(0),
				},
			},
			Metrics: []autoscalingv2beta2.MetricSpec{
				{
					Type: "Resource",
					Resource: &autoscalingv2beta2.ResourceMetricSource{
						Name: "Memory",
						Target: autoscalingv2beta2.MetricTarget{
							Type:               "Utilization",
							AverageUtilization: pointer.Int32Ptr(80),
						},
					},
				},
			},
		},
	}
}

// NewQuerierHorizontalPodAutoscaler creates a k8s autoscaler for the querier deployment set
func NewQuerierHorizontalPodAutoscaler(opts Options) *autoscalingv2beta2.HorizontalPodAutoscaler {
	l := ComponentLabels(LabelQuerierComponent, opts.Name)
	policy := autoscalingv2beta2.MinPolicySelect

	return &autoscalingv2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: autoscalingv2beta2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   horizontalAutoscalerName(LabelQuerierComponent),
			Labels: l,
		},
		Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       QuerierName(opts.Name),
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
			MinReplicas: pointer.Int32Ptr(opts.Stack.Template.Querier.Replicas),
			MaxReplicas: opts.Stack.Template.Querier.Replicas * 4,
			Behavior: &autoscalingv2beta2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2beta2.HPAScalingRules{
					Policies: []autoscalingv2beta2.HPAScalingPolicy{
						{
							Type:          "Percent",
							Value:         20,
							PeriodSeconds: 60,
						},
					},
					StabilizationWindowSeconds: pointer.Int32Ptr(300),
				},
				ScaleUp: &autoscalingv2beta2.HPAScalingRules{
					Policies: []autoscalingv2beta2.HPAScalingPolicy{
						{
							Type:          "Percent",
							Value:         100,
							PeriodSeconds: 15,
						},
						{
							Type:          "Pods",
							Value:         4,
							PeriodSeconds: 15,
						},
					},
					SelectPolicy:               &policy,
					StabilizationWindowSeconds: pointer.Int32Ptr(0),
				},
			},
			Metrics: []autoscalingv2beta2.MetricSpec{
				{
					Type: "Resource",
					Resource: &autoscalingv2beta2.ResourceMetricSource{
						Name: "Memory",
						Target: autoscalingv2beta2.MetricTarget{
							Type:               "Utilization",
							AverageUtilization: pointer.Int32Ptr(80),
						},
					},
				},
			},
		},
	}
}
