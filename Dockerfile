FROM golang:1.24.5 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod downoload
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/app

FROM gcr.io/distroless/base-debian12
COPY --from=build /bin/app /app
COPY assets /assets
COPY config.local.yaml /config.yaml
WORKDIR /
ENV CONFIG=/config.yaml
EXPOSE 8081
ENTRYPOINT ["/app"]




