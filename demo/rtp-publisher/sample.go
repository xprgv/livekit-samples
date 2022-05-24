package main

import "github.com/pion/webrtc/v3/pkg/media"

type MySampleProvider struct{}

func (p *MySampleProvider) NextSample() (media.Sample, error) {
	sample := media.Sample{}

	return sample, nil
}
