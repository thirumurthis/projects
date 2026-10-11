///usr/bin/env jbang "$0" "$@" ; exit $?


package app;
//JAVA 25+

//RUNTIME_OPTIONS -Dspring.config.location=file:./config/application.yaml

//DEPS org.springframework.boot:spring-boot-dependencies:4.0.1@pom
//DEPS org.springframework.boot:spring-boot-starter-web
//DEPS org.springframework.boot:spring-boot-starter-data-jpa
//DEPS org.postgresql:postgresql:42.7.13
//DEPS org.slf4j:slf4j-simple:2.0.13
//DEPS org.slf4j:slf4j-api:2.0.13

import jakarta.persistence.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@SpringBootApplication
@RestController
@RequestMapping("/products")
public class App {

      // Define the SLF4J logger instance
    private static final Logger log = LoggerFactory.getLogger(App.class);

    public static void main(String[] args) {
        SpringApplication.run(App.class, args);
    }

    private final ProductRepository repository;

    public App(ProductRepository repository) {
        this.repository = repository;
    }

    @GetMapping
    public List<Product> getAll() {
         log.info("Received GET request to fetch all products.");

         List<Product> products = repository.findAll();
         log.info("Fetched {} products from the database. products {}", products.size(), products);

        return products;
    }

    @PostMapping
    public Product create(@RequestBody Product product) {
        log.info("Received POST request to create a product. Request body: {}", product);
        
        Product savedProduct = repository.save(product);
        
        // Log the resulting values returned after the database transaction
        log.info("Successfully saved product to repository: {}", savedProduct);
        return savedProduct;
    }
}

@Entity
@Table(name = "products")
class Product {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    private String name;

    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }

    @Override
    public String toString() {
        return "Product{id=" + id +", name='" + name + "'}";
    }
}

@Repository
interface ProductRepository extends JpaRepository<Product, Long> {
}
