# Microservices-API-in-Go


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