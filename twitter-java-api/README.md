# Twitter Clone Backend

A robust REST API for the Twitter clone application, built with Java and Spring Boot.

## Features

- **Authentication**: JWT-based authentication and authorization.
- **REST API**: Endpoints for users, tweets, interactions, notifications, and search.
- **Database**: PostgreSQL integration with Spring Data JPA.
- **Security**: Spring Security configuration.
- **Validation**: Input validation using Bean Validation.
- **Rate Limiting**: Protects APIs using Bucket4j (100 requests/minute/IP).
- **Testing**: Unit and integration tests with JUnit and Mockito.

## Tech Stack

- **Framework**: [Spring Boot 3](https://spring.io/projects/spring-boot)
- **Language**: Java 17+
- **Database**: PostgreSQL
- **Build Tool**: Maven
- **ORM**: Spring Data JPA / Hibernate
- **Security**: Spring Security / JWT

## Getting Started

### Prerequisites

- Java JDK 17+
- Maven 3.8+
- PostgreSQL database

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/yourusername/twitter-clone-azure.git
    cd twitter-clone-azure/twitter-backend
    ```

2.  Configure database:
    Update `src/main/resources/application.properties` (or `application.yml`) with your PostgreSQL connection details:
    ```properties
    spring.datasource.url=jdbc:postgresql://localhost:5432/twitter_db
    spring.datasource.username=postgres
    spring.datasource.password=password
    ```

3.  Build the project:
    ```bash
    mvn clean package
    ```

4.  Run the application:
    ```bash
    mvn spring-boot:run
    ```
    The server will start on port `8080`.

## API Documentation

- Swagger UI: `http://localhost:8080/swagger-ui.html` (if enabled)
- API Base URL: `http://localhost:8080/api/v1`

## Deployment

This application is deployed via Azure Container Apps using GitHub Actions.
Please refer to the [Root README](../README.md#deployment-azure--github-actions) for detailed deployment instructions and CI/CD setup.
