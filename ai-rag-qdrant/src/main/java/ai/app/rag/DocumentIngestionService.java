package ai.app.rag;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.ai.document.Document;
import org.springframework.ai.reader.TextReader;
import org.springframework.ai.reader.pdf.PagePdfDocumentReader;
import org.springframework.ai.transformer.splitter.TokenTextSplitter;
import org.springframework.ai.vectorstore.VectorStore;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.CommandLineRunner;
import org.springframework.core.io.FileSystemResource;
import org.springframework.core.io.Resource;
import org.springframework.stereotype.Component;

import java.io.File;
import java.util.List;

@Component
public class DocumentIngestionService implements CommandLineRunner {

    private static final Logger log = LoggerFactory.getLogger(DocumentIngestionService.class);

    private final VectorStore vectorStore;

    public DocumentIngestionService(VectorStore vectorStore){
        this.vectorStore = vectorStore;
    }

    @Value("classpath:docs/*")
    //@Value("${file.directory.path}")
    private Resource[] docsPath;


    @Override
    public void run(String... args) throws Exception {

        //Properties properties = new Properties();
        //properties.put("DOCKER_DESKTOP","unix:///var/run/docker.sock");
        //System.setProperties(properties);

        for(Resource file : docsPath ){

            loadTextToQdrant(file.getFile().getAbsolutePath());
            //loadPdfToQdrant(file.getFile().getAbsolutePath());
        }

        //var pdfReader = new PagePdfDocumentReader(pdfDoc);//ParagraphPdfDocumentReader(pdfDoc);
        //var pdfReader = new TikaDocumentReader(pdfDoc);
        //TextSplitter textSplitter = new TokenTextSplitter(400, 150, 25, 10000, true);

        //List<Document> documents = textSplitter.split(pdfReader.read());

        //documents.forEach(item -> log.info("document - {}",item));
        //vectorStore.accept(documents);
        //List<Document> documents = textSplitter.split(pdfReader.read());
        //vectorStore.accept(documents);
        //log.info("data loaded completed ...");

    }


    public void loadPdfToQdrant(String filePath) {
        // 1. Read the PDF from the local file path
        String resourcePath = "file:" + filePath;
        log.info("file name from the folder {}",resourcePath);
        PagePdfDocumentReader pdfReader = new PagePdfDocumentReader(resourcePath);
        List<Document> rawDocuments = pdfReader.get();

        // 2. Split text into smaller chunks for better embedding accuracy
        TokenTextSplitter textSplitter = new TokenTextSplitter();
        List<Document> splitDocuments = textSplitter.apply(rawDocuments);

        // 3. Write chunks and their vectors automatically to Qdrant
        vectorStore.accept(splitDocuments);
    }

    public void loadTextToQdrant(String filePath) {
        // 1. Load text file
        String resourcePath = "file:" + filePath;
        log.info("file name from the folder {}",resourcePath);
        // 1. Load the text file
        TextReader textReader = new TextReader(new FileSystemResource(new File(filePath)));
        List<Document> documents = textReader.get();

        // 2. Tokenize/Chunk the documents based on token counts
        // TokenTextSplitter(chunkSize, minChunkSizeChars, minChunkLengthToEmbed,
        //                    maxNumChunks, keepSeparator, punctuationMarks)
        TokenTextSplitter textSplitter = new TokenTextSplitter(
                800,
                200,
                200,
                200,
                true,
                List.of(',',' ',':')
        );

        List<Document> splitDocuments = textSplitter.apply(documents);

        // 3. Generate chunks and their vectors automatically to Qdrant
        vectorStore.accept(splitDocuments);
    }
}
