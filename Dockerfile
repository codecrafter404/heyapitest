FROM oven/bun:1 AS bun
COPY . .
RUN cd frontend && bun install && bun run build

FROM golang:1.25.7-alpine AS go-builder
COPY --from=bun ./backend/ .
RUN cd ./backend/ && go build -o output .

FROM alpine:3.23.3
COPY --from=go-builder ./output .
ENTRYPOINT ["./output"]
