#!/bin/bash

# ffmpeg -re -stream_loop -1 \
#     -i ../files/big_buck_bunny.mp4 \
#     -c:v copy \
#     -an \
#     -sdp_file h264_rtp.sdp \
#     -f rtp "rtp://127.0.0.1:5500"


ffmpeg -re -stream_loop -1 \
    -an \
    -f lavfi -i testsrc=size=640x480:rate=30 \
    -pix_fmt yuv420p \
    -c:v libx264 \
    -g 10 \
    -tune zerolatency \
    -f rtp rtp://127.0.0.1:5500?pkt_size=1200