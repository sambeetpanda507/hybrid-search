FROM alpine:latest

RUN mkdir /app

WORKDIR /app

COPY brokerApp .

RUN chmod +x brokerApp

RUN ls -la -t

CMD ["./brokerApp"]