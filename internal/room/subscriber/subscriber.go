package subscriber

type Config struct{}

type WebrtcLivekitSubscriber struct{}

func New(config Config) *WebrtcLivekitSubscriber {
	return &WebrtcLivekitSubscriber{}
}
