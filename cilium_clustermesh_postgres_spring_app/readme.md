### Instruction to configure the cilium clustermesh and deploy postgres with spring boot app

Execute below command coloning the repo and navigating to this path cilium_postgres

Create the kind cluster 

```sh
kind create cluster --name cluster-dev0 --config cluster_0_kind_config.yaml
```

```sh
kind create cluster --name cluster-dev1 --config cluster_1_kind_config.yaml
```

Set the variable with the Kuberentes context

```sh
CLUSTER0=kind-cluster-dev0
CLUSTER1=kind-cluster-dev1 
```

Apply the cilium cluster mesh on CLUSTER 0

```bash
helm upgrade -i cilium oci://quay.io/cilium/charts/cilium --version 1.20.2 \
   --namespace kube-system \
    -f clusters_cilium_common_config.yaml \
    -f cluster_0_cilium_config.yaml \
    --kube-context $CLUSTER0
```

Copy the cert ca-cilium to CLUSTER 1

```bash
kubectl --context $CLUSTER0 get secret -n kube-system cilium-ca -o yaml | \
    kubectl --context $CLUSTER1 create -f -
```

Apply the cilium cluster mesh on CLUSTER 1

```bash
helm upgrade -i cilium oci://quay.io/cilium/charts/cilium --version 1.20.2 \
   --namespace kube-system \
    -f clusters_cilium_common_config.yaml \
    -f cluster_1_cilium_config.yaml \
    --kube-context $CLUSTER1
```

Apply the monitoring stack to the CLUSTER 0 and CLUSTER 1

The configuration is example from - `https://raw.githubusercontent.com/cilium/cilium/refs/heads/main/examples/kubernetes/addons/prometheus/monitoring-example.yaml`

Since the grafana configuration should include envrionment

```yaml
    env:
    - name: GF_SERVER_ROOT_URL
        value: "%(protocol)s://%(domain)s:%(http_port)s/grafana/"
    - name: GF_SERVER_SERVE_FROM_SUB_PATH
        value: "true"
```

Also the grafana config prometheus datasource

```yaml
    datasources:
    - name: prometheus
      # routing the endpoint path the /prometheus should be added
      url: http://prometheus:9090/prometheus
      #...
```

-- Apply configuration on both CLUSTER 0 and CLUSTER 1

```bash
kubectl --context $CLUSTER0 -n cilium-monitoring apply -f clusters_cilium_metrics_monitoring.yaml
```

```bash
kubectl --context $CLUSTER1 -n cilium-monitoring apply -f clusters_cilium_metrics_monitoring.yaml
```

Deploy the gateway configuration on CLUSTER 0 and CLUSTER 1

```bash
kubectl --context $CLUSTER0 apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.6.1/standard-install.yaml

kubectl --context $CLUSTER1 apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.6.1/standard-install.yaml
```

Apply the gatewayclass manifest CLUSTER 0, namespace included in the manifest

```bash 
kubectl --context $CLUSTER0 apply -f clusters_gatewayClass.yaml
```

-- patch the node port, this port should match the kind config extraportmappings poperty

```bash
kubectl --context $CLUSTER0 patch svc cilium-gateway-nodeport-gateway -n cilium-monitoring -p='{"spec":{"ports":[{"port":80,"nodePort":31081},{"port":5432,"nodePort":31235}]}}'
```

Apply the Http routes for grafana, prometheus, etc CLUSTER 0. The namespace included in the manfiest. 
```bash
kubectl --context $CLUSTER0 apply -f clusters_routes_grafana_prometheus_hubble.yaml
```

Apply the gatewayclass manifest CLUSTER 1. The namespace included in the manfiest

```bash
kubectl --context $CLUSTER1 apply -f clusters_gatewayClass.yaml 
```

Patch the node port created by the cilium, these port should match the kind config extraportmappings poperty

```bash
kubectl --context $CLUSTER1 patch svc cilium-gateway-nodeport-gateway -n cilium-monitoring -p='{"spec":{"ports":[{"port":80,"nodePort":31082},{"port":5432,"nodePort":31236}]}}'
```

Apply the Http routes for grafana, prometheus, etc in CLUSTER 1. The namespace included in the manifest

```bash
kubectl --context $CLUSTER1  apply -f clusters_routes_grafana_prometheus_hubble.yaml
```

Access the ui, from browser use http://localhost:8081/grafana, http://localhost:8082/prometheus

