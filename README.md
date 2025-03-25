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

### 1. Using Reflex from the Command Line

To run Reflex directly from the terminal, execute the following command in the project root:
```bash
reflex -r '\.go$' -s -- sh -c "go run ./cmd/api/main.go"
```
Explanation of the parameters:

- -r '\.go$': A regular expression pattern that tells Reflex to monitor only files with the .go extension.
- -s: Starts the command immediately (without waiting for an initial change).
- --: Separates Reflex arguments from the command to be executed.
- sh -c "go run ./cmd/api/main.go": The command Reflex will run whenever a .go file is modified. Here, it starts - the Go server from cmd/api/main.go.

This command will monitor all .go files in the project and restart the server automatically whenever a change is detected.

Example: Monitoring specific files
If you want to monitor only files in a specific directory (e.g., api), adjust the pattern:

```bash
reflex -r 'api/.*\.go$' -s -- sh -c "go run ./cmd/api/main.go"
```

### 2. Using Reflex with a Configuration File
You can also configure Reflex using a reflex.conf file in the project root. This is useful for projects with more complex setups or to share the configuration with other developers.

#### 2.1 Create a file named reflex.conf in the project root:
```bash
# reflex.conf
-r '\.go$' -s -- sh -c "go run ./cmd/api/main.go"
```

#### 2.2 Run Reflex with the configuration file:
```bash
reflex -c reflex.conf
```
Explanation:

- The reflex.conf file contains the same configuration as the command-line version.
- -c reflex.conf: Tells Reflex to use the specified configuration file.

### 3. Installing Reflex
If Reflex is not already installed, you can install it with the following command:
```bash
go install github.com/cespare/reflex@latest
```

Ensure that the GOPATH/bin directory (or $HOME/go/bin by default) is in your PATH so the reflex command is recognized:

```bash
export PATH=$PATH:$HOME/go/bin
```
Alternatively, the project includes a pre-built Reflex binary in the bin directory (bin/reflex), but installing it globally as shown above ensures you are using the latest version.


### 4. Integrating Reflex with the Project
To use Reflex with this project, ensure that the .env file is properly configured, as main.go relies on environment variables such as LOCAL_HOST, AUTH_LOCAL_HOST, and JWT_SECRET_KEY. Example .env file:
```bash
# .env
LOCAL_HOST=localhost:8080
AUTH_LOCAL_HOST=localhost:8181
JWT_SECRET_KEY=your-secure-secret-key-here
```
Then, run Reflex as described above to start the server with automatic reloading.

### 5. Additional Tips
- Monitoring Specific Files: To monitor only files in a specific directory (e.g., api), adjust the pattern:
```bash
reflex -r 'api/.*\.go$' -s -- sh -c "go run ./cmd/api/main.go"
```

- Running in the Background: Use & to run Reflex in the background:
```bash
reflex -r '\.go$' -s -- sh -c "go run ./cmd/api/main.go" &
```

- Stopping Reflex: To stop Reflex, find the process and terminate it:
```bash
killall reflex
```