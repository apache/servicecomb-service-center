FROM golang:alpine
WORKDIR /app
COPY ./go.mod .
ENV GOPROXY https://goproxy.cn
RUN go mod download
COPY . .
CMD ["go", "run", "main.go"]
