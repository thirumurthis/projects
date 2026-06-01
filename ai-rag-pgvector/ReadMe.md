- To configure the ollama in docker with WSL2

Use below, navigate to the project path. In this case a level above project directory. 

- run docker in non detached mode

```
docker run -d -v ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama
```

- exec to the container, we use default embedding. we can use inmemory embedding when we use lang4j

```
docker exec -it ollama sh

# ollama pull  mxbai-embed-large

# ollama run llama3.2
```

- The postgresDb should be executed using below command, navigate to the project level

```
docker-compose up
```