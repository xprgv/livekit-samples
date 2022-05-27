package sampler

import (
	"errors"
	"log"
	"net"

	"github.com/livekit/server-sdk-go/pkg/samplebuilder"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3/pkg/media"
)

type RtpH264StreamSampleProvider struct {
	conn          *net.UDPConn
	samplebuilder *samplebuilder.SampleBuilder

	samplesQueue chan *media.Sample
}

func NewRtpH264StreamSampleProvider() (*RtpH264StreamSampleProvider, error) {
	p := &RtpH264StreamSampleProvider{
		samplesQueue: make(chan *media.Sample),
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20360})
	if err != nil {
		return nil, err
	}
	p.conn = listener

	h264SampleBuider := samplebuilder.New(1000, &codecs.H264Packet{}, 90000)
	p.samplebuilder = h264SampleBuider

	go func() {
		buf := make([]byte, 1500)
		for {
			n, _, err := p.conn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}
			rtpPacket := rtp.Packet{}
			if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
				log.Fatal(err)
			}
			switch rtpPacket.PayloadType {
			case 96:
				p.samplebuilder.Push(&rtpPacket)
				for sample := p.samplebuilder.Pop(); sample != nil; sample = p.samplebuilder.Pop() {
					p.samplesQueue <- sample
				}
			}
		}
	}()

	return p, nil
}

func (p *RtpH264StreamSampleProvider) NextSample() (media.Sample, error) {
	sample, ok := <-p.samplesQueue
	if !ok {
		return media.Sample{}, errors.New("failed to create sample")
	}
	// fmt.Println("get sample")

	return *sample, nil
}

func (p *RtpH264StreamSampleProvider) OnBind() error {
	return nil
}

func (p *RtpH264StreamSampleProvider) OnUnbind() error {
	p.conn.Close()
	return nil
}
