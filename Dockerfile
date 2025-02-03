FROM golang:1.23-bookworm

RUN apt-get update && apt-get upgrade -y

RUN useradd -u 1000 -m user

COPY ./gemserve /app/gemserve

WORKDIR /app

RUN chmod +x /app/gemserve && \
    chown -R user:user /app

USER user
CMD ["/app/gemserve","0.0.0.0:1965"]
