#!/bin/bash

ffmpeg -re -stream_loop -1 \
    -i ../files/big_buck_bunny.mp4 \
    -c:v libx264 \
    -bsf:v h264_mp4toannexb \
    -keyint_min 120 \
    -qmax 50 \
    -maxrate 2M \
    -b:v 300K \
    -profile baseline \
    -an \
    -sdp_file h264_rtp.sdp \
    -f rtp "rtp://127.0.0.1:5500"