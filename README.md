# certpress [![Build Status](https://travis-ci.org/dylanmei/certpress.svg?branch=master)](https://travis-ci.org/dylanmei/certpress)

Because we can't just have PEM-encoded things.

### Usage

Why would you use this? Use [jimmidyson/pemtokeystore](https://github.com/jimmidyson/pemtokeystore) instead.

### Example

Fetch PEM certificates and convert to Java KeyStore files.

```
make build example
bin/certpress \
  --server.key=example/server-key.pem \
  --server.certificate=example/server.pem \
  --server.certificate-authority=example/ca.pem \
  --server.password=changeme \
  --truststore.certificate-authority=example/ca.pem

docker-compose up
kafkacat -b localhost:9093 -X security.protocol=SSL -X ssl.ca.location=example/ca.pem -L
```
