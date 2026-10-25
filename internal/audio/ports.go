package audio

import (
	"io"
	"time"
)

// Decoder turns an audio file into a raw PCM stream.
//
// Reading the stream returns a non-EOF error when decoding fails, and Close
// stops decoding. The ffmpeg package implements it.
type Decoder interface {
	Decode(path string, f Format) (io.ReadCloser, error)
}

// Prober reads track metadata without decoding it.
type Prober interface {
	Duration(path string) (time.Duration, error)
}

// Output is a sound device that plays PCM streams. The otoout package
// implements it.
type Output interface {
	NewVoice(src io.ReadSeeker) Voice
}

// Voice is one PCM stream playing on an Output. It pulls data from its source
// on its own; *oto.Player implements it.
type Voice interface {
	Play()
	Pause()
	IsPlaying() bool
	SetVolume(volume float64)
	// BufferedSize is the amount of data read from the source but not played yet.
	BufferedSize() int
	Seek(offset int64, whence int) (int64, error)
	Err() error
}
