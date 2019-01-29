package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("utils")

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
	for _, finalizer := range finalizers {
		if !HasFinalizer(instance, finalizer) {
			instance.SetFinalizers(append(instance.Finalizers, finalizer))
		}
	}
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
