#docker build -t go-rest-balance-charges .
#docker run -dit --name go-rest-balance-charges -p 3000:3000 go-rest-balance-charges

FROM golang:1.21 As builder

WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go build -o go-rest-balance-charges -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-rest-balance-charges .

CMD ["/app/go-rest-balance-charges"]