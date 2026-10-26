// Package otoout implements the audio Output port with oto, which talks to
// the sound server through ALSA (PipeWire and PulseAudio provide it).
package otoout

import (
	"io"
	"time"

	"github.com/ebitengine/oto/v3"

	"github.com/Dawe257/tplay/internal/audio"
)

type Output struct {
	ctx *oto.Context
}

// New opens the sound device. Only one Output can exist per process.
func New(f audio.Format) (*Output, error) {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   f.SampleRate,
		ChannelCount: f.Channels,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   100 * time.Millisecond,
	})
	if err != nil {
		return nil, err
	}
	<-ready
	return &Output{ctx: ctx}, nil
}

func (o *Output) NewVoice(src io.ReadSeeker) audio.Voice {
	return o.ctx.NewPlayer(src)
}
