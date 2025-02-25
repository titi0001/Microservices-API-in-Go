# Microservices-API-in-Go

# Building Microservices API in Go

This repository contains the source code for **Building Microservices API in Go**, covering advanced development practices such as authentication, authorization, JWT tokens, and refresh tokens. It also includes techniques for writing reliable code through unit testing and essential concepts such as logging, error handling, and modularization.

With this project, you will be able to develop secure and scalable web applications using Go, following best practices for code organization and reusability.

## 📌 Features Covered

- Implementation of authentication and authorization, including role-based access control (RBAC)
- Generation and management of JWT tokens for secure authentication
- Unit testing strategies, including state-based testing, route and service testing, as well as creating and using mocks and stubs
- Structured logging and efficient error handling
- Code modularization by extracting reusable packages
- Integration of modules into a banking API for authentication and transaction management
- Use of Claims for JWT token parsing and implementation of refresh tokens for continuous and secure access

## 🎯 Target Audience

This project is ideal for developers looking to enhance their skills in Go and REST-based microservices. It is also useful for software professionals who want to learn about building scalable and secure APIs.

## 🚀 How to Use This Repository

Clone the repository and explore the files to understand the implementations and best practices used. If you want to contribute or extend functionalities, feel free to open issues or pull requests.


## Using `reflex` for Live Reloading

During development, it’s useful to have your server automatically restart when you make changes to your code. For this, we can use the `reflex` package, which watches for file changes and restarts your Go application automatically, similar to `nodemon` in Node.js.

### Installing `reflex`

First, install `reflex`:

```bash
go install github.com/cespare/reflex@latest
```

### Configuring reflex
You can use reflex directly from the command line or set up a configuration file.

Using reflex from the command line:

```bash
reflex -r '\.go$' -s -- sh -c "go run main.go"
```
### This command tells reflex to watch all .go files and restart the application whenever a change is detected.

### Using reflex with a configuration file:

```bash
# Reflex config
-r '\.go$' -s -- sh -c "go run main.go"
```