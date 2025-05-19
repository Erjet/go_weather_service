FROM golang:1.24.3 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o appmain .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/appmain .
COPY settings.json .
COPY index.html .

EXPOSE 8080

CMD ["./appmain"]