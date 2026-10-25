package audio

import (
	"errors"
	"io"
	"sync"
)

// pcmBuffer holds the decoded PCM of one track. The decoder fills it in the
// background while the voice reads from it, so playback starts at once and
// seeking inside the already decoded part is instant.
type pcmBuffer struct {
	frameSize int64

	mu     sync.Mutex
	cond   *sync.Cond
	data   []byte
	pos    int64
	done   bool
	closed bool
	err    error
}

func newPCMBuffer(frameSize int64) *pcmBuffer {
	b := &pcmBuffer{frameSize: frameSize}
	b.cond = sync.NewCond(&b.mu)
	return b
}

func (b *pcmBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0, io.ErrClosedPipe
	}
	b.data = append(b.data, p...)
	b.cond.Broadcast()
	return len(p), nil
}

// Read blocks while the reader is ahead of the decoder.
func (b *pcmBuffer) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for b.pos >= int64(len(b.data)) && !b.done && !b.closed {
		b.cond.Wait()
	}
	if b.closed || b.pos >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n := copy(p, b.data[b.pos:])
	b.pos += int64(n)
	return n, nil
}

func (b *pcmBuffer) Seek(offset int64, whence int) (int64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch whence {
	case io.SeekCurrent:
		offset += b.pos
	case io.SeekEnd:
		offset += int64(len(b.data))
	}
	if offset < 0 {
		return 0, errors.New("pcm: negative position")
	}
	offset -= offset % b.frameSize
	if b.done && offset > int64(len(b.data)) {
		offset = int64(len(b.data))
	}
	b.pos = offset
	return offset, nil
}

// fill copies the decoded stream into the buffer until it ends.
func (b *pcmBuffer) fill(src io.Reader) {
	_, err := io.Copy(b, src)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.done = true
	if !b.closed {
		b.err = err
	}
	b.cond.Broadcast()
}

// Close unblocks pending reads and drops the decoded data.
func (b *pcmBuffer) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.data = nil
	b.cond.Broadcast()
}

type pcmState struct {
	pos  int64
	size int64
	done bool
	err  error
}

func (b *pcmBuffer) state() pcmState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return pcmState{pos: b.pos, size: int64(len(b.data)), done: b.done, err: b.err}
}
