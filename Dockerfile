FROM golang:1.9.2-stretch
ENTRYPOINT [ "/app/entrypoint.sh" ]
EXPOSE 80 443

# Install packages.
RUN apt-get update && \
    apt-get install -y apt-transport-https && \
    apt-get clean
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - && \
    echo "deb https://dl.yarnpkg.com/debian/ stable main" > /etc/apt/sources.list.d/yarn.list && \
    curl -sL https://deb.nodesource.com/setup_9.x | bash - && \
    apt-get install -y nginx nodejs patch unzip yarn && \
    apt-get clean && \
    npm install -g pretty-swag && \
    curl https://glide.sh/get | sh

# Generate SSL certs.
RUN mkdir -p /app/ssl && cd /app/ssl && \
    openssl req -x509 -nodes -newkey rsa:4096 -sha256 \
                -keyout privkey.pem -out fullchain.pem \
                -days 36500 -subj '/CN=localhost' && \
    openssl dhparam -dsaparam -out dhparam.pem 4096

# Build the frontend dependencies.
ADD frontend/package.json frontend/yarn.lock /app/frontend/
RUN cd /app/frontend && \
    yarn install

# Build the frontend bundle.
ADD frontend/ /app/frontend/
RUN cd /app/frontend && \
    IS_PRODUCTION=true ./node_modules/webpack/bin/webpack.js --progress --colors --bail

# Generate swagger docs.
ADD swagger/swagger.yml swagger/config.json /app/swagger/
RUN pretty-swag -c /app/swagger/config.json

# Build the Go dependencies.
RUN curl https://glide.sh/get | sh
ADD glide.yaml glide.lock /go/src/github.com/Nextdoor/conductor/
RUN cd /go/src/github.com/Nextdoor/conductor/ && \
    glide install

# Build Go app.
ADD cmd/ /go/src/github.com/Nextdoor/conductor/cmd
ADD core/ /go/src/github.com/Nextdoor/conductor/core
ADD services/ /go/src/github.com/Nextdoor/conductor/services
ADD shared/ /go/src/github.com/Nextdoor/conductor/shared
RUN go build -o /app/conductor /go/src/github.com/Nextdoor/conductor/cmd/conductor/conductor.go

# Add static resources.
ADD resources/ /app
