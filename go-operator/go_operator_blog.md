In this blog have created an example of how operator-sdk can be used to deploy a SpringBoot application to Kubernetes cluster.

Following this [operator-sdk blog](https://thirumurthi.hashnode.dev/extend-kubernetes-api-with-operator-sdk), this article briefly shows the Go code to programatically create the manifests and deploy to K8S cluster.

Pre-requesites:
 - KinD CLI installed and Cluster running
 - WSL2 installed 
 - GoLang installed in WSL2
 - Operator-SDK cli installed in WSL2

In order to demonstrate the operator-sdk usage, we have a simple SpringBoot application which exposes an simple endpoint at 8080.
The endpoint will read the value from the configuration file `application.yaml` or if the values passed via environment variables.

Create image of the SpringBoot application
   - Build the jar for the SpringBoot application, using `maven clean install`. 
   - Create a Dockerfile to convert the jar artifact to image and push it to Dockerhub.

1. Conventional way of deploying the image to Kubernetes (without operator-sdk)
  - Create different manifests
        - ConfigMap
        - Service
        - Deployment (mount the ConfigMap as a volume)

2. Operator-Sdk approach
  - Once the project is scaffold and initalized, we need to update the `*types.go` file generated. This file will be used to generate the CRD yaml. Which can be generated when using the `make generate manifests`, from the WSL2 terminal.
  - The Reconciler logic needs to be updated, in this case we use the `k8s.io` library to create ConfigMap, Service and Deployment GoLang objects. In the reconciler logic, we simply create the resource if not found, else we update it.


- Below is the SpringBoot application entry point 

```java
package com.app.app;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@SpringBootApplication
@RequestMapping("/api")
@RestController
public class AppApplication {

    @Bean
    GreetConfiguration getGreetConfig(){
        return new GreetConfiguration();
    }

	public static void main(String[] args) {
		SpringApplication.run(AppApplication.class, args);
	}

    @Value("${env.name:no-env-provided}")
    private String envName;
    public record Greeting(String content,String source) { }

    @GetMapping("/hello")
    public Greeting hello(@RequestParam(value = "name", defaultValue = "anonymous") String name) {

        return new Greeting(String.format("Hello %s!", name),
                            String.format("%s-%s",getGreetConfig().getSource().toUpperCase(),envName));
    }
}
```

- Simple configuration class, to read the application.yaml when application context is loaded by Spring

```java
package com.app.app;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Configuration
@ConfigurationProperties(prefix="greeting")
@Data
public class GreetConfiguration {
    private String source = "from-spring-code";
}
```

- pom.xml
```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
	<modelVersion>4.0.0</modelVersion>
	<parent>
		<groupId>org.springframework.boot</groupId>
		<artifactId>spring-boot-starter-parent</artifactId>
		<version>3.1.2</version>
		<relativePath/> <!-- lookup parent from repository -->
	</parent>
	<groupId>com.app</groupId>
	<artifactId>app</artifactId>
	<version>0.0.1-SNAPSHOT</version>
	<name>app</name>
	<description>Simple app with </description>
	<properties>
		<java.version>17</java.version>
		<spring-cloud.version>2022.0.3</spring-cloud.version>
	</properties>

	<dependencies>
	    <dependency>
		<groupId>org.springframework.boot</groupId>
		<artifactId>spring-boot-starter-web</artifactId>
	   </dependency>
           <dependency>
		<groupId>org.projectlombok</groupId>
		<artifactId>lombok</artifactId>
		<optional>true</optional>
	    </dependency>
            <dependency>
		<groupId>org.springframework.boot</groupId>
		<artifactId>spring-boot-starter-web</artifactId>
	    </dependency>
	    <dependency>
		<groupId>org.springframework.boot</groupId>
		<artifactId>spring-boot-starter-test</artifactId>
		<scope>test</scope>
	    </dependency>
	</dependencies>
	<build>
		<plugins>
			<plugin>
				<groupId>org.springframework.boot</groupId>
				<artifactId>spring-boot-maven-plugin</artifactId>
				<configuration>
					<excludes>
						<exclude>
							<groupId>org.projectlombok</groupId>
							<artifactId>lombok</artifactId>
						</exclude>
					</excludes>
				</configuration>
			</plugin>
			<plugin>
				<artifactId>maven-surefire-plugin</artifactId>
				<version>3.0.0-M4</version>
			</plugin>
		</plugins>
	</build>
</project>
```

- Leave the application.properties file empty when building the jar.

- Dockerfile 
```Dockerfile
FROM eclipse-temurin:17-jdk-alpine
ARG JAR_FILE
COPY ${JAR_FILE} app.jar
ENTRYPOINT ["java","-jar","/app.jar"]
```

- To create the jar file,  issuing `mvn clean install`. The jar file will be created in `target\` folder.

- Once the Jar file is generated, to build the docker image we can issue below command.

```
docker build --build-arg JAR_FILE=target/*.jar -t local/app .
```

- Once the image is build, to run the SpringBoot application in Docker Desktop, use the below  command.
     - The `env.name` value is passed as docker environment variable `ENV_NAME`.
     - After the container is ready, the endpoint `http://localhost:8080/api/hello?name=test` should return response

```
docker run --name app -e ENV_NAME="docker-env" -p 8080:8080 -d local/app:latest
```

- The output of the application when hitting the endpoint

```
$ curl -i http://localhost:8080/api/hello?name=test
HTTP/1.1 200                                                    
Content-Type: application/json                                  
Transfer-Encoding: chunked                                      
Date: Tue, 15 Aug 2023 03:36:14 GMT                             
                                                                
{"content":"Hello test!","source":"FROM-SPRING-CODE-docker-env"}

```

- With below docker command, additional environment variable GREETING_SOURCE is passed with value, the output section is how the response looks like.

```
docker run --name app -e ENV_NAME="docker-env" -e GREETING_SOURCE="from-docker-cli" -p 8080:8080 -d local/app:latest

```

- Output
 
```
$ curl -i http://localhost:8080/api/hello?name=test                                
HTTP/1.1 200 
Content-Type: application/json
Transfer-Encoding: chunked
Date: Tue, 15 Aug 2023 03:40:40 GMT

{"content":"Hello test!","source":"FROM-DOCKER-CLI-docker-env"}
```

- To push the the images to the Dockerhub, use below command, the repository name would be the name 

```
docker build --build-arg JAR_FILE=target/*.jar -t <repoistory-name>/app:v1 .
```

## Once the image is available, below is just the deployment manifest that can be deployed to Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    name: first-app
  name: app-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      name: first-app
  template:
    metadata:
      creationTimestamp: null
      labels:
        name: first-app
    spec:
      containers:
      - image: <repository-name>/app:v1
        imagePullPolicy: Always
        name: app-pod
        ports:
        - containerPort: 80
          name: app-port
        resources: {}
        volumeMounts:
        - mountPath: /opt/app
          name: app-mount-vol
      volumes:
      - configMap:
          name: app-cfg
        name: app-mount-vol
```

## Operator-sdk framework with the custom resource definition changes

- Below is the  `*types.go` file, most of the datatype is go struct. The service struct includes, a `corev1.ServiceSpec`.

```go
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GreetSpec defines the desired state of Greet
type GreetSpec struct {

	//Name of the resource
	// +kubebuilder:validation:MaxLength=25
	// +kubebuilder:validation:MinLength=1
	Name       string     `json:"name"`
	Deployment Deployment `json:"deployment"`
}

type Deployment struct {
	Name      string  `json:"name"`
	Pod       PodInfo `json:"pod"`
	Replicas  int32   `json:"replicas,omitempty"`
	ConfigMap Config  `json:"config,omitempty"`
	Service   Service `json:"service,omitempty"`
}

type Config struct {
	Name     string `json:"name,omitempty"`
	FileName string `json:"fileName,omitempty"`
	Data     string `json:"data,omitempty"`
}

type PodInfo struct {
	Image           string   `json:"image"`
	ImagePullPolicy string   `json:"imagePullPolicy,omitempty"`
	MountName       string   `json:"mountName,omitempty"`
	MountPath       string   `json:"mountPath,omitempty"`
	PodPort         int32    `json:"podPort,omitempty"`
	Command         []string `json:"command,omitempty"`
	Args            []string `json:"args,omitempty"`
}

type Service struct {
	Name string             `json:"name,omitempty"`
	Spec corev1.ServiceSpec `json:"spec,omitempty"`
}

// GreetStatus defines the observed state of Greet
type GreetStatus struct {
	Status string `json:"status,omitempty"`
}

//Don't leave any space between the marker

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="APPNAME",type="string",JSONPath=".spec.name",description="Name of the app"
//+kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.status",description="Status of the app"
//+kubebuilder:subresource:status
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Greet App"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployment,v1,\"A Kubernetes Deployment of greet app\""

type Greet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   GreetSpec   `json:"spec,omitempty"`
	Status GreetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GreetList contains a list of Greet
type GreetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Greet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Greet{}, &GreetList{})
}
```

- Below is the controller code, where `Reconcile()` calls the function to create the Service, ConfigMap and Deployment.

```go
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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

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

func (r *GreetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

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
```

- The Custom Resource looks like below, where the service, configmap and deployment is defined in CR.

```yaml
apiVersion: greet.greetapp.com/v1alpha1
kind: Greet
metadata:
  labels:
    app.kubernetes.io/name: greet
    app.kubernetes.io/instance: greet-sample
    app.kubernetes.io/part-of: app-op
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: app-op
  name: greet-sample
spec:
  # TODO(user): Add fields here
  name: first-app
  deployment:
    name: app
    replicas: 1
    pod:
      image: thirumurthi/app:v1
      imagePullPolicy: Always
      mountName: app-mount
      mountPath: /opt/app
      podPort: 8080
        #command: ["java","-jar","app.jar","--spring.config.location=file://opt/app/application.yaml"]
      command: ["java"]
      args: ["-jar","app.jar","--spring.config.location=file:/opt/app/application.yaml"]
    config: 
      name: app
      fileName: application.yaml
      data: |
        env.name: Kubernetes-k8s-000
    service:
      name: app
      spec:
        selector:
          name: first-app
        ports:
        - name: svc-port
          protocol: TCP
          port: 80
          targetPort: 8080
```

- During local development, from the operator-sdk project, form WSL2 issue  `make generate manifests install run` to deploy the controller to KinD cluster. The command will display the logs and once we deploy the sample Custom Resource (like in below code snippet). The log should display the message logged in the Reconcilde code. Below snapshot where the deployment info is printed in console.
  
![image](https://github.com/thirumurthis/projects/assets/6425536/d296ea1a-e8e8-48ff-9684-5bfc721f2e2f)

- Sample Custom Resource manifests, save this content in a file and issue `kubectl apply -f <file-name-with-content>.yaml` to deploy the resources.
- In below the deployment has pod, config, service sections. The reconcile code will use these values to deploy them in Kuberented cluster.

```yaml
apiVersion: greet.greetapp.com/v1alpha1
kind: Greet
metadata:
  labels:
    app.kubernetes.io/name: greet
    app.kubernetes.io/instance: greet-sample
    app.kubernetes.io/part-of: app-op
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: app-op
  name: greet-sample
spec:
  # TODO(user): Add fields here
  name: first-app
  deployment:
    name: app
    replicas: 1
    pod:
      image: thirumurthi/app:v1
      imagePullPolicy: Always
      mountName: app-mount
      mountPath: /opt/app
      podPort: 8080
      command: ["java"]
      args: ["-jar","app.jar","--spring.config.location=file:/opt/app/application.yaml"]
    config: 
      name: app
      fileName: application.yaml
      data: |
        env.name: k8s-kind-dev-env
        greeting.source: from-k8s-configMap
    service:
      name: app
      spec:
        selector:
          name: first-app
        ports:
        - name: svc-port
          protocol: TCP
          port: 80
          targetPort: 8080

```

- Once application is up, with the nginx pod we can hit the endpoint and the repsonce will look like in below snapshot
![image](https://github.com/thirumurthis/projects/assets/6425536/c0658ac8-04b2-46b2-be79-d334b35cb03a)

- The CRD generated in the operator project looks like below

```yaml
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.11.1
  creationTimestamp: null
  name: greets.greet.greetapp.com
spec:
  group: greet.greetapp.com
  names:
    kind: Greet
    listKind: GreetList
    plural: greets
    singular: greet
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - description: Name of the app
      jsonPath: .spec.name
      name: APPNAME
      type: string
    - description: Status of the app
      jsonPath: .status.status
      name: STATUS
      type: string
    name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: GreetSpec defines the desired state of Greet
            properties:
              deployment:
                properties:
                  config:
                    properties:
                      data:
                        type: string
                      fileName:
                        type: string
                      name:
                        type: string
                    type: object
                  name:
                    type: string
                  pod:
                    properties:
                      args:
                        items:
                          type: string
                        type: array
                      command:
                        items:
                          type: string
                        type: array
                      image:
                        type: string
                      imagePullPolicy:
                        type: string
                      mountName:
                        type: string
                      mountPath:
                        type: string
                      podPort:
                        format: int32
                        type: integer
                    required:
                    - image
                    type: object
                  replicas:
                    format: int32
                    type: integer
                  service:
                    properties:
                      name:
                        type: string
                      spec:
                        description: ServiceSpec describes the attributes that a user
                          creates on a service.
                        properties:
                          allocateLoadBalancerNodePorts:
                            description: allocateLoadBalancerNodePorts defines if
                              NodePorts will be automatically allocated for services
                              with type LoadBalancer.  Default is "true". It may be
                              set to "false" if the cluster load-balancer does not
                              rely on NodePorts.  If the caller requests specific
                              NodePorts (by specifying a value), those requests will
                              be respected, regardless of this field. This field may
                              only be set for services with type LoadBalancer and
                              will be cleared if the type is changed to any other
                              type.
                            type: boolean
                          clusterIP:
                            description: 'clusterIP is the IP address of the service
                              and is usually assigned randomly. If an address is specified
                              manually, is in-range (as per system configuration),
                              and is not in use, it will be allocated to the service;
                              otherwise creation of the service will fail. This field
                              may not be changed through updates unless the type field
                              is also being changed to ExternalName (which requires
                              this field to be blank) or the type field is being changed
                              from ExternalName (in which case this field may optionally
                              be specified, as describe above).  Valid values are
                              "None", empty string (""), or a valid IP address. Setting
                              this to "None" makes a "headless service" (no virtual
                              IP), which is useful when direct endpoint connections
                              are preferred and proxying is not required.  Only applies
                              to types ClusterIP, NodePort, and LoadBalancer. If this
                              field is specified when creating a Service of type ExternalName,
                              creation will fail. This field will be wiped when updating
                              a Service to type ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies'
                            type: string
                          clusterIPs:
                            description: "ClusterIPs is a list of IP addresses assigned
                              to this service, and are usually assigned randomly.
                              \ If an address is specified manually, is in-range (as
                              per system configuration), and is not in use, it will
                              be allocated to the service; otherwise creation of the
                              service will fail. This field may not be changed through
                              updates unless the type field is also being changed
                              to ExternalName (which requires this field to be empty)
                              or the type field is being changed from ExternalName
                              (in which case this field may optionally be specified,
                              as describe above).  Valid values are \"None\", empty
                              string (\"\"), or a valid IP address.  Setting this
                              to \"None\" makes a \"headless service\" (no virtual
                              IP), which is useful when direct endpoint connections
                              are preferred and proxying is not required.  Only applies
                              to types ClusterIP, NodePort, and LoadBalancer. If this
                              field is specified when creating a Service of type ExternalName,
                              creation will fail. This field will be wiped when updating
                              a Service to type ExternalName.  If this field is not
                              specified, it will be initialized from the clusterIP
                              field.  If this field is specified, clients must ensure
                              that clusterIPs[0] and clusterIP have the same value.
                              \n This field may hold a maximum of two entries (dual-stack
                              IPs, in either order). These IPs must correspond to
                              the values of the ipFamilies field. Both clusterIPs
                              and ipFamilies are governed by the ipFamilyPolicy field.
                              More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies"
                            items:
                              type: string
                            type: array
                            x-kubernetes-list-type: atomic
                          externalIPs:
                            description: externalIPs is a list of IP addresses for
                              which nodes in the cluster will also accept traffic
                              for this service.  These IPs are not managed by Kubernetes.  The
                              user is responsible for ensuring that traffic arrives
                              at a node with this IP.  A common example is external
                              load-balancers that are not part of the Kubernetes system.
                            items:
                              type: string
                            type: array
                          externalName:
                            description: externalName is the external reference that
                              discovery mechanisms will return as an alias for this
                              service (e.g. a DNS CNAME record). No proxying will
                              be involved.  Must be a lowercase RFC-1123 hostname
                              (https://tools.ietf.org/html/rfc1123) and requires `type`
                              to be "ExternalName".
                            type: string
                          externalTrafficPolicy:
                            description: externalTrafficPolicy describes how nodes
                              distribute service traffic they receive on one of the
                              Service's "externally-facing" addresses (NodePorts,
                              ExternalIPs, and LoadBalancer IPs). If set to "Local",
                              the proxy will configure the service in a way that assumes
                              that external load balancers will take care of balancing
                              the service traffic between nodes, and so each node
                              will deliver traffic only to the node-local endpoints
                              of the service, without masquerading the client source
                              IP. (Traffic mistakenly sent to a node with no endpoints
                              will be dropped.) The default value, "Cluster", uses
                              the standard behavior of routing to all endpoints evenly
                              (possibly modified by topology and other features).
                              Note that traffic sent to an External IP or LoadBalancer
                              IP from within the cluster will always get "Cluster"
                              semantics, but clients sending to a NodePort from within
                              the cluster may need to take traffic policy into account
                              when picking a node.
                            type: string
                          healthCheckNodePort:
                            description: healthCheckNodePort specifies the healthcheck
                              nodePort for the service. This only applies when type
                              is set to LoadBalancer and externalTrafficPolicy is
                              set to Local. If a value is specified, is in-range,
                              and is not in use, it will be used.  If not specified,
                              a value will be automatically allocated.  External systems
                              (e.g. load-balancers) can use this port to determine
                              if a given node holds endpoints for this service or
                              not.  If this field is specified when creating a Service
                              which does not need it, creation will fail. This field
                              will be wiped when updating a Service to no longer need
                              it (e.g. changing type). This field cannot be updated
                              once set.
                            format: int32
                            type: integer
                          internalTrafficPolicy:
                            description: InternalTrafficPolicy describes how nodes
                              distribute service traffic they receive on the ClusterIP.
                              If set to "Local", the proxy will assume that pods only
                              want to talk to endpoints of the service on the same
                              node as the pod, dropping the traffic if there are no
                              local endpoints. The default value, "Cluster", uses
                              the standard behavior of routing to all endpoints evenly
                              (possibly modified by topology and other features).
                            type: string
                          ipFamilies:
                            description: "IPFamilies is a list of IP families (e.g.
                              IPv4, IPv6) assigned to this service. This field is
                              usually assigned automatically based on cluster configuration
                              and the ipFamilyPolicy field. If this field is specified
                              manually, the requested family is available in the cluster,
                              and ipFamilyPolicy allows it, it will be used; otherwise
                              creation of the service will fail. This field is conditionally
                              mutable: it allows for adding or removing a secondary
                              IP family, but it does not allow changing the primary
                              IP family of the Service. Valid values are \"IPv4\"
                              and \"IPv6\".  This field only applies to Services of
                              types ClusterIP, NodePort, and LoadBalancer, and does
                              apply to \"headless\" services. This field will be wiped
                              when updating a Service to type ExternalName. \n This
                              field may hold a maximum of two entries (dual-stack
                              families, in either order).  These families must correspond
                              to the values of the clusterIPs field, if specified.
                              Both clusterIPs and ipFamilies are governed by the ipFamilyPolicy
                              field."
                            items:
                              description: IPFamily represents the IP Family (IPv4
                                or IPv6). This type is used to express the family
                                of an IP expressed by a type (e.g. service.spec.ipFamilies).
                              type: string
                            type: array
                            x-kubernetes-list-type: atomic
                          ipFamilyPolicy:
                            description: IPFamilyPolicy represents the dual-stack-ness
                              requested or required by this Service. If there is no
                              value provided, then this field will be set to SingleStack.
                              Services can be "SingleStack" (a single IP family),
                              "PreferDualStack" (two IP families on dual-stack configured
                              clusters or a single IP family on single-stack clusters),
                              or "RequireDualStack" (two IP families on dual-stack
                              configured clusters, otherwise fail). The ipFamilies
                              and clusterIPs fields depend on the value of this field.
                              This field will be wiped when updating a service to
                              type ExternalName.
                            type: string
                          loadBalancerClass:
                            description: loadBalancerClass is the class of the load
                              balancer implementation this Service belongs to. If
                              specified, the value of this field must be a label-style
                              identifier, with an optional prefix, e.g. "internal-vip"
                              or "example.com/internal-vip". Unprefixed names are
                              reserved for end-users. This field can only be set when
                              the Service type is 'LoadBalancer'. If not set, the
                              default load balancer implementation is used, today
                              this is typically done through the cloud provider integration,
                              but should apply for any default implementation. If
                              set, it is assumed that a load balancer implementation
                              is watching for Services with a matching class. Any
                              default load balancer implementation (e.g. cloud providers)
                              should ignore Services that set this field. This field
                              can only be set when creating or updating a Service
                              to type 'LoadBalancer'. Once set, it can not be changed.
                              This field will be wiped when a service is updated to
                              a non 'LoadBalancer' type.
                            type: string
                          loadBalancerIP:
                            description: 'Only applies to Service Type: LoadBalancer.
                              This feature depends on whether the underlying cloud-provider
                              supports specifying the loadBalancerIP when a load balancer
                              is created. This field will be ignored if the cloud-provider
                              does not support the feature. Deprecated: This field
                              was under-specified and its meaning varies across implementations,
                              and it cannot support dual-stack. As of Kubernetes v1.24,
                              users are encouraged to use implementation-specific
                              annotations when available. This field may be removed
                              in a future API version.'
                            type: string
                          loadBalancerSourceRanges:
                            description: 'If specified and supported by the platform,
                              this will restrict traffic through the cloud-provider
                              load-balancer will be restricted to the specified client
                              IPs. This field will be ignored if the cloud-provider
                              does not support the feature." More info: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/'
                            items:
                              type: string
                            type: array
                          ports:
                            description: 'The list of ports that are exposed by this
                              service. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies'
                            items:
                              description: ServicePort contains information on service's
                                port.
                              properties:
                                appProtocol:
                                  description: The application protocol for this port.
                                    This field follows standard Kubernetes label syntax.
                                    Un-prefixed names are reserved for IANA standard
                                    service names (as per RFC-6335 and https://www.iana.org/assignments/service-names).
                                    Non-standard protocols should use prefixed names
                                    such as mycompany.com/my-custom-protocol.
                                  type: string
                                name:
                                  description: The name of this port within the service.
                                    This must be a DNS_LABEL. All ports within a ServiceSpec
                                    must have unique names. When considering the endpoints
                                    for a Service, this must match the 'name' field
                                    in the EndpointPort. Optional if only one ServicePort
                                    is defined on this service.
                                  type: string
                                nodePort:
                                  description: 'Service to no longer need it (e.g. changing type
                                    from NodePort to ClusterIP). More info: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport'
                                  format: int32
                                  type: integer
                                port:
                                  description: The port that will be exposed by this
                                    service.
                                  format: int32
                                  type: integer
                                protocol:
                                  default: TCP
                                  description: The IP protocol for this port. Supports
                                    "TCP", "UDP", and "SCTP". Default is TCP.
                                  type: string
                                targetPort:
                                  anyOf:
                                  - type: integer
                                  - type: string
                                  description: 'Number or name of the port to access
                                    on the pods targeted by the service. Number must
                                    be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                    If this is a string, it will be looked up as a
                                    named port in the target Pod''s container ports.
                                    If this is not specified, the value of the ''port''
                                    field is used (an identity map). This field is
                                    ignored for services with clusterIP=None, and
                                    should be omitted or set equal to the ''port''
                                    field. More info: https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service'
                                  x-kubernetes-int-or-string: true
                              required:
                              - port
                              type: object
                            type: array
                            x-kubernetes-list-map-keys:
                            - port
                            - protocol
                            x-kubernetes-list-type: map
                          publishNotReadyAddresses:
                            description: publishNotReadyAddresses indicates that any
                              agent which deals with endpoints for this Service should
                              disregard any indications of ready/not-ready. The primary
                              use case for setting this field is for a StatefulSet's
                              Headless Service to propagate SRV DNS records for its
                              Pods for the purpose of peer discovery. The Kubernetes
                              controllers that generate Endpoints and EndpointSlice
                              resources for Services interpret this to mean that all
                              endpoints are considered "ready" even if the Pods themselves
                              are not. Agents which consume only Kubernetes generated
                              endpoints through the Endpoints or EndpointSlice resources
                              can safely assume this behavior.
                            type: boolean
                          selector:
                            additionalProperties:
                              type: string
                            description: 'Route service traffic to pods with label
                              keys and values matching this selector. If empty or
                              not present, the service is assumed to have an external
                              process managing its endpoints, which Kubernetes will
                              not modify. Only applies to types ClusterIP, NodePort,
                              and LoadBalancer. Ignored if type is ExternalName. More
                              info: https://kubernetes.io/docs/concepts/services-networking/service/'
                            type: object
                            x-kubernetes-map-type: atomic
                          sessionAffinity:
                            description: 'Supports "ClientIP" and "None". Used to
                              maintain session affinity. Enable client IP based session
                              affinity. Must be ClientIP or None. Defaults to None.
                              More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies'
                            type: string
                          sessionAffinityConfig:
                            description: sessionAffinityConfig contains the configurations
                              of session affinity.
                            properties:
                              clientIP:
                                description: clientIP contains the configurations
                                  of Client IP based session affinity.
                                properties:
                                  timeoutSeconds:
                                    description: timeoutSeconds specifies the seconds
                                      of ClientIP type session sticky time. The value
                                      must be >0 && <=86400(for 1 day) if ServiceAffinity
                                      == "ClientIP". Default value is 10800(for 3
                                      hours).
                                    format: int32
                                    type: integer
                                type: object
                            type: object
                          type:
                            description: 'do not apply to ExternalName services. More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types'
                            type: string
                        type: object
                    type: object
                required:
                - name
                - pod
                type: object
              name:
                description: Name of the resource
                maxLength: 25
                minLength: 1
                type: string
            required:
            - deployment
            - name
            type: object
          status:
            description: GreetStatus defines the observed state of Greet
            properties:
              status:
                description: 'INSERT ADDITIONAL STATUS FIELD - define observed state
                  of cluster Important: Run "make" to regenerate code after modifying
                  this file'
                type: string
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
```

### Creating an image for the operator or controller changes.
- With the below command from operator-sdk project the controller image can be created and pushed to Docker.

```
 make docker-build docker-push IMG=thirumurthi/app-op:v1
```

### To deploy the Controller, CRD and CR from manifest to production like environment.

- Below is the deployment manifest which uses the operator or controller image from dockerhub, we can deploy this to Kuberentes cluster.
- Following the operator deployment, we can deploy the CRD and the CR. 

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    name: operator-controller
  name: app-controller
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      name: app-controller
  template:
    metadata:
      labels:
        name: app-controller
    spec:
      containers:
      - image: thirumurthi/app-op:v1
        name: app-controller
```

- In case if the there are exception deploying the operator or controller, we need to create a ServiceAccount

The exception message,

```
E0816 03:51:14.671318       1 reflector.go:148] pkg/mod/k8s.io/client-go@v0.27.4/tools/cache/reflector.go:231: Failed to watch *v1alpha1.Greet: failed to list *v1alpha1.Greet: greets.greet.greetapp.com is forbidden: User "system:serviceaccount:default:default" cannot list resource "greets" in API group "greet.greetapp.com" at the cluster scope
```

- We need to create the service account
```
kubectl create clusterrole deployer --verb=get,list,watch,create,delete,patch,update --resource=deployments.apps
kubectl create clusterrolebinding deployer-srvacct-default-binding --clusterrole=deployer --serviceaccount=default:default
```


- The deployed resources looks like below,
![image](https://github.com/thirumurthis/projects/assets/6425536/fbe560a4-0cf7-4691-9bff-eecae614298b)


