FROM alpine:latest
LABEL app=toadlester
LABEL org.opencontainers.image.source=https://github.com/maroda/toadlester
WORKDIR /app
COPY toadlester .
EXPOSE 8899
CMD ["./toadlester"]