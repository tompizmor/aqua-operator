package common

import (
	"encoding/base64"

	util "github.com/aokoli/goutils"
	corev1 "k8s.io/api/core/v1"
)

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

func randAlphaNumeric(count int) string {
	// It is not possible, it appears, to actually generate an error here.
	r, _ := util.RandomAlphaNumeric(count)
	return r
}

func CreateRundomPassword() string {
	rand := randAlphaNumeric(20)
	pass := base64.StdEncoding.EncodeToString([]byte(rand))

	return pass
}
