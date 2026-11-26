# --- SINGLE STAGE ---
# FROM golang:1.23-alpine

# WORKDIR /app

# COPY go.mod ./
# RUN go mod download

# COPY . .

# RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# EXPOSE 8080

# CMD ["./main"]



# --- MULTI STAGE ---

# --- STAGE 1: Builder ---
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Unduh dependensi
COPY ./app/go.mod ./
RUN go mod download


COPY ./app/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .


# --- STAGE 2: Final Image ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/


COPY --from=builder /build/main .

EXPOSE 8080

CMD ["./main"]