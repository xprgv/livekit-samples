#!/bin/bash

# ffmpeg -re -stream_loop -1 \
#     -i ../files/bike.webm \
#     -c:v copy \
#     -an \
#     -sdp_file vp8_rtp.sdp \
#     -f rtp "rtp://127.0.0.1:5500"


# ffmpeg -re \
#     -f lavfi \
#     -i testsrc=size=1920x1080:rate=30 \
#     -vcodec libvpx \
#     -cpu-used 5 \
#     -deadline 1 \
#     -g 10 \
#     -error-resilient 1 \
#     -auto-alt-ref 1 \
#     -f rtp rtp://127.0.0.1:5500?pkt_size=1200

ffmpeg -re -stream_loop -1 \
    -an \
    -i ../files/big_buck_bunny.mp4 \
    -vcodec libvpx \
    -cpu-used 5 \
    -deadline 1 \
    -g 10 \
    -error-resilient 1 \
    -auto-alt-ref 1 \
    -f rtp rtp://127.0.0.1:5500?pkt_size=1200