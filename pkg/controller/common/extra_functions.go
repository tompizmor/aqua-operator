package common

import corev1 "k8s.io/api/core/v1"

func Int32Ptr(i int32) *int32 {
	return &i
}

func GetPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
