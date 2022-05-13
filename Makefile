
stream_file:
	go run demo/stream_file/main.go

ffmpeg_pipe:
	go run demo/ffmpeg_pipe/main.go

sdp:
	go run demo/sdp/main.go

publish:
	go run cmd/publisher_rtp_stream/main.go