FROM golang:latest AS builder

WORKDIR /app

RUN export GO111MODULE=on

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -v -o ./grandma_service ./

FROM alpine:latest AS runner

COPY --from=builder /app/ .

EXPOSE 9000 27017 9092

CMD [ "./grandma_service" ]
