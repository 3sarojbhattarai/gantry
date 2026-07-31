package docker

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

// frame builds one multiplexed Docker stream frame: an 8-byte header (stream
// type + big-endian payload length) followed by the payload.
func frame(streamType byte, payload string) []byte {
	h := make([]byte, streamHeaderSize)
	h[0] = streamType
	binary.BigEndian.PutUint32(h[4:streamHeaderSize], uint32(len(payload)))
	return append(h, payload...)
}

func demuxString(t *testing.T, in []byte) string {
	t.Helper()
	var out bytes.Buffer
	if err := demux(&out, bytes.NewReader(in)); err != nil {
		t.Fatalf("demux: unexpected error: %v", err)
	}
	return out.String()
}

func TestDemuxInterleavesStdoutAndStderr(t *testing.T) {
	in := bytes.Join([][]byte{
		frame(streamStdout, "hello "),
		frame(streamStderr, "cruel "),
		frame(streamStdout, "world"),
	}, nil)
	if got := demuxString(t, in); got != "hello cruel world" {
		t.Fatalf("got %q, want %q", got, "hello cruel world")
	}
}

func TestDemuxSkipsStdin(t *testing.T) {
	in := bytes.Join([][]byte{
		frame(streamStdout, "keep"),
		frame(streamStdin, "DROP THIS"),
		frame(streamStdout, "me"),
	}, nil)
	if got := demuxString(t, in); got != "keepme" {
		t.Fatalf("got %q, want %q", got, "keepme")
	}
}

func TestDemuxZeroLengthFrame(t *testing.T) {
	in := bytes.Join([][]byte{
		frame(streamStdout, "a"),
		frame(streamStdout, ""), // zero-length payload, header only
		frame(streamStdout, "b"),
	}, nil)
	if got := demuxString(t, in); got != "ab" {
		t.Fatalf("got %q, want %q", got, "ab")
	}
}

func TestDemuxEmptyInput(t *testing.T) {
	if got := demuxString(t, nil); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestDemuxPayloadLargerThanCopyBuffer(t *testing.T) {
	// Exercise a single payload bigger than io.Copy's internal buffer to catch
	// short-read bugs in the length-bounded copy.
	big := strings.Repeat("x", 200_000)
	if got := demuxString(t, frame(streamStdout, big)); got != big {
		t.Fatalf("got %d bytes, want %d", len(got), len(big))
	}
}

func TestDemuxTruncatedHeader(t *testing.T) {
	var out bytes.Buffer
	// Four bytes is a partial header: a boundary read that ends mid-header is
	// malformed, not a clean end of stream.
	if err := demux(&out, bytes.NewReader([]byte{1, 0, 0, 0})); err == nil {
		t.Fatal("expected error on truncated header, got nil")
	}
}

func TestDemuxTruncatedPayload(t *testing.T) {
	var out bytes.Buffer
	// Header announces 10 payload bytes, but only 2 are provided.
	h := make([]byte, streamHeaderSize)
	h[0] = streamStdout
	binary.BigEndian.PutUint32(h[4:streamHeaderSize], 10)
	in := append(h, 'h', 'i')
	if err := demux(&out, bytes.NewReader(in)); err == nil {
		t.Fatal("expected error on truncated payload, got nil")
	}
}
