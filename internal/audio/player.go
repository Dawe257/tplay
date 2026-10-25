// Package audio is the playback engine. It plays one track at a time and
// reaches the outside world only through the Decoder, Prober and Output ports.
package audio

import "time"

// Status is a snapshot of the playback state.
type Status struct {
	Position time.Duration
	Duration time.Duration // 0 while unknown
	Volume   float64
	Paused   bool
	// Finished means the track played to the end or failed to play (see Err).
	Finished bool
	Err      error
}

type Player struct {
	decoder Decoder
	prober  Prober
	output  Output
	format  Format

	track   *track
	loadErr error
	volume  float64
	paused  bool
}

type track struct {
	stream   interface{ Close() error }
	pcm      *pcmBuffer
	voice    Voice
	duration time.Duration
}

// NewPlayer wires the engine to its adapters. prober may be nil; the duration
// is then known only once the track is fully decoded.
func NewPlayer(decoder Decoder, prober Prober, output Output, format Format) *Player {
	return &Player{decoder: decoder, prober: prober, output: output, format: format, volume: 1}
}

// Load stops the current track and starts path from the beginning. A failure
// is reported through Status, like any other playback error.
func (p *Player) Load(path string) {
	p.Stop()
	stream, err := p.decoder.Decode(path, p.format)
	if err != nil {
		p.loadErr = err
		return
	}

	pcm := newPCMBuffer(p.format.FrameSize())
	go pcm.fill(stream)
	t := &track{stream: stream, pcm: pcm, voice: p.output.NewVoice(pcm)}
	if p.prober != nil {
		if d, err := p.prober.Duration(path); err == nil {
			t.duration = d
		}
	}
	t.voice.SetVolume(p.volume)
	if !p.paused {
		t.voice.Play()
	}
	p.track = t
}

func (p *Player) Stop() {
	p.loadErr = nil
	t := p.track
	if t == nil {
		return
	}
	p.track = nil
	// Close the buffer first: it releases a read the voice may be blocked in.
	t.pcm.Close()
	t.voice.Pause()
	t.stream.Close()
}

func (p *Player) TogglePause() {
	p.paused = !p.paused
	if p.track == nil {
		return
	}
	if p.paused {
		p.track.voice.Pause()
	} else {
		p.track.voice.Play()
	}
}

func (p *Player) SetVolume(v float64) {
	p.volume = min(max(v, 0), 1)
	if p.track != nil {
		p.track.voice.SetVolume(p.volume)
	}
}

// Seek moves to the given time, clamped to the track bounds.
func (p *Player) Seek(to time.Duration) {
	if p.track == nil {
		return
	}
	if d := p.Status().Duration; d > 0 {
		to = min(to, d)
	}
	p.track.voice.Seek(p.format.Bytes(max(to, 0)), 0)
}

// SeekBy moves relative to the current position.
func (p *Player) SeekBy(delta time.Duration) {
	p.Seek(p.Status().Position + delta)
}

func (p *Player) Status() Status {
	s := Status{Volume: p.volume, Paused: p.paused}
	t := p.track
	if t == nil {
		s.Finished, s.Err = p.loadErr != nil, p.loadErr
		return s
	}

	pcm := t.pcm.state()
	// The voice has read ahead of what the listener hears by its buffer.
	s.Position = p.format.Duration(max(pcm.pos-int64(t.voice.BufferedSize()), 0))
	s.Duration = t.duration
	if s.Duration == 0 && pcm.done {
		s.Duration = p.format.Duration(pcm.size)
	}

	s.Err = pcm.err
	if s.Err == nil {
		s.Err = t.voice.Err()
	}
	drained := pcm.done && pcm.pos >= pcm.size && !t.voice.IsPlaying()
	s.Finished = !p.paused && (drained || t.voice.Err() != nil)
	return s
}
