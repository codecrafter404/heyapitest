FROM oven/bun:1 AS bun
WORKDIR /work/
COPY . .
RUN cd frontend && bun install && bun run build 

FROM golang:1.25.7-alpine AS go-builder
WORKDIR /work/
COPY --from=bun /work/backend/ .
RUN cd ./backend/ && go build -o output .

FROM alpine:3.23.3
COPY --from=go-builder ./output .
ENTRYPOINT ["./output"]
