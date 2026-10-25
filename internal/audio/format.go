package audio

import "time"

const bytesPerSample = 2

// Format describes the raw PCM passed from the decoder to the output:
// interleaved signed 16-bit little-endian samples.
type Format struct {
	SampleRate int
	Channels   int
}

var DefaultFormat = Format{SampleRate: 44100, Channels: 2}

// FrameSize is the size of one sample for every channel.
func (f Format) FrameSize() int64 { return int64(f.Channels * bytesPerSample) }

func (f Format) BytesPerSecond() int64 { return int64(f.SampleRate) * f.FrameSize() }

// Duration converts a PCM byte count to playback time.
func (f Format) Duration(n int64) time.Duration {
	return time.Duration(float64(n) / float64(f.BytesPerSecond()) * float64(time.Second))
}

// Bytes converts playback time to a PCM byte offset aligned to a frame.
func (f Format) Bytes(d time.Duration) int64 {
	n := int64(d.Seconds() * float64(f.BytesPerSecond()))
	return n - n%f.FrameSize()
}
