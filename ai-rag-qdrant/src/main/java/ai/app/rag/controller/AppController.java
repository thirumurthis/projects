package ai.app.rag.controller;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.chat.client.advisor.vectorstore.QuestionAnswerAdvisor;
import org.springframework.ai.ollama.OllamaChatModel;
import org.springframework.ai.vectorstore.VectorStore;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api")
public class AppController {

    private static final Logger log = LoggerFactory.getLogger(AppController.class);
    private final OllamaChatModel chatModel;
    private final VectorStore vectorStore;

    AppController(OllamaChatModel chatModel, VectorStore vectorStore){
        this.chatModel = chatModel;
        this.vectorStore = vectorStore;
    }
    @PostMapping("/input")
    public String chat(@RequestBody String input){

        log.info("input received - {}",input);

        return ChatClient.builder(chatModel)
                .build()
                .prompt()
                .advisors(QuestionAnswerAdvisor.builder(vectorStore).build())
                .user(input)
                .call()
                .content();
    }

}
