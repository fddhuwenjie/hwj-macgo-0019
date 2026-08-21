package journal

import (
	"bytes"
	"errors"
	"testing"
)

// TestPayloadTamperIsRejected confirms the integrity fix: swapping Payload
// bytes on disk while keeping Length/headers identical must now be detected
// via the checksum (it previously decoded as valid, replaying wrong state).
func TestPayloadTamperIsRejected(t *testing.T) {
	rec := Record{Sequence: 1, Version: 7, Length: 3, Payload: []byte("abc"), Timestamp: 42}
	rec.Checksum = rec.computeChecksum()
	raw := rec.Encode()

	tampered := append([]byte(nil), raw...)
	tampered[recordHeaderLen] ^= 0xFF // flip first payload byte; Length unchanged

	if _, _, err := DecodeRecord(tampered); err == nil {
		t.Fatalf("tampered payload decoded as valid; expected rejection")
	}
}

// TestHeaderTamperIsRejected confirms the second part of the report: every
// header field covered by the checksum is still authenticated, so tampering
// any one of them is rejected.
func TestHeaderTamperIsRejected(t *testing.T) {
	cases := map[string]int{
		"Sequence":  4,  // bytes [4:12)
		"Version":   12, // bytes [12:20)
		"Length":    20, // bytes [20:24)
		"Timestamp": 24, // bytes [24:32)
	}
	for field, off := range cases {
		rec := Record{Sequence: 1, Version: 7, Length: 3, Payload: []byte("abc"), Timestamp: 42}
		rec.Checksum = rec.computeChecksum()
		raw := rec.Encode()

		tampered := append([]byte(nil), raw...)
		tampered[off] ^= 0xFF
		if _, _, err := DecodeRecord(tampered); err == nil {
			t.Fatalf("%s tamper decoded as valid; expected rejection", field)
		}
	}
}

// TestRoundtripPreservesPayload is a regression guard: legitimate records
// still encode/decode intact after the checksum change.
func TestRoundtripPreservesPayload(t *testing.T) {
	want := Record{Sequence: 5, Version: 9, Length: 4, Payload: []byte("good"), Timestamp: 100}
	want.Checksum = want.computeChecksum()
	got, n, err := DecodeRecord(want.Encode())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if n != len(want.Encode()) {
		t.Fatalf("consumed %d bytes, frame %d", n, len(want.Encode()))
	}
	if !bytes.Equal(got.Payload, want.Payload) || got.Sequence != want.Sequence ||
		got.Version != want.Version || got.Length != want.Length ||
		got.Timestamp != want.Timestamp || got.Checksum != want.Checksum {
		t.Fatalf("roundtrip mismatch:\n got  %+v\n want %+v", got, want)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("re-validate failed: %v", err)
	}
}

// TestEmptyPayloadRoundtrip guards the zero-length edge: an empty Payload is
// still bound by the checksum (it hashes the header with no trailing bytes).
func TestEmptyPayloadRoundtrip(t *testing.T) {
	rec := Record{Sequence: 1, Version: 1, Length: 0, Payload: []byte{}, Timestamp: 7}
	rec.Checksum = rec.computeChecksum()
	raw := rec.Encode()

	// Append junk after the empty record; DecodeRecord must stop at the frame.
	if _, n, err := DecodeRecord(append(raw, 0xFF)); err != nil {
		t.Fatalf("decode empty record: %v", err)
	} else if n != len(raw) {
		t.Fatalf("consumed %d, frame %d", n, len(raw))
	}

	// Injecting a single byte where Payload was empty must fail checksum.
	injected := append(append([]byte(nil), raw[:recordHeaderLen]...), 0xAA)
	injected = append(injected, rec.Checksum[:]...)
	if _, _, err := DecodeRecord(injected); err == nil || !errors.Is(err, ErrCorruptRecord) {
		t.Fatalf("injected byte into empty payload decoded; want ErrCorruptRecord, got %v", err)
	}
}
