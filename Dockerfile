FROM alpine:latest 
RUN apk --no-cache add ca-certificates

WORKDIR /root/
ADD build/gopherciser_docker /root/gopherciser
