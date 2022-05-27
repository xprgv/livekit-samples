package sampler

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/livekit/server-sdk-go/pkg/samplebuilder"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3/pkg/media"
)

type RtpOpusSampleProvider struct {
	conn          *net.UDPConn
	samplebuilder *samplebuilder.SampleBuilder

	samplesQueue chan *media.Sample
}

func NewRtpOpusSampleProvider() (*RtpOpusSampleProvider, error) {
	p := &RtpOpusSampleProvider{
		samplesQueue: make(chan *media.Sample),
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 21360})
	if err != nil {
		return nil, err
	}
	p.conn = listener

	opusSampleBuilder := samplebuilder.New(0, &codecs.OpusPacket{}, 48000)
	p.samplebuilder = opusSampleBuilder

	go func() {
		buf := make([]byte, 1500)
		for {
			n, _, err := p.conn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
				// log.Println(err)
				// return
			}
			rtpPacket := rtp.Packet{}
			if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
				log.Fatal(err)
			}
			switch rtpPacket.PayloadType {
			case 97: // opus media
				p.samplebuilder.Push(&rtpPacket)
				for sample := p.samplebuilder.Pop(); sample != nil; sample = p.samplebuilder.Pop() {
					p.samplesQueue <- sample
				}
			}
		}
	}()

	return p, nil
}

func (p *RtpOpusSampleProvider) NextSample() (media.Sample, error) {
	sample, ok := <-p.samplesQueue
	if !ok {
		err := errors.New("failed to create sample")
		fmt.Println(err)
		return media.Sample{}, err
	}
	// fmt.Println(sample.Duration)
	return *sample, nil
}

func (p *RtpOpusSampleProvider) OnBind() error {
	fmt.Println("on bind")
	return nil
}

func (p *RtpOpusSampleProvider) OnUnbind() error {
	fmt.Println("on unbind")
	p.conn.Close()
	return nil
}
