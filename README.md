# certpress [![Build Status](https://travis-ci.org/dylanmei/certpress.svg?branch=master)](https://travis-ci.org/dylanmei/certpress)

Because we can't just have PEM-encoded things.

### Usage

Why would you use this?

### Example

Download PEM certificates and convert to PKCS12.

```
certpress \
  --server.key=s3://cat-bucket/cat-key.pem \
  --server.certificate=s3://cat-bucket/cat.pem \
  --server.certificate-authority=$HOME/src/ca.pem \
  --server.password=biscuit \
  --replication.key=s3://dog-bucket/dog-key.pem \
  --replication.certificate=s3://dog-bucket/dog.pem \
  --replication.certificate-authority=$HOME/src/ca.pem \
  --replication.password=caramel
```

If, heaven forbid, one had to further convert to JKS.

```
keytool -noprompt -importkeystore \
  -srckeystore server.pkcs12 -srcstoretype PKCS12 -srcstorepass "biscuit" \
  -deststorepass "biscuit" -destkeypass "biscuit" -destkeystore server-keystore.jks

keytool -noprompt -importkeystore \
  -srckeystore replication.pkcs12 -srcstoretype PKCS12 -srcstorepass "caramel" \
  -deststorepass "caramel" -destkeypass "caramel" -destkeystore replication-keystore.jks
```

And since we've taken it this far, extract the ca-certificate from PKCS12 and import into a separate "truststore".

```
openssl pkcs12 -nokeys -cacerts -nodes \
  -in server.pkcs12 -out server-ca-certificate.pem \
  -password pass:biscuit
keytool -import -noprompt \
  -file server-ca-certificate.pem \
  -keystore server-truststore.jks \
  -storepass "biscuit"

openssl pkcs12 -nokeys -cacerts -nodes \
  -in replication.pkcs12 \
  -out replication-ca-certificate.pem \
  -password pass:caramel
keytool -import -noprompt \
  -file replication-ca-certificate.pem \
  -keystore replication-truststore.jks \
  -storepass "caramel"
```
