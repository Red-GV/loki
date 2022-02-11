package manifests

import (
	lokiv1beta1 "github.com/grafana/loki/operator/api/v1beta1"
	"github.com/grafana/loki/operator/internal/manifests/internal"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	smallestSize = lokiv1beta1.SizeOneXSmall
	largestSize  = lokiv1beta1.SizeOneXMedium
)

type autoscalerBuilder struct {
	name      string
	namespace string
	labelSet  labels.Set
	objectRef autoscalingv2beta2.CrossVersionObjectReference
	min       int32
	max       int32
	config    *lokiv1beta1.HorizontalAutoscalingSpec
}

// BuildHorizontalPodAutoscalers builds the horizontal pod autoscalers
func BuildHorizontalPodAutoscalers(opts Options) []client.Object {
	objects := []client.Object{}

	if opts.Stack.Autoscaling.IngestionAutoscaling.HorizontalAutoscaling != nil {
		objects = append(objects, NewIngesterHorizontalPodAutoscaler(opts))
	}

	if opts.Stack.Autoscaling.QueryAutoscaling.HorizontalAutoscaling != nil {
		objects = append(objects, NewQuerierHorizontalPodAutoscaler(opts))
	}

	return objects
}

// NewQuerierHorizontalPodAutoscaler creates a k8s autoscaler for the querier deployment
func NewQuerierHorizontalPodAutoscaler(opts Options) *autoscalingv2beta2.HorizontalPodAutoscaler {
	b := autoscalerBuilder{
		name:      horizontalAutoscalerName(LabelQuerierComponent),
		namespace: opts.Namespace,
		labelSet:  ComponentLabels(LabelQuerierComponent, opts.Name),
		objectRef: newDeploymentCrossVersionObjectReference(QuerierName(opts.Name)),
		min:       internal.StackSizeTable[smallestSize].Template.Querier.Replicas,
		max:       internal.StackSizeTable[largestSize].Template.Querier.Replicas,
		config:    opts.Stack.Autoscaling.QueryAutoscaling.HorizontalAutoscaling,
	}

	return newHorizontalPodAutoscaler(b)
}

// NewIngesterHorizontalPodAutoscaler creates a k8s autoscaler for the ingester stateful set
func NewIngesterHorizontalPodAutoscaler(opts Options) *autoscalingv2beta2.HorizontalPodAutoscaler {
	b := autoscalerBuilder{
		name:      horizontalAutoscalerName(LabelIngesterComponent),
		namespace: opts.Namespace,
		labelSet:  ComponentLabels(LabelIngesterComponent, opts.Name),
		objectRef: newStatefulSetCrossVersionObjectReference(IngesterName(opts.Name)),
		min:       internal.StackSizeTable[smallestSize].Template.Ingester.Replicas,
		max:       internal.StackSizeTable[largestSize].Template.Ingester.Replicas,
		config:    opts.Stack.Autoscaling.IngestionAutoscaling.HorizontalAutoscaling,
	}

	return newHorizontalPodAutoscaler(b)
}

func policyPtr(p autoscalingv2beta2.ScalingPolicySelect) *autoscalingv2beta2.ScalingPolicySelect {
	return &p
}

func newDeploymentCrossVersionObjectReference(name string) autoscalingv2beta2.CrossVersionObjectReference {
	return autoscalingv2beta2.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       name,
		APIVersion: appsv1.SchemeGroupVersion.String(),
	}
}

func newStatefulSetCrossVersionObjectReference(name string) autoscalingv2beta2.CrossVersionObjectReference {
	return autoscalingv2beta2.CrossVersionObjectReference{
		Kind:       "StatefulSet",
		Name:       name,
		APIVersion: appsv1.SchemeGroupVersion.String(),
	}
}

func newHorizontalPodAutoscaler(b autoscalerBuilder) *autoscalingv2beta2.HorizontalPodAutoscaler {
	return &autoscalingv2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: autoscalingv2beta2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name,
			Namespace: b.namespace,
			Labels:    b.labelSet,
		},
		Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: b.objectRef,
			MinReplicas:    pointer.Int32Ptr(b.min),
			MaxReplicas:    b.max,
			Behavior: &autoscalingv2beta2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2beta2.HPAScalingRules{
					Policies: []autoscalingv2beta2.HPAScalingPolicy{
						{
							Type:          autoscalingv2beta2.PercentScalingPolicy,
							Value:         b.config.ScaleDownPercentage,
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
							Value:         b.config.ScaleUpPercentage,
							PeriodSeconds: 15,
						},
					},
					SelectPolicy:               policyPtr(autoscalingv2beta2.MaxPolicySelect),
					StabilizationWindowSeconds: pointer.Int32Ptr(30),
				},
			},
			Metrics: []autoscalingv2beta2.MetricSpec{
				{
					Type: autoscalingv2beta2.ResourceMetricSourceType,
					Resource: &autoscalingv2beta2.ResourceMetricSource{
						Name: corev1.ResourceMemory,
						Target: autoscalingv2beta2.MetricTarget{
							Type:               autoscalingv2beta2.UtilizationMetricType,
							AverageUtilization: pointer.Int32Ptr(70),
						},
					},
				},
			},
		},
	}
}
