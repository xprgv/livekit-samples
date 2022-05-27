package sampleprovider

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/jech/samplebuilder"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
)

type RtpSampleProvider struct {
	mimeType      string
	stream        io.ReadCloser
	samplebuilder *samplebuilder.SampleBuilder
	samplesQueue  chan *media.Sample
}

func NewRtpSampleProvider(mimeType string, stream io.ReadCloser, samplebuilder *samplebuilder.SampleBuilder) *RtpSampleProvider {
	p := &RtpSampleProvider{
		mimeType:      mimeType,
		samplesQueue:  make(chan *media.Sample),
		stream:        stream,
		samplebuilder: samplebuilder,
	}

	go func() {
		buf := make([]byte, 1500)
		for {
			fmt.Println("before read")
			n, err := p.stream.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			rtpPacket := rtp.Packet{}
			if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
				log.Fatal(err)
			}
			fmt.Println(rtpPacket.PayloadType)
			switch rtpPacket.PayloadType {
			case 97: // opus media
				p.samplebuilder.Push(&rtpPacket)
				for sample := p.samplebuilder.Pop(); sample != nil; sample = p.samplebuilder.Pop() {
					p.samplesQueue <- sample
				}
			}
		}
	}()

	return p
}

func (p *RtpSampleProvider) NextSample() (media.Sample, error) {
	sample, ok := <-p.samplesQueue
	if !ok {
		err := errors.New("failed to create sample")
		fmt.Println(err)
		return media.Sample{}, err
	}

	return *sample, nil
}

func (p *RtpSampleProvider) OnBind() error {
	fmt.Println("on bind")
	return nil
}

func (p *RtpSampleProvider) OnUnbind() error {
	fmt.Println("on unbind")
	p.stream.Close()
	return nil
}
