/*
Copyright 2023.

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
	"fmt"

	//appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	//import package for record
	"k8s.io/client-go/tools/record"

	greetv1alpha1 "github.com/thirumurthis/app-operator/api/v1alpha1"
)

// GreetReconciler reconciles a Greet object
// added Recorder to the struct
type GreetReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=greet.greetapp.com,resources=greets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=greet.greetapp.com,resources=greets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=greet.greetapp.com,resources=greets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Greet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *GreetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	log.Log.Info("Reconciler invoked..")
	instance := &greetv1alpha1.Greet{}

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		r.Recorder.Event(instance, corev1.EventTypeWarning, "Object", "Failed to read Object")
		log.Log.Info("Error while reading the object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	appName := instance.Spec.Name

	if instance.Spec.Name != "" {
		log.Log.Info(fmt.Sprintf("appName for CRD is - %s ", instance.Spec.Name))
		r.Recorder.Event(instance, corev1.EventTypeWarning, "Greet", fmt.Sprintf("Created - %s ", appName))
	} else {
		log.Log.Info("instance.Spec.Name - NOT FOUND")
	}

	if instance.Status.Status == "" {
		instance.Status.Status = "OK"
		log.Log.Info("instance.Spec.Name - is set to OK")
	}

	if err := r.Status().Update(ctx, instance); err != nil {
		log.Log.Info("Error while reading the object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GreetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&greetv1alpha1.Greet{}).
		//Owns(&appsv1.Deployment{}).
		Complete(r)
}
