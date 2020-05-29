Generates a relationship tree between terms.


rocksdb install

CGO_CFLAGS="-I/rocksdb/rocksdb-6.8.1/include" \
CGO_LDFLAGS="-L${SRCDIR}/rocksdb/rocksdb-6.8.1 -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd" \
  go get github.com/tecbot/gorocksdb