To create the docker image, we need to use Jbang CLI to resolve dependency, this will help in reducing the start up time for the app

Navigate to the spring-app folder and use below Jbang command to resolve dependency, this will create a lib folder with jars resolved and App.jar

```sh
jbang export portable app/App.java
```

To build the image use docker command to deploy use docker command below

```sh
docker build --progress plain -t my-api-app:1.0.1 . -f jar.Dockerfile
```
