package publisher

type Config struct{}

type WebrtcLivekitPublisher struct{}

func New(config Config) *WebrtcLivekitPublisher {
	return &WebrtcLivekitPublisher{}
}
