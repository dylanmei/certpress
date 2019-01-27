#!/bin/bash -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
PASS=changeme

setup_ca() {
  echo "SETUP CA"
  pushd $DIR
  rm -f ca.index* ca.srl* ca-key.pem ca.pem

  touch ca.index
  touch ca.index.attr
  echo '01' > ca.srl

  openssl req -newkey rsa:2048 -x509 -keyout $DIR/ca-key.pem -out $DIR/ca.pem -days 365 \
    -subj '/CN=CERTPRESS/OU=CERTPRESS/O=CERTPRESS/C=US' \
    -passin pass:$PASS -passout pass:$PASS

  popd
}

setup_certificate() {
  name=$1

  printf "\nSETUP $name CERTIFICATE\n"
  pushd $DIR
  rm -f $name-key.pem $name.pem $name.csr

  # private key and csr
  openssl req -new -newkey rsa:2048 -nodes \
    -batch -config $name.conf \
    -keyout $name-key.pem -out $name.csr

  # sign it
  openssl ca -config ca.conf -batch \
    -in $name.csr -passin "pass:$PASS" \
    -extfile $name.extensions \
    -out $name.pem
  
  popd
}

cleanup() {
  rm -f $DIR/*.srl* $DIR/*.index* $DIR/*.csr $DIR/*.pem
}

cleanup
setup_ca
setup_certificate "server"
printf "\nDONE"
