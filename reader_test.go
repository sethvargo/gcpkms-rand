package gcpkmsrand_test

import (
	"crypto/rand"
	"math/big"
	mathrand "math/rand"
	"os"
	"strings"
	"testing"

	gcpkmsrand "github.com/sethvargo/gcpkms-rand"
)

func testReader(tb testing.TB) *gcpkmsrand.Reader {
	tb.Helper()

	r, err := gcpkmsrand.NewReader(os.Getenv("GOOGLE_CLOUD_KMS_LOCATION"))
	if err != nil {
		tb.Fatal(err)
	}
	return r
}

func TestReader_Read(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		b    []byte
		err  string
	}{
		{
			name: "overflow",
			b:    make([]byte, 2<<33),
			err:  "cannot request more than",
		},
		{
			// okay because upsize to 8
			name: "too_small",
			b:    make([]byte, 0),
		},
		{
			name: "default",
			b:    make([]byte, 16),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := testReader(t)

			n, err := r.Read(tc.b)
			if err != nil {
				if tc.err == "" {
					t.Fatal(err)
				} else if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected %q to contain %q", err.Error(), tc.err)
				}
				return
			} else {
				if tc.err != "" {
					t.Errorf("expected error, but got nothing")
				}
			}

			if got, want := n, len(tc.b); got != want {
				t.Errorf("not enough bytes (got %d, want %d)", got, want)
			}
		})
	}
}

func TestReader_RandInt(t *testing.T) {
	t.Parallel()

	r := testReader(t)

	b, err := rand.Int(r, big.NewInt(27))
	if err != nil {
		t.Fatal(err)
	}
	_ = b.Int64()
}

func TestReader_RandSource(t *testing.T) {
	t.Parallel()

	r := testReader(t)

	rnd := mathrand.New(r)
	_ = rnd.Int()
}

func ExampleNewReader() {
	r, err := gcpkmsrand.NewReader("projects/my-project/locations/us-east1")
	if err != nil {
		// handle error
	}
	_ = r
}
