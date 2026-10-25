package audio

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"
)

// testFormat makes one second exactly 100 bytes: 50 frames of one channel.
var testFormat = Format{SampleRate: 50, Channels: 1}

type fakeStream struct {
	io.Reader
	closed bool
}

func (s *fakeStream) Close() error { s.closed = true; return nil }

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

type fakeDecoder struct {
	data    []byte
	readErr error // returned by the stream after data
	openErr error
	stream  *fakeStream
}

func (d *fakeDecoder) Decode(string, Format) (io.ReadCloser, error) {
	if d.openErr != nil {
		return nil, d.openErr
	}
	var r io.Reader = bytes.NewReader(d.data)
	if d.readErr != nil {
		r = io.MultiReader(r, failingReader{d.readErr})
	}
	d.stream = &fakeStream{Reader: r}
	return d.stream, nil
}

type fakeProber time.Duration

func (p fakeProber) Duration(string) (time.Duration, error) { return time.Duration(p), nil }

// fakeVoice plays nothing by itself: tests pull data with consume.
type fakeVoice struct {
	src     io.ReadSeeker
	playing bool
	volume  float64
}

func (v *fakeVoice) Play()                 { v.playing = true }
func (v *fakeVoice) Pause()                { v.playing = false }
func (v *fakeVoice) IsPlaying() bool       { return v.playing }
func (v *fakeVoice) SetVolume(vol float64) { v.volume = vol }
func (v *fakeVoice) BufferedSize() int     { return 0 }
func (v *fakeVoice) Err() error            { return nil }
func (v *fakeVoice) Seek(off int64, whence int) (int64, error) {
	return v.src.Seek(off, whence)
}

// consume plays n bytes; at the end of the source it stops like a real voice.
func (v *fakeVoice) consume(t *testing.T, n int) {
	t.Helper()
	buf := make([]byte, n)
	if _, err := io.ReadFull(v.src, buf); err != nil {
		v.playing = false
	}
}

type fakeOutput struct{ voice *fakeVoice }

func (o *fakeOutput) NewVoice(src io.ReadSeeker) Voice {
	o.voice = &fakeVoice{src: src}
	return o.voice
}

func newTestPlayer(dec *fakeDecoder, prober Prober) (*Player, *fakeOutput) {
	out := &fakeOutput{}
	return NewPlayer(dec, prober, out, testFormat), out
}

