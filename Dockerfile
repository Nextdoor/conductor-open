FROM golang:1.9.2-stretch
ENTRYPOINT [ "/app/entrypoint.sh" ]
EXPOSE 80 443

# Install packages.
RUN curl -sL https://deb.nodesource.com/setup_9.x | bash - && \
    apt-get install -y nginx nodejs patch unzip && \
    apt-get clean && \
    npm install -g pretty-swag@0.1.144

# Generate SSL certs.
RUN mkdir -p /app/ssl && cd /app/ssl && \
    openssl req -x509 -nodes -newkey rsa:4096 -sha256 \
                -keyout privkey.pem -out fullchain.pem \
                -days 36500 -subj '/CN=localhost' && \
    openssl dhparam -dsaparam -out dhparam.pem 4096

# Generate swagger docs.
ADD swagger/swagger.yml swagger/config.json /app/swagger/
RUN pretty-swag -c /app/swagger/config.json


# Set up Go app.
ADD .build /go/src/github.com/Nextdoor/conductor/
RUN go build -o /app/conductor /go/src/github.com/Nextdoor/conductor/cmd/conductor/conductor.go

# Add static resources.
ADD resources/ /app
