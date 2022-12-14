/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"knative.dev/pkg/apis"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	floopowderv1alpha1 "github.com/floo-powder/pipeline/api/v1alpha1"
	"github.com/floo-powder/pkg/common"
)

// SourceReconciler reconciles a Source object
type SourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=floo-powder.dev,resources=sources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=floo-powder.dev,resources=sources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=floo-powder.dev,resources=sources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Source object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var obj floopowderv1alpha1.Source
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		logger.Error(err, "get request source failed")
		return ctrl.Result{}, err
	}
	patch := client.MergeFrom(obj.DeepCopy())
	if obj.Status.Status == nil {
		obj.Status.Status = &duckv1.Status{}
	}
	for _, v := range obj.Status.Conditions {
		if v.Type == apis.ConditionReady && v.Status == corev1.ConditionTrue {
			return ctrl.Result{}, nil
		}
	}
	var pluginObj floopowderv1alpha1.Plugin
	condition := apis.Condition{
		Type:   apis.ConditionReady,
		Status: corev1.ConditionFalse,
	}
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Spec.Plugin}, &pluginObj); err != nil {
		condition.Reason = string(common.NotFound)
		condition.Message = "Not Found Plugin"
		obj.Status.SetConditions(apis.Conditions{condition})
		err = r.Patch(ctx, &obj, patch)
		return ctrl.Result{}, err
	}
	condition.Reason, condition.Message = obj.Spec.ValidateConfig(pluginObj)
	if condition.Reason == string(common.Succeed) {
		condition.Status = corev1.ConditionTrue
	}
	err := r.Patch(ctx, &obj, patch)
	if err != nil {
		logger.Error(err, "patch source status failed")
	}
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *SourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&floopowderv1alpha1.Source{}).
		Complete(r)
}
