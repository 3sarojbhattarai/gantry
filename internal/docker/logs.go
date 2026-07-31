package docker

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Docker multiplexes a container's stdout and stderr into a single stream when
// the container has no TTY. Each chunk is framed with an 8-byte header:
//
//	[0]     stream type (0=stdin, 1=stdout, 2=stderr)
//	[1:4]   reserved (zero)
//	[4:8]   payload length, big-endian uint32
//
// followed by exactly that many payload bytes. demux reassembles the payloads
// in order. We implement it ourselves — rather than lean on the SDK — so the
// framing logic is domain code and testable against hand-crafted frames with
// no daemon. Off-by-one bugs here corrupt output silently, so it has dedicated
// tests.
const streamHeaderSize = 8

const (
	streamStdin  = 0
	streamStdout = 1
	streamStderr = 2
)

// demux reads Docker's multiplexed stream from src and writes the stdout and
// stderr payloads, in the order they arrive, to dst. stdin frames are skipped.
// It returns nil on a clean end of stream (a header boundary followed by EOF)
// and an error if the stream ends mid-frame or dst fails.
func demux(dst io.Writer, src io.Reader) error {
	header := make([]byte, streamHeaderSize)
	for {
		if _, err := io.ReadFull(src, header); err != nil {
			if err == io.EOF {
				return nil // clean boundary
			}
			return fmt.Errorf("gantry: reading log frame header: %w", err)
		}
		size := int64(binary.BigEndian.Uint32(header[4:streamHeaderSize]))
		if size == 0 {
			continue
		}
		out := dst
		if header[0] == streamStdin {
			out = io.Discard
		}
		if _, err := io.CopyN(out, src, size); err != nil {
			return fmt.Errorf("gantry: reading log frame payload: %w", err)
		}
	}
}

// demuxReadCloser adapts the demux pipeline to io.ReadCloser: callers read the
// reassembled stream, and Close tears down both the pipe and the underlying
// source stream.
type demuxReadCloser struct {
	pr  *io.PipeReader
	src io.Closer
}

func (d *demuxReadCloser) Read(p []byte) (int, error) { return d.pr.Read(p) }

func (d *demuxReadCloser) Close() error {
	err := d.pr.Close()
	if cerr := d.src.Close(); cerr != nil && err == nil {
		err = cerr
	}
	return err
}

// newDemuxReader spins up demux over src in the background and returns a reader
// of the reassembled output. Closing the returned reader also closes src.
func newDemuxReader(src io.ReadCloser) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(demux(pw, src))
	}()
	return &demuxReadCloser{pr: pr, src: src}
}
