FROM debian:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get dist-upgrade -y && apt-get install -y ca-certificates bash procps

COPY wp-cookie-verify /
 
ENTRYPOINT [ "/wp-cookie-verify" ]
