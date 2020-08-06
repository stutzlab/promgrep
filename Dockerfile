FROM golang:1.14.3-alpine3.11 AS BUILD

RUN apk add gcc build-base

ENV COUNTER1_METRIC_NAME ''
ENV COUNTER1_REGEX ''

ENV GAUGE1_METRIC_NAME ''
ENV GAUGE1_REGEX ''

WORKDIR /app

ADD /go.mod /app/
ADD /go.sum /app/

RUN go mod download

ADD / /app/

RUN go test -v -p 1

WORKDIR /app/cli
RUN go build -o /bin/promgrep
RUN chmod +x /bin/promgrep

CMD [ "/app/dist.sh" ]

