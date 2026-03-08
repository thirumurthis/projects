### Simple JBang project spring-boot

For more details refer the [link](https://github.com/thirumurthis/Learnings/blob/main/K8s/stakater_reloader/stakater_jbang_spring_app/stakater_reloader.md) 

- Install the stakater

```
helm repo add stakater https://stakater.github.io/stakater-charts
helm repo update
helm install reloader stakater/reloader
```
- docker build command

```
 docker build -t jbang-spring-app .
```
or
```
 docker build --progress=plain -t jbang-spring-app .
```

- To run the application

```
 docker run -d -p 8080:8085 jbang-spring-app
```

- To deploy the image with kind cluster 

```
 kind load docker-image jbang-spring-app:latest --name test
```

To patch the config map and since the skataer reloader is installed, upon patching 

```
kubectl -n jbang-app patch configmap jbang-app-config --type merge \
  -p '{
    "data": {
      "application.yaml": "app:\n  message: \"new-message\""
    }
  }'
```
