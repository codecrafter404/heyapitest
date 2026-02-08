FROM dhi.io/bun:1-alpine3.22 AS bun
COPY . .
RUN cd frontend && bun install && bun build 

FROM golang:1.25.7-alpine AS go-builder
COPY --from=bun ./backend/ .
RUN cd ./backend/ && go build -o output .

FROM alpine:1.25.7
COPY --from=go-builder ./output .
ENTRYPOINT ["./output"]
