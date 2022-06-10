package publisher

import (
	lksdk "github.com/livekit/server-sdk-go"
)

type Config struct {
}

type WebrtcLivekitPublisher struct {
	room *lksdk.Room
}

func New(config Config) *WebrtcLivekitPublisher {
	return &WebrtcLivekitPublisher{}
}

func (p *WebrtcLivekitPublisher) ConnectToRoom() error {

	return nil
}
