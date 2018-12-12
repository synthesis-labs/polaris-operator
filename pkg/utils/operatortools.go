package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HasFinalizer checks whether a particular finalizer is set
//
func HasFinalizer(instance *metav1.ObjectMeta, finalizer string) bool {
	for _, containedFinalizer := range instance.Finalizers {
		if containedFinalizer == finalizer {
			return true
		}
	}
	return false
}

// AddFinalizers adds a particular finalizer to an instance
//
func AddFinalizers(instance *metav1.ObjectMeta, finalizers ...string) {
	instance.SetFinalizers(append(instance.Finalizers, finalizers...))
}

// RemoveFinalizer removes a particular finalizer from an instance
//
func RemoveFinalizer(instance *metav1.ObjectMeta, finalizer string) {
	newFinalizers := []string{}
	for _, containedFinalizer := range instance.Finalizers {
		if containedFinalizer != finalizer {
			newFinalizers = append(newFinalizers, containedFinalizer)
		}
	}
	instance.SetFinalizers(newFinalizers)
}
