FROM golang:alpine

RUN mkdir /app
WORKDIR /app
COPY . .

CMD ["sh", "-c", "/app/octagon-ql --config /app/${ENVIRONMENT}.toml"]
