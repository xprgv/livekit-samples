package main

import "github.com/pion/webrtc/v3"

var (
	codecH264 = webrtc.RTPCodecCapability{
		MimeType: webrtc.MimeTypeH264,
	}

	codecOpus = webrtc.RTPCodecCapability{
		MimeType: webrtc.MimeTypeOpus,
	}
)
