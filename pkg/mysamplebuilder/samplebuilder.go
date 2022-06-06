package mysamplebuilder

import "github.com/pion/rtp"

type SampleBuilder struct {
	packets []rtp.Packet

	maxLate      uint16
	depacketizer rtp.Depacketizer
}

// func New(maxLate uint16, depacketizer rtp.Depacketizer, sampleRate uint32) *SampleBuilder {
// 	if maxLate < 2 {
// 		maxLate = 2
// 	}
// 	if maxLate > 0x7FFF {
// 		maxLate = 0x7FFF
// 	}
// }
