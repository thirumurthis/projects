## Operator pattern to deploy SpringBoot application
 
In this article we will be using Operator SDK to create Custom Resource Definition (CRD) and Controller logic to deploy a Spring application in Kubernetes cluster.

With reference to my previous [operator-sdk blog](https://thirumurthi.hashnode.dev/extend-kubernetes-api-with-operator-sdk), we will use the same scaffolded and initialized project structure to create the CRD and reconcile logic. The reconciler logic will programmatically create Kubernetes resources object (using k8s.io/core libraries) like Deployment, ConfigMap, Service and deploy them to the cluster.

Pre-requisites:
 - Basic understanding of Kubernetes
 - KinD CLI installed and Cluster running
 - WSL2 installed 
 - GoLang installed in WSL2
 - Operator-SDK cli installed in WSL2

To demonstrate the use of Operator SDK in design and create CRD, update reconcile logic and deployment workflow - created a SpringBoot application that exposes an end point in 8080 port. This end point will render the response by reading the value from configuration file or environment variable.

What are the resources used for deploying Spring application?

- ConfigMap
        - The application.yaml content is deployed as ConfigMap resource. The deployment manifest should mount the ConfigMap as volume, so Spring application can access it. The mounted configMap is passed as external configuration, using `--spring.config.location` option when starting the Spring application. For more details refer the Kubernetes documentation.
- Deployment 
    - The deployment will deploy the Spring application in the container.
- Service
     - The service created will forward traffic to 8080 port.

### Operator pattern key components

- Custom Resource Definition (CRD), Custom Resource (CR) and Controller/Operator are the key components
- CRD & CR 
   - Custom Resource Definition (CRD) and Custom Resource (CR) plays a crucial part and should be designed based on requirements.
   - The Kubernetes API can be extended with Custom Resource Definition (CRD), this defines the schema for the Custom Resource (CR).
   - Once the CRD is defined we can create an instance of the Custom Resource.

- Controller
   - The controller in Operator SDK is responsible for watching for events and reconciling the Custom Resource (CR) to desire state based on logic defined in  reconciliation flow. Without controller just deploying the CRD to cluster doesn't do anything.

### Using Operator SDK

With Operator SDK we can design and define the Customer Resource Definition (CRD) with corresponding reconciler logic. Part of the reconcile logic we can deploy different Kubernetes resources.

#### Design 

The `spec` section in the Custom Resource (CR) is the design starting point, the properties that are required to deploy the application goes under this section.

In the below Custom Resource (CR) snippet, the `spec` section includes `name` and `deployment` properties. The `deployment` section further includes `name`, `pod`, `config` and `service` section.
- The `pod` section includes the properties related to Pod manifest,  like image, container port, etc. This will be read in the reconciler logic in Operator SDK project and a deployment object will be created to be deployed in cluster.
- The `config` section will include the data to create the ConfigMap resource.
- The `service` section include `name` and `spec`. The `spec` section in the `api/v1alpha1/*type.go` uses the k8s.io/core library object, so we can use the regular Service manifest properties like `port`, `targetPort`, etc.

- Below is the Custom Resource (CR) used to deploy the Spring application.
- By reading further we will see how this file is being deployed.

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

- With reference to the CR above, we can compare it with the Operator SDK generated `api/v1alpha1/*type.go` file content where we can easily follow the pattern.

Note:-
  - The `api/v1alpha1/*type.go` file includes different Go struct data types defined which will in a way used in the Custom Resource YAML.
  - `pod` and `config` a custom Go struct datatype is defined.
  - In `service`, the `spec` uses the k8s.io/core api *corev1.ServiceSpec* with which we can specify regular Service manifest properties.

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
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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

### Image creation of Spring application

   - First, build the jar artifact for the SpringBoot application, using `maven clean install`.
   - With the Dockerfile defined we can use Docker CLI to build image using `docker build` and `docker push` to push the image to Dockerhub.

### Approaches to deploy the Spring application image to Kubernetes

#### Using individual resource manifests files
1. The conventional way of deploying the app is to create different resource manifest YAML either in single file or different file. In this case we have to create three manifests for ConfigMap, Service and Deployment resources.

#### Using operator framework
2. With the Operator SDK project we need to define the CRD in here we have used GoLang and build the reconciler logic.
- Using the utilities provided in Operator SDK, we generate the CRD files and Controller image. 
- Once we developed the controller logic, in order to deploy the Spring application we need to deploy the controller image as deployment, then the CRD manifest and the Custom Resource (CR) to the cluster. This CR includes the image of the Spring application.

Note:- 
  - During development the `api/v1alpha1/*types.go` file will be updated with custom Go struct datatypes, whenever we update this file we need to execute `make generate manifests` command in WSL2 terminal to generate the CRDs YAML file.

### Code 
  
#### SpringBoot application entry point

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

#### Spring Configuration class
- Configuration class to read the application.yaml when application context is loaded

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

#### Dependencies

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

Note:- 
- Leave the `application.properties` file empty when building the jar.

#### Dockerfile 

```Dockerfile
FROM eclipse-temurin:17-jdk-alpine
ARG JAR_FILE
COPY ${JAR_FILE} app.jar
ENTRYPOINT ["java","-jar","/app.jar"]
```

### Image creation

- Create the jar file using`mvn clean install`. The jar file will be created in `target\` folder.

- Once the Jar file is generated, use below command to build the docker image.

```
docker build --build-arg JAR_FILE=target/*.jar -t local/app .
```

- We can run the Spring application image in Docker Desktop wiht below command.

```
docker run --name app -e ENV_NAME="docker-env" -p 8080:8080 -d local/app:latest
```

Note:-
     - The `env.name` value is passed as docker environment variable `ENV_NAME`.
     - After the container is ready, the endpoint `http://localhost:8080/api/hello?name=test` should return response

#### Output when running the application in Docker Desktop

```
$ curl -i http://localhost:8080/api/hello?name=test
HTTP/1.1 200                                                    
Content-Type: application/json                                  
Transfer-Encoding: chunked                                      
Date: Tue, 15 Aug 2023 03:36:14 GMT                             
                                                                
{"content":"Hello test!","source":"FROM-SPRING-CODE-docker-env"}
```

- Passing GREETING_SOURCE environment variable in Docker run command the output looks like below.

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

#### Push image to Dockerhub

- The images can pushed to the Dockerhub with below command. Use appropriate repository name.

```
docker build --build-arg JAR_FILE=target/*.jar -t <repoistory-name>/app:v1 .
```

### Deployment manifest without CRD or Operator SDK

- Once the image is pushed to Dockerhub or private registry, we can deploy the Spring application and the Deployment manifest looks like below. Assuming the ConfigMap is created as `app-cfg` 

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
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
        volumeMounts:
        - mountPath: /opt/app
          name: app-mount-vol
      volumes:
      - configMap:
          name: app-cfg
        name: app-mount-vol
```

### CRDs and controller logic in Operator SDK

- The  fragment of code from `api/v1alpha1/*types.go` file where we define custom Go struct data types, which will generate the CRDs accordingly.
- The Service struct uses the k8s.io/api/core  `corev1.ServiceSpec` so we can use `port`, `targetPort` properties like in regular Service manifest file. This is an example to demonstrate that it is always not necessary to define all the properties we can reuse the k8s.io core library object as well.

```go

type Deployment struct {
	Name      string  `json:"name"`
	Pod       PodInfo `json:"pod"`
	Replicas  int32   `json:"replicas,omitempty"`
	ConfigMap Config  `json:"config,omitempty"`
	Service   Service `json:"service,omitempty"`
}

type Service struct {
	Name string             `json:"name,omitempty"`
	Spec corev1.ServiceSpec `json:"spec,omitempty"`
}
```

#### Reconciler logic

- The `Reconcile()` method calls three different function which creates the Service, ConfigMap and Deployment resource.
- When the CR is applied, the create event will be triggered and the resources will be deployed. If the CR is deployed with any updates, the update event will update the resources. 

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

	// Used to deserialize
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

- The Custom Resource file content is show above.

- The CRD file can be found from [Github](https://github.com/thirumurthis/projects/blob/main/go-operator/app-op/config/crd/bases/greet.greetapp.com_greets.yaml)

#### Deployment tips

- During local development to run the operator in KinD we can use `make generate manifests install run` command. 
- Upon deploying the Custom Resource the logs will be displayed like in below snapshot. 
  
![image](https://github.com/thirumurthis/projects/assets/6425536/d296ea1a-e8e8-48ff-9684-5bfc721f2e2f)

- To deploy the Custom Resource manifests use save the content to a file and use the `kubectl apply -f <file-name-with-content>.yaml` command.

- In below the deployment has pod, config, service sections. The reconcile code will use these values to deploy them in Kuberented cluster.

- Note the config in the Custom Resource (CR) file

```yaml
    config: 
      name: app
      fileName: application.yaml
      data: |
        env.name: k8s-kind-dev-env
        greeting.source: from-k8s-configMap
````

- Once application is deployed, with nginx container we can hit the endpoint and validate the response, which will look like in below snapshot

![image](https://github.com/thirumurthis/projects/assets/6425536/c0658ac8-04b2-46b2-be79-d334b35cb03a)

### Creating controller image

- From WSL2 we can use below command to create the controller image and push it to Dockerhub.

```
 make docker-build docker-push IMG=thirumurthi/app-op:v1
```

### To deploy the Controller, CRD and CR from manifest to test or production like environment, we need to 
 - Deploy the Controller as deployment 
 - Deploy the CRD 
 - Deploy the CR

- Below is the deployment manifest to deploy the controller.

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
- To deploy the CRD, use the below command 

```
kubectl apply -f https://github.com/thirumurthis/projects/blob/main/go-operator/app-op/config/crd/bases/greet.greetapp.com_greets.yaml
```

- To deploy the CR, use the below command

```
kubectl apply -f https://github.com/thirumurthis/projects/blob/main/go-operator/app-op/config/samples/greet_v1alpha1_greet.yaml
```

- When I tried to deploy the controller as deployment, there was an due to permission and it requires a ServiceAccount to be created. The ServiceAccount can be created in Operator SDK but for simplicity executed the following command manually.

- The exception message that you would notice in case if ServiceAccount is not created.

```
E0816 03:51:14.671318       1 reflector.go:148] pkg/mod/k8s.io/client-go@v0.27.4/tools/cache/reflector.go:231: Failed to watch *v1alpha1.Greet: failed to list *v1alpha1.Greet: greets.greet.greetapp.com is forbidden: User "system:serviceaccount:default:default" cannot list resource "greets" in API group "greet.greetapp.com" at the cluster scope
```

- To manually create the ServiceAccount we can use below command

```
kubectl create clusterrole deployr --verb=get,list,watch,create,delete,patch,update --resource=deployments.apps
kubectl create clusterrolebinding deployr-srvacct-default-binding --clusterrole=deployr --serviceaccount=default:default
```

### Output after deploying the Controller, CRD and CR

- Once Controller, CRD and CR are deployed the list of resources and the output accessing the deployed spring application looks like below snapshot.

![image](https://github.com/thirumurthis/projects/assets/6425536/fbe560a4-0cf7-4691-9bff-eecae614298b)


