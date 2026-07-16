# ---- Stage 1: Builder ----
# This stage installs all necessary tools and builds the application.
FROM golang:1.25-alpine AS builder

# Install build dependencies: Node.js, npm, git (for Go modules)
RUN apk add --no-cache nodejs npm git

# Set the working directory inside the container
WORKDIR /app

# Copy dependency definition files and install dependencies
# This leverages Docker's layer caching.
COPY go.mod go.sum ./
RUN go mod download

COPY package.json package-lock.json ./
RUN npm install

# Copy the rest of the application source code
COPY . .

# --- Run Build Steps ---

# 1. Install and run templ
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1020
RUN /go/bin/templ generate

# 2. Install and run Tailwind CSS
RUN npm install -g tailwindcss@3.4.14
RUN tailwindcss -c ./tailwind.config.js -i ./in.css -o ./app/assets/styles/main.css --minify

# 3. Compile the Go binary for a minimal Linux environment
# The -a flag forces rebuilding of packages that are already up-to-date.
# -ldflags "-s -w" strips debugging information, reducing binary size.
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -o /controlroom ./app

# ---- Stage 2: Final Image ----
# This stage creates the final, minimal image for production.
FROM alpine:latest

# Add Certificate Authority root certificates for making HTTPS requests
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the compiled binary from the 'builder' stage
COPY --from=builder /controlroom .

# Expose the port the app runs on (optional, but good practice)
EXPOSE 3080

# The command to run the application
CMD ["./controlroom"]