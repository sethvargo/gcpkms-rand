package gcpkmsrand

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	mathrand "math/rand"

	kms "cloud.google.com/go/kms/apiv1"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

// Ensure Reader implements Read and Close.
var _ io.ReadCloser = (*Reader)(nil)

// Ensure Reader is a math/rand.Source.
var _ mathrand.Source = (*Reader)(nil)

// Reader is the random reader.
type Reader struct {
	client   *kms.KeyManagementClient
	location string
}

// NewReader creates a new random reader. It establishes a connection to Google
// Cloud KMS with the provided parameters. The location is specified as the
// "project/<project>/locations/<location>".
func NewReader(location string) (*Reader, error) {
	ctx := context.Background()
	client, err := kms.NewKeyManagementClient(ctx,
		option.WithUserAgent("sethvargo:gcpkms-rand/0.1.0"))
	if err != nil {
		return nil, fmt.Errorf("failed to create kms client: %w", err)
	}

	return &Reader{
		client:   client,
		location: location,
	}, nil
}

// Read implements the io.Reader interface to be used as a random generator.
func (r *Reader) Read(b []byte) (int, error) {
	// The API actually restricts to 1024, but we don't want to hardcode that here
	// in case it's increased in the future.
	numWantedBytes := len(b)
	if numWantedBytes > math.MaxInt32 {
		return 0, fmt.Errorf("cannot request more than %d bytes", math.MaxInt32)
	}

	// It's okay if the user requested less, but the minimum for the API is 8
	// bytes.
	if numWantedBytes < 8 {
		numWantedBytes = 8
	}

	ctx := context.Background()
	result, err := r.client.GenerateRandomBytes(ctx, &kmspb.GenerateRandomBytesRequest{
		Location:        r.location,
		LengthBytes:     int32(numWantedBytes),
		ProtectionLevel: kmspb.ProtectionLevel_HSM,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	numRecvBytes := len(result.Data)
	if numRecvBytes < numWantedBytes {
		return 0, fmt.Errorf("not enough bytes returned (%d wanted %d)",
			numRecvBytes, numWantedBytes)
	}

	for i := range b {
		b[i] = result.Data[i]
	}
	return len(b), nil
}

// Close closes the reader and the connection to upstream resources.
func (r *Reader) Close() error {
	return r.client.Close()
}

// Int63 implements the Source interface.
func (r *Reader) Int63() int64 {
	n, err := rand.Int(r, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(err)
	}
	return n.Int64()
}

// Uint64 implements the Source64 interface.
func (r *Reader) Uint64() uint64 {
	n, err := rand.Int(r, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(err)
	}
	return n.Uint64()
}

// Seed is required for the Source interface, but seeding is not supported.
func (r *Reader) Seed(_ int64) {}
