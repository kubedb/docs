This package is fork of [golang/crypto/pkcs12](https://github.com/golang/crypto/tree/master/pkcs12).
 
**Reason of fork:**

Currently, `pkcs12` can not decode a keystore if it has more than one key-cert pair. Ref: https://github.com/golang/go/issues/14015

**Solution:** 

There is already a pending PR [here](https://github.com/golang/crypto/pull/38) that introduce `DecodeAll` method that solve this issue. This forked
version take the changes from that PR.

> Use original package when https://github.com/golang/crypto/pull/38 is merged. 