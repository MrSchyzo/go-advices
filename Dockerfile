FROM golang:1.15.6-alpine3.13 as building
RUN mkdir /build
ADD * /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o go-advices

FROM alpine:3.13.1
COPY --from=building /build/go-advices ./go-advices
ENTRYPOINT [ "./go-advices" ]
EXPOSE 10000