// waitDecoded waits for the background decoding of the fake stream.
func waitDecoded(t *testing.T, p *Player) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !p.track.pcm.state().done {
		if time.Now().After(deadline) {
			t.Fatal("decoding did not finish")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestPlaysToTheEnd(t *testing.T) {
	p, out := newTestPlayer(&fakeDecoder{data: make([]byte, 300)}, nil)
	p.Load("song")
	waitDecoded(t, p)

	out.voice.consume(t, 100)
	s := p.Status()
	if s.Position != time.Second || s.Duration != 3*time.Second || s.Finished {
		t.Fatalf("after 1s: %+v", s)
	}
	out.voice.consume(t, 300)
	if s := p.Status(); !s.Finished || s.Err != nil {
		t.Fatalf("after the end: %+v", s)
	}
}

func TestDurationFromProber(t *testing.T) {
	p, _ := newTestPlayer(&fakeDecoder{data: make([]byte, 100)}, fakeProber(7*time.Second))
	p.Load("song")
	if d := p.Status().Duration; d != 7*time.Second {
		t.Fatalf("duration %v, want 7s", d)
	}
}

func TestSeekIsClamped(t *testing.T) {
	p, _ := newTestPlayer(&fakeDecoder{data: make([]byte, 500)}, nil)
	p.Load("song")
	waitDecoded(t, p)

	p.Seek(2 * time.Second)
	p.SeekBy(1500 * time.Millisecond)
	if pos := p.Status().Position; pos != 3500*time.Millisecond {
		t.Fatalf("position %v, want 3.5s", pos)
	}
	p.Seek(time.Minute)
	if pos := p.Status().Position; pos != 5*time.Second {
		t.Fatalf("position %v, want the end", pos)
	}
	p.SeekBy(-time.Hour)
	if pos := p.Status().Position; pos != 0 {
		t.Fatalf("position %v, want 0", pos)
	}
}

func TestPauseKeepsTrackUnfinished(t *testing.T) {
	p, out := newTestPlayer(&fakeDecoder{data: make([]byte, 100)}, nil)
	p.Load("song")
	waitDecoded(t, p)
	p.TogglePause()
	if out.voice.playing || !p.Status().Paused {
		t.Fatal("pause did not stop the voice")
	}
	if p.Status().Finished {
		t.Fatal("a paused track reported finished")
	}
	p.TogglePause()
	out.voice.consume(t, 200)
	if !p.Status().Finished {
		t.Fatal("track not finished after resuming at the end")
	}
}

func TestPauseAndVolumeCarryOverToNextTrack(t *testing.T) {
	p, out := newTestPlayer(&fakeDecoder{data: make([]byte, 100)}, nil)
	p.SetVolume(1.7)
	p.TogglePause()
	p.Load("song")
	if out.voice.playing || out.voice.volume != 1 {
		t.Fatalf("voice playing=%v volume=%v, want paused at 1", out.voice.playing, out.voice.volume)
	}
}

func TestDecodeErrorFinishesTrack(t *testing.T) {
	boom := errors.New("invalid data")
	p, out := newTestPlayer(&fakeDecoder{readErr: boom}, nil)
	p.Load("broken")
	waitDecoded(t, p)
	out.voice.consume(t, 1)
	if s := p.Status(); !s.Finished || !errors.Is(s.Err, boom) {
		t.Fatalf("status %+v, want finished with %v", s, boom)
	}
}

func TestOpenErrorFinishesTrack(t *testing.T) {
	boom := errors.New("ffmpeg not found")
	p, _ := newTestPlayer(&fakeDecoder{openErr: boom}, nil)
	p.Load("song")
	if s := p.Status(); !s.Finished || s.Err != boom {
		t.Fatalf("status %+v, want finished with %v", s, boom)
	}
}

func TestStopClosesStream(t *testing.T) {
	dec := &fakeDecoder{data: make([]byte, 100)}
	p, _ := newTestPlayer(dec, nil)
	p.Load("song")
	p.Stop()
	if !dec.stream.closed {
		t.Fatal("stream left open")
	}
	if s := p.Status(); s.Finished || s.Err != nil {
		t.Fatalf("status after Stop: %+v", s)
	}
}

func TestPCMBufferBlocksUntilDecoded(t *testing.T) {
	b := newPCMBuffer(2)
	got := make(chan []byte)
	go func() {
		buf := make([]byte, 4)
		n, _ := b.Read(buf)
		got <- buf[:n]
	}()
	select {
	case <-got:
		t.Fatal("Read returned before any data was written")
	case <-time.After(20 * time.Millisecond):
	}
	b.Write([]byte{1, 2, 3, 4})
	if data := <-got; !bytes.Equal(data, []byte{1, 2, 3, 4}) {
		t.Fatalf("read %v", data)
	}
	b.Close()
	if _, err := b.Read(make([]byte, 1)); err != io.EOF {
		t.Fatalf("read after Close: %v, want EOF", err)
	}
}

func TestPCMBufferSeekAlignsToFrame(t *testing.T) {
	b := newPCMBuffer(4)
	b.fill(bytes.NewReader(make([]byte, 16)))
	if pos, _ := b.Seek(7, io.SeekStart); pos != 4 {
		t.Fatalf("seek to 7: %d, want 4", pos)
	}
	if pos, _ := b.Seek(100, io.SeekStart); pos != 16 {
		t.Fatalf("seek past the end: %d, want 16", pos)
	}
	if _, err := b.Seek(-1, io.SeekStart); err == nil {
		t.Fatal("negative seek accepted")
	}
}
