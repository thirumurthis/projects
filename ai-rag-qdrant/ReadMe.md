- To configure the ollama in docker with WSL2

Use below, navigate to the project path. In this case a level above project directory. 

- run docker in non detached mode

```
docker run -e OLLAMA_KEEP_ALIVE=-1 -v ollama:/root/.ollama -d -p 11434:11434 --name ollama ollama/ollama
```

- exec to the container, we use default embedding. we can use inmemory embedding when we use lang4j

```
docker exec -it ollama sh

# ollama pull nomic-embed-text:v1.5
# ollama pull  mxbai-embed-large

# ollama run llama3.2
```

cd /mnt/c/thiru/edu/ai-models

```
docker run --name qdrant -d -p 6333:6333 -p 6334:6334 \
-v $(pwd)/qdrant:/qdrant/storage \
qdrant/qdrant
```

UI
```
http://localhost:6333/dashboard
```