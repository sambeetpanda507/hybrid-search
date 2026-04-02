FROM alpine:latest

RUN mkdir /app

WORKDIR /app

COPY searchApp .

COPY internal/database/migrations.sql /app/internal/database/

COPY internal/assets /app/internal/assets/

RUN chmod +x searchApp

RUN ls -la -t

CMD ["./searchApp"]