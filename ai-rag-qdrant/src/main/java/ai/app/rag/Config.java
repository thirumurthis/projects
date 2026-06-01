package ai.app.rag;

import io.qdrant.client.QdrantClient;
import io.qdrant.client.QdrantGrpcClient;
import org.springframework.ai.embedding.EmbeddingModel;
import org.springframework.ai.embedding.TokenCountBatchingStrategy;
import org.springframework.ai.ollama.OllamaEmbeddingModel;
import org.springframework.ai.vectorstore.VectorStore;
import org.springframework.ai.vectorstore.qdrant.QdrantVectorStore;
import org.springframework.context.annotation.Bean;
import org.springframework.stereotype.Component;

//@Component
public class Config {

    //private QdrantClient qdrantClient;

    //@Bean
    public VectorStore vectorStore(QdrantClient qdrantClient, EmbeddingModel embeddingModel) {
        return QdrantVectorStore.builder(qdrantClient, embeddingModel)
                .collectionName("custom-collection")     // Optional: defaults to "vector_store"
                .initializeSchema(true)                  // Optional: defaults to false
                .batchingStrategy(new TokenCountBatchingStrategy()) // Optional: defaults to TokenCountBatchingStrategy
                .build();
    }

    /*
    @Bean
    public QdrantClient qdrantClient() {
        QdrantGrpcClient.Builder grpcClientBuilder =
                QdrantGrpcClient.newBuilder("localhost", 6334, false);
        //grpcClientBuilder.withApiKey("<QDRANT_API_KEY>");

        return new QdrantClient(grpcClientBuilder.build());
    }
     */

    /*
    @Bean
    public EmbeddingModel embeddingModel(){
        return new OllamaEmbeddingModel()
    }
     */
}
