FROM eclipse-temurin:25-jre-jammy
WORKDIR /app

# Copy dependencies
COPY /lib /app/lib

# Copy application code
COPY App.jar /app/App.jar

# get the class name entry point using jar -tvf App.jar uses package 
ENTRYPOINT ["java", "-cp", "/app/App.jar:/app/lib/*", "app.App"]