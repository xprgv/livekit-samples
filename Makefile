
stream_file:
	go run examples/stream_file/main.go

publish:
	go run cmd/publisher_rtp_stream/main.go

ffmpeg_pipe:
	go run examples/ffmpeg_pipe/main.go
