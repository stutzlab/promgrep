FROM golang:1.14.3-alpine3.11 AS BUILD

ENV COUNTER1_METRIC_NAME ''
ENV COUNTER1_REGEX ''

ENV GAUGE1_METRIC_NAME ''
ENV GAUGE1_REGEX ''

WORKDIR /app

ADD /go.mod /app/
ADD /go.sum /app/

RUN go mod download

ADD / /app/
# RUN echo "TEST stats" && cd /app/stats && go test -v
# RUN echo "TEST detectors" && cd /app/detectors && go test -v

RUN go build -o /bin/stdin2prometheus
RUN chmod +x /bin/stdint2prometheus

CMD [ "/app/startup.sh" ]

