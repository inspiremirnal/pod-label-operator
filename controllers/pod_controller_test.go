package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

const (
	PodNamePrefix = "test-pod-"
	PodNamespace  = "default"
	timeout       = time.Second * 3
	duration      = time.Second * 3
	interval      = time.Millisecond * 250
)

var _ = Describe("Pod Controller", func() {
	ctx := context.Background()
	Context("When a Pod has annotation but not the label", func() {
		It("Should add the label", func() {
			podName := PodNamePrefix + "with-annotation-without-label"
			createPod(ctx, podName, true, false)
			validatePod(ctx, podName, true)
		})
	})

	Context("When a Pod does not have the annotation but has the label", func() {
		It("Should remove the label", func() {
			PodName := PodNamePrefix + "without-annotation-with-label"
			createPod(ctx, PodName, false, true)
			validatePod(ctx, PodName, false)
		})
	})

	Context("When a Pod has the annotation and the label", func() {
		It("Should keep the label", func() {
			PodName := PodNamePrefix + "with-annotation-with-label"
			createPod(ctx, PodName, true, true)
			validatePod(ctx, PodName, true)
		})
	})

	Context("When a Pod has neither the annotation nor the label", func() {
		It("Should not add the label", func() {
			PodName := PodNamePrefix + "without-annotation-without-label"
			createPod(ctx, PodName, false, false)
			validatePod(ctx, PodName, false)
		})
	})

})

func createPod(ctx context.Context, name string, withAnnotation, withLabel bool) {
	By("Creating a new Pod")
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: PodNamespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
	if withAnnotation {
		pod.Annotations = map[string]string{
			addPodNameLabelAnnotation: "true",
		}
	}

	if withLabel {
		pod.Labels = map[string]string{
			podNameLabel: name,
		}
	}
	Expect(k8sClient.Create(ctx, pod)).Should(Succeed())
}

func validatePod(ctx context.Context, name string, shouldHaveLabel bool) {
	By("Checking the pod has or does not have the label")
	podIsValid := func() bool {
		podLookupKey := types.NamespacedName{
			Name: name, Namespace: PodNamespace,
		}
		var createPod corev1.Pod
		err := k8sClient.Get(ctx, podLookupKey, &createPod)
		if err != nil {
			return false
		}

		if shouldHaveLabel {
			return (createPod.Labels[podNameLabel] == name) == shouldHaveLabel
		}

		_, hasLabel := createPod.Labels[podNameLabel]
		return !hasLabel
	}

	Eventually(podIsValid, timeout, interval).Should(BeTrue())
	Consistently(podIsValid, duration, interval).Should(BeTrue())
}
