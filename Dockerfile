FROM golang:1.11.13
ENTRYPOINT [ "/app/entrypoint.sh" ]
EXPOSE 80 443

# Install packages.
RUN curl -sL https://deb.nodesource.com/setup_9.x | bash - && \
    apt-get install -y jq nginx nodejs patch unzip && \
    apt-get clean

# Generate SSL certs.
RUN mkdir -p /app/ssl && cd /app/ssl && \
    openssl req -x509 -nodes -newkey rsa:4096 -sha256 \
                -keyout privkey.pem -out fullchain.pem \
                -days 36500 -subj '/CN=localhost' && \
    openssl dhparam -dsaparam -out dhparam.pem 4096

# Generate swagger docs.
RUN apt-get install -y npm && npm install -g pretty-swag@0.1.144
ADD swagger/swagger.yml swagger/config.json /app/swagger/
RUN ls /app/swagger/
RUN cd /app && pretty-swag -c /app/swagger/config.json

# Add awscli
RUN apt-get update && \
    apt-get install -y \
        python3 \
        python3-pip \
        python3-setuptools \
        groff \
        less \
    && pip3 install --upgrade pip \
    && apt-get clean

RUN pip3 --no-cache-dir install --upgrade awscli


# Set up Go app.
ADD .build /src/github.com/Nextdoor/conductor/
ADD .build /go/src/github.com/Nextdoor/conductor/
RUN cd /src/github.com/Nextdoor/conductor/ && go build -o /app/conductor /src/github.com/Nextdoor/conductor/cmd/conductor/conductor.go

# Add static resources.
ADD resources/ /app
