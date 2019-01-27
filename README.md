# certpress [![Build Status](https://travis-ci.org/dylanmei/certpress.svg?branch=master)](https://travis-ci.org/dylanmei/certpress)

Because we can't just have PEM-encoded things.

### Usage

Why would you use this? Use [jimmidyson/pemtokeystore](https://github.com/jimmidyson/pemtokeystore) instead.

### Example

Fetch PEM certificates and convert to Java KeyStore files.

```
certpress \
  --server.key=cat-key.pem \
  --server.certificate=cat.pem \
  --server.certificate-authority=ca.pem \
  --server.password=biscuit \
  --replication.key=s3://dog-bucket/dog-key.pem \
  --replication.certificate=s3://dog-bucket/dog.pem \
  --replication.certificate-authority=s3://dog-bucket/ca.pem \
  --replication.password=caramel
```
