FROM debian:12-slim

RUN apt-get update && apt-get upgrade -y

RUN useradd -u 1000 -m user

COPY ./gemserve /app/gemserve

WORKDIR /app

RUN chmod +x /app/gemserve && \
    chown -R root:root /app && \
    chmod -R 755 /app

USER user
CMD ["/app/gemserve","--listen","0.0.0.0:1965","--root-path","/srv"]
