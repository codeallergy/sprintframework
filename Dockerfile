FROM codeallergy/ubuntu-golang as builder

ARG VERSION
ARG BUILD

WORKDIR /go/src/github.com/codeallergy/sprintframework
ADD . .

ENV GONOSUMDB github.com

RUN go build -o /sprint -v -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"

FROM ubuntu:18.04
WORKDIR /app/bin

COPY --from=builder /sprint .

EXPOSE 8080 8443 8444

CMD ["/app/bin/sprint", "run"]

