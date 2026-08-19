package journal

import "testing"

func FuzzDecodeRecord(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("invalid-record"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _, _ = DecodeRecord(data)
	})
}
