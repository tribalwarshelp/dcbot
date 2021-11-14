FROM golang:1.17.2-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN go build -o twhelpdcbot .

######## Start a new stage from scratch #######
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file and translations from the previous stage
COPY --from=builder /app/message/translations ./message/translations
COPY --from=builder /app/twhelpdcbot .

ENV APP_MODE=production
ENV GIN_MODE=release
EXPOSE 8080

CMD ./twhelpdcbot
