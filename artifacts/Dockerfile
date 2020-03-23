# VERSION 0.2
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
ARG PORT=9090
ADD gopherciser /root/
ADD testrunner.sh /root/

EXPOSE $PORT

RUN sed -i -e 's/METRICPORT/'$PORT'/g' testrunner.sh

ENTRYPOINT [ "./testrunner.sh" ]