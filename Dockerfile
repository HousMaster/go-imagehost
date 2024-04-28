FROM alpine:latest

# Install Go
RUN wget https://golang.org/dl/go1.22.2.linux-amd64.tar.gz && \
    tar -C / -xzf go1.22.2.linux-amd64.tar.gz && \
    rm go1.22.2.linux-amd64.tar.gz

ENV PATH="/go/bin:${PATH}"

COPY . /app
WORKDIR /app

RUN go build -o ./bin/app

CMD [ "./bin/app" ]