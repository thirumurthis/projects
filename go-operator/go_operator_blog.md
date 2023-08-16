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

- From the WSL2, issue `make generate manifests install run` will deploy the controller to KinD cluster.
- Operator log once the sample manifests is deployed
![image](https://github.com/thirumurthis/projects/assets/6425536/d296ea1a-e8e8-48ff-9684-5bfc721f2e2f)

- Below is the sample Custom Resource manifests, which is deployed once the controller and CRD's are deployed.

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

- Once application is up, with the nginx pod we can hit the endpoint and below is the output
![image](https://github.com/thirumurthis/projects/assets/6425536/c0658ac8-04b2-46b2-be79-d334b35cb03a)


- Below command is used to build and deploy the image to docker hub
```
 make docker-build docker-push IMG=thirumurthi/app-op:v1
```