For hubble ui, set the host file with hubble.local, then in browser use http://hubble.local:8081
--- 

Deploying the postgres 

Create the namespace for zalando operator and postgres CLUSTER 0 and CLUSTER 1

```bash
kubectl --context $CLUSTER0 create ns zalando
kubectl --context $CLUSTER0 create ns postgres
```

```bash
kubectl --context $CLUSTER1 create ns zalando
kubectl --context $CLUSTER1 create ns postgres
```

Deploy the zalando operator using helm command on CLUSTER 0, the configuration added to persist the secret

```bash
helm upgrade --install  postgres-operator \
postgres-operator-charts/postgres-operator \
-n zalando \
--set configKubernetes.enable_secrets_deletion=false \
--create-namespace \
--kube-context $CLUSTER0
```

Deploy the zalando opertor using helm command on CLUSTER 1,  the configuration added to persist the secret 

```bash
helm upgrade --install postgres-operator \
postgres-operator-charts/postgres-operator \
-n zalando \
--set configKubernetes.enable_secrets_deletion=false \
--create-namespace  \
--kube-context $CLUSTER1
```

Create the secret specific name for zalandao opertor format for primary and standby on CLUSTER 0 and CLUSTER 1

```bash
kubectl --context $CLUSTER0 apply -f clusters_primary_standby_postgres_secret.yaml
kubectl --context $CLUSTER1 apply -f clusters_primary_standby_postgres_secret.yaml
```

Deploy the postgres cluster config manifest using kubectl on the active side CLUSTER 0

```bash
kubectl --context $CLUSTER0 -n postgres apply -f cluster_0_postgres_config.yaml
```

Wait for few minutes 

Apply the TCP route config for the postgres db on CLUSTER 0

```bash
kubectl --context $CLUSTER0 -n postgres apply -f clusters_tcpRoute_postgres.yaml
```

Deploy the postgres cluster config manifest using kubectl on the standby side CLUSTER 1 

```bash
kubectl --context $CLUSTER1 -n postgres apply -f cluster_1_postgres_config.yaml
```

wait for few minutes 

Apply the TCP route config for the postgres db on CLUSTER 1

```bash
kubectl --context $CLUSTER1 -n postgres apply -f clusters_tcpRoute_postgres.yaml
```

To connect using the VSCode, with the Microsoft extension, connection string to be used from the host machine. The url looks like below, passing the password in the password text box and try to test connect. Template if connection string is  postgres://[user]:[password]@[host]:[port]/[database_name]

```bash
postgres://postgres:@127.0.0.1:8085/postgres
```

At this point we can check the postgres status 

```sh
kubectl --context $CLUSTER1 -n postgres exec pod/postgres-cluster-0 -- patronictl list
```

---
Deploy the spring boot app

Create the app namespace on both the cluster

```bash
kubectl --context $CLUSTER0 create ns app
kubectl --context $CLUSTER1 create ns app
```
 
-- The connection database url is base64 encoded, the postgres info is copied from postgres secret created earlier 

-- create secret for spring app in CLUSTER 0

```bash
kubectl --context $CLUSTER0 get -n postgres secret postgres.postgres-cluster.credentials.postgresql.acid.zalan.do -o go-template='{{ printf "apiVersion: v1\nkind: Secret\nmetadata:\n  name: postgres-credentials\n  namespace: app\ntype: Opaque\ndata:\n  db-user: %s\n  db-password: %s\n  db-url:  'amRiYzpwb3N0Z3Jlc3FsOi8vMTI3LjAuMC4xOjgwODEvcG9zdGdyZXM/dGFyZ2V0U2VydmVyVHlwZT1wcmltYXJ5Jmhvc3RSZWNoZWNrU2Vjb25kcz0xMA=='\n" (index .data "username") (index .data "password") }}' | kubectl --context $CLUSTER0 apply -f -
```

-- create secret for spring app in CLUSTER 1

```bash
kubectl --context $CLUSTER1 get -n postgres secret postgres.postgres-cluster.credentials.postgresql.acid.zalan.do -o go-template='{{ printf "apiVersion: v1\nkind: Secret\nmetadata:\n  name: postgres-credentials\n  namespace: app\ntype: Opaque\ndata:\n  db-user: %s\n  db-password: %s\n  db-url: 'amRiYzpwb3N0Z3Jlc3FsOi8vMTI3LjAuMC4xOjgwODEvcG9zdGdyZXM/dGFyZ2V0U2VydmVyVHlwZT1wcmltYXJ5Jmhvc3RSZWNoZWNrU2Vjb25kcz0xMA=='\n" (index .data "username") (index .data "password") }}' | kubectl --context $CLUSTER1 apply -f -
```

