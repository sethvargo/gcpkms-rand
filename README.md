# Google Cloud KMS Go io.Reader and rand.Source

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://go.pkg.dev/github.com/sethvargo/gcpkms-rand)

This package provides a struct that implements Go's `io.Reader` and `math/rand.Source` interfaces, using Google Cloud KMS HSMs to generate entropy.

## Usage

```go
r, err := gcpkmsrand.NewReader("projects/my-project/locations/us-east1")
if err != nil {
  // handle error
}

// Directly
b := make([]byte, 32)
if _, err := r.Read(b); err != nil {
  // handle error
}

// Via the math package
rnd := rand.New(r)
rnd.Uint32()
```

## Limitations
-   The maximum number of random bytes is 1024 at this time.
