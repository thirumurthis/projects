package ai.app.rag;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.ai.document.Document;
import org.springframework.ai.reader.pdf.PagePdfDocumentReader;
import org.springframework.ai.reader.pdf.ParagraphPdfDocumentReader;
import org.springframework.ai.transformer.splitter.TextSplitter;
import org.springframework.ai.transformer.splitter.TokenTextSplitter;
import org.springframework.ai.vectorstore.VectorStore;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.CommandLineRunner;
import org.springframework.core.io.Resource;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Properties;

@Component
public class DocumentIngestionService implements CommandLineRunner {

    private static final Logger log = LoggerFactory.getLogger(DocumentIngestionService.class);

    private final VectorStore vectorStore;

    public DocumentIngestionService(VectorStore vectorStore){
        this.vectorStore = vectorStore;
    }

    @Value("classpath:/docs/streaming_watermarks.pdf")
    private Resource pdfDoc;


    @Override
    public void run(String... args) throws Exception {

        //Properties properties = new Properties();
        //properties.put("DOCKER_DESKTOP","unix:///var/run/docker.sock");
        //System.setProperties(properties);

        var pdfReader = new PagePdfDocumentReader(pdfDoc);//ParagraphPdfDocumentReader(pdfDoc);
        //var pdfReader = new TikaDocumentReader(pdfDoc);
        TextSplitter textSplitter = new TokenTextSplitter(400, 150, 25, 10000, true);

        //List<Document> documents = textSplitter.split(pdfReader.read());

        //documents.forEach(item -> log.info("document - {}",item));
        //vectorStore.accept(documents);
        List<Document> documents = textSplitter.split(pdfReader.read());
        vectorStore.accept(documents);
        log.info("data loaded completed ...");

    }
}
