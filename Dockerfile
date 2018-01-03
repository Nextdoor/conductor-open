FROM golang:1.7.3-alpine

# Set up necessary generic packages.
RUN apk add --no-cache coreutils openssl

# Set up nginx.
RUN apk add --no-cache nginx && \
    mkdir -p /app/logs

# Set up swagger-ui.
ADD swagger/swagger-ui-dist.patch /app/

# Download swagger-ui at known working version.
# The zip contents are in a subdirectory.
# We only need the dist directory, so we remove everything else.
# Then we copy the directory for the patch, apply it, and clean up.
RUN cd /app && \
    VERSION=2.2.6 && \
    wget -O swagger-ui.zip https://github.com/swagger-api/swagger-ui/archive/v$VERSION.zip && \
    mkdir swagger-ui-dist && \
    unzip swagger-ui.zip -d swagger-ui-dist && \
    SUBDIR=swagger-ui-dist/swagger-ui-$VERSION && \
    mv $SUBDIR/dist/* swagger-ui-dist/ && \
    cp -rf swagger-ui-dist/ swagger-ui-dist-patched/ && \
    patch -p0 < swagger-ui-dist.patch && \
    mv swagger-ui-dist-patched swagger && \
    rm -rf swagger-ui.zip swagger-ui-dist.patch swagger-ui-dist
ADD swagger/swagger.yml /app/swagger/

# Generate SSL certs.
RUN mkdir /app/ssl && cd /app/ssl && \
    openssl req -x509 -nodes -newkey rsa:4096 -sha256 \
                -keyout privkey.pem -out fullchain.pem \
                -days 36500 -subj '/CN=localhost' && \
    openssl dhparam -dsaparam -out dhparam.pem 4096

# Set up Go app.
ADD .build /go/src/github.com/Nextdoor/conductor/
RUN go build -o /app/conductor /go/src/github.com/Nextdoor/conductor/cmd/conductor/conductor.go

ADD resources/ /app

# Compute bundle hash for cache-busting on frontend changes.
RUN BUNDLE_HASH=$(md5sum /app/frontend/gen/bundle.js | awk '{print $1}'); \
    sed "s/<BUNDLE_HASH>/${BUNDLE_HASH}/g" -i /app/frontend/index.html

EXPOSE 80 443
ENTRYPOINT [ "/app/entrypoint.sh" ]
