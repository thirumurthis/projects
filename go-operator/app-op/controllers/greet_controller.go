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
	"time"

	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	//"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
//+kubebuilder:rbac:groups=greet.greetapp.com,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=greet.greetapp.com,resources=pods,verbs=get;list;create;update;patch

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

	checkAndCreateConfigMapResource(r, instance, ctx)
	checkAndCreateServiceResource(r, instance, ctx)
	checkAndCreateDeploymentResource(r, instance, ctx)

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
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func checkAndCreateConfigMapResource(r *GreetReconciler, instance *greetv1alpha1.Greet,
	ctx context.Context) (ctrl.Result, error) {

	fileName := instance.Spec.Deployment.ConfigMap.FileName
	data := make(map[string]string)
	content := instance.Spec.Deployment.ConfigMap.Data
	data[fileName] = content
	identifiedConfigMap := &corev1.ConfigMap{}
	configMapName := instance.Spec.Deployment.ConfigMap.Name + "-cfg"
	log.Log.Info(fmt.Sprintf("data in config %v", data))
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: instance.Namespace,
		},
		Data: data,
	}
	if err := r.Get(ctx,
		types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
		identifiedConfigMap); err != nil && errors.IsNotFound(err) {
		log.Log.Info("Creating ConfigMap", "ConfigMap", configMapName)
		// Error occurred while creating the ConfigMap
		//r.Log.Info("Creating ConfigMap", "ConfigMap", configMap)
		if err := r.Create(ctx, configMap); err != nil {
			// Error occurred while creating the ConfigMap
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else {
		log.Log.Info(fmt.Sprintf("Updating ConfigMap %v", configMap))
		if err := r.Update(ctx, configMap); err != nil {
			// Error occurred while updating the ConfigMap
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Duration(60 * time.Second)}, err
	}
}

func checkAndCreateServiceResource(r *GreetReconciler, instance *greetv1alpha1.Greet,
	ctx context.Context) (ctrl.Result, error) {

	labels := make(map[string]string)
	labels["name"] = instance.Spec.Name

	serviceSpec := instance.Spec.Deployment.Service.Spec
	identifiedService := &corev1.Service{}
	name := instance.Spec.Deployment.Service.Name + "-svc"
	log.Log.Info(fmt.Sprintf("data in service %v", serviceSpec))
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: serviceSpec,
	}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
		identifiedService); err != nil && errors.IsNotFound(err) {
		log.Log.Info("Creating Service", "Service", serviceSpec)
		// Error occurred while creating the ConfigMap
		//r.Log.Info("Creating ConfigMap", "ConfigMap", configMap)
		if err := r.Create(ctx, service); err != nil {
			// Error occurred while creating the ConfigMap
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else {
		log.Log.Info(fmt.Sprintf("Updating service %v", service))
		if err := r.Update(ctx, service); err != nil {
			// Error occurred while updating the ConfigMap
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Duration(60 * time.Second)}, err
	}
}

func checkAndCreateDeploymentResource(r *GreetReconciler, instance *greetv1alpha1.Greet,
	ctx context.Context) (ctrl.Result, error) {

	labels := make(map[string]string)
	labels["name"] = instance.Spec.Name

	deploymentName := instance.Spec.Deployment.Name + "-deploy"
	replicas := instance.Spec.Deployment.Replicas
	imageUrl := instance.Spec.Deployment.Pod.Image
	imagePullPolicy := instance.Spec.Deployment.Pod.ImagePullPolicy
	mountName := instance.Spec.Deployment.Pod.MountName
	mountPath := instance.Spec.Deployment.Pod.MountPath
	port := instance.Spec.Deployment.Pod.PodPort

	configMapName := instance.Spec.Deployment.ConfigMap.Name + "-cfg"

	identifiedDeployment := &appsv1.Deployment{}

	log.Log.Info(fmt.Sprintf("Deployment - %s %s %s %d", deploymentName, imageUrl, imagePullPolicy, replicas))

	volumeName := mountName + "-vol"

	var containers []corev1.Container
	var policyNameType corev1.PullPolicy

	container := corev1.Container{}
	container.Image = imageUrl
	var containerPorts []corev1.ContainerPort
	cPort := corev1.ContainerPort{}

	if port == 0 {
		port = 80
	}
	cPort.ContainerPort = port
	var portName string
	portName = instance.Spec.Deployment.Name
	if len(instance.Spec.Deployment.Name) > 10 {
		portName = instance.Spec.Deployment.Name[:10]
	}
	cPort.Name = portName + "-port"
	containerPorts = append(containerPorts, cPort)
	container.Ports = containerPorts
	container.Name = instance.Spec.Deployment.Name + "-pod"
	//pod command
	container.Command = instance.Spec.Deployment.Pod.Command
	container.Args = instance.Spec.Deployment.Pod.Args

	if imagePullPolicy == "" || len(imagePullPolicy) == 0 {
		policyNameType = corev1.PullAlways
	}
	if imagePullPolicy == "Always" {
		policyNameType = corev1.PullAlways
	}
	if imagePullPolicy == "IfNotPresent" {
		policyNameType = corev1.PullIfNotPresent
	}

	container.ImagePullPolicy = policyNameType
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
	})

	containers = append(containers, container)
	log.Log.Info(fmt.Sprintf("Container - %#v", containers))
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	}

	var volumes []corev1.Volume

	volumes = append(volumes, volume)

	deployer := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: containers,
					Volumes:    volumes,
				},
			},
		},
	}

	// Used to deserializer
	deployUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(deployer)

	if err != nil {
		log.Log.Error(err, "Error occurred in unstructuring")
	}

	encoder, err := yaml.Marshal(deployUnstructured)
	if err != nil {
		log.Log.Error(err, "Error occurred in transforming")
	}
	//prints the yaml format of the deployment object created
	log.Log.Info(fmt.Sprintf("%#v", string(encoder)))

	log.Log.Info(fmt.Sprintf("yaml: \n%#v\n", deployer))
	if err := r.Get(ctx,
		types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
		identifiedDeployment); err != nil && errors.IsNotFound(err) {
		log.Log.Info("Creating Deployment", "Deploy", deployer)
		// Error occurred while creating the ConfigMap
		//r.Log.Info("Creating ConfigMap", "ConfigMap", configMap)
		if err := r.Create(ctx, deployer); err != nil {
			// Error occurred while creating the ConfigMap
			return ctrl.Result{}, err
		}
		//instance as the owner and controller
		ctrl.SetControllerReference(instance, deployer, r.Scheme)
		return ctrl.Result{Requeue: true}, nil
	} else {
		log.Log.Info(fmt.Sprintf("Updating deployment %v", deployer))
		if err := r.Update(ctx, deployer); err != nil {
			// Error occurred while updating the ConfigMap
			return ctrl.Result{}, err
		}
		//instance as the owner and controller
		ctrl.SetControllerReference(instance, deployer, r.Scheme)
		return ctrl.Result{RequeueAfter: time.Duration(60 * time.Second)}, err
	}
}

/*
func (r *GreetReconciler) CreateResource(resource client.Object, ctx context.Context, resourceType string) (ctrl.Result, error) {

	name := resource.GetName()
	nameSpace := resource.GetNamespace()

	log.Log.Info(fmt.Sprintf("creating resource name: %s, namespace: %s", name, nameSpace))
	if err := r.Create(ctx, resource); err != nil {
		log.Log.Info("Error created", err)
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *GreetReconciler) UpdateResource(resource client.Object, ctx context.Context, resourceType string) (ctrl.Result, error) {
	name := resource.GetName()
	nameSpace := resource.GetNamespace()

	log.Log.Info(fmt.Sprintf("creating resource name: %s, namespace: %s", name, nameSpace))
	if err := r.Update(ctx, resource); err != nil {
		log.Log.Info("Error created", err)
	}

	return ctrl.Result{RequeueAfter: time.Duration(60 * time.Second)}, nil
}
*/