Create the secret first, which will copy the postgres db secret info on CLUSTER 0 and CLUSTER 1

Deploy the spring app using deployment manifest kubectl on CLUSTER 0 and CLUSTER 1

The image is not loaded to the docker hub, so build the image in local docker and load it to the kind cluster, use imagePullPolicy IfNotPresent in manifest so imported image will be used in deployment.

The spring application is built with Jbang Spring Boot exposes simple endpoint

To build the image, we can resolve the dependencies locally using below jbang export command. Navigate to the spring-app/ and then issue below command. This will create a lib folder and App.jar. The Dockerfile will copy these dependencies and jar so the dependencies are readily available for application to start.

```bash
  jbang export portable app/App.java
```

To build the image in local docker, use below command, note `--progress plain` used to print the output in the console. 

```bash
docker build --progress plain -t my-api-app . -f jar.Dockerfile
```

To load the image to the both clusters, we can use the kind cli itself. And also the manfiest imagePullPolicy is set to IfNotPresent. Note, the --name doesn't include the kuberentes context name instead kind cluster name. The kind- prefix will be added by kind when creating the cluster.


```bash 
kind load docker-image my-api-app:latest --name cluster-dev0
kind load docker-image my-api-app:latest --name cluster-dev1
```

To deploy the application with the api

```bash
kubectl --context $CLUSTER0 -n app apply -f spring-app/k8s/deployment.yaml 
```

```bash
kubectl --context $CLUSTER1 -n app apply -f spring-app/k8s/deployment.yaml 
```

Deploy the routes for the spring app on CLUSTER 0 and CLUSTER 1

```bash
kubectl --context $CLUSTER0 -n app apply -f spring-app/k8s/app_route.yaml
kubectl --context $CLUSTER1 -n app apply -f spring-app/k8s/app_route.yaml
```

From the host now access the app api using curl like below 

-- create input using post on CLUSTER 0 which is 8081 port

```bash  
curl -X POST http://localhost:8081/my-app/products -H "Content-Type: application/json"      -d '{"name": "test-01"}'
```

-- get the data from db on CLUSTER 0 

```bash
curl http://localhost:8081/my-app/products
```

-- get the data from db on CLUSTER 1 -- replica works 

```bash
curl http://localhost:8082/my-app/products
```

The postgres itself is synchronous streaming replica, that is Active - Standby.
The spring application itself is  Active - Active, hikari connection pool is configured to follow the primary database at any point of time.

- The service discovery is used in the data source name and hikari connection configuration helps to keep the connection alive to active leader 

Now perform the switch over and we should see the logs indicating the switch over, for a short moment there would be disrutpion

Manual switch over using kubectl command

- If CLUSTER0 is active and CLUSTER1 is standby - then the standby configuration on the standby side should set to null to promote it to active, the current active cluster should be patched with the standby_host and port so it could be demoted to standby. Below is the switch over command 

```bash
kubectl --context $CLUSTER1 patch postgresql postgres-cluster -n postgres --type=merge -p '{"spec":{"standby": null}}'
```

```bash
kubectl --context $CLUSTER0 patch postgresql postgres-cluster -n postgres --type='merge' -p '{"spec":{"standby":{"standby_host":"postgres-cluster.postgres.svc.cluster.local","standby_port":"5432"}}}'
```

After switch over execute below command to check the status 

```bash
 kubectl --context $CLUSTER0 -n postgres exec pod/postgres-cluster-0 -- patronictl list
 kubectl --context $CLUSTER1 -n postgres exec pod/postgres-cluster-0 -- patronictl list
```

-- If CLUSTER0 is standby and CLUSTER1 is active, then below commands can be used to switch over

```bash
kubectl --context $CLUSTER0 patch postgresql postgres-cluster -n postgres --type=merge -p '{"spec":{"standby": null}}'
```

```bash
kubectl --context $CLUSTER1 patch postgresql postgres-cluster -n postgres --type='merge' -p '{"spec":{"standby":{"standby_host":"postgres-cluster.postgres.svc.cluster.local","standby_port":"5432"}}}'
```
