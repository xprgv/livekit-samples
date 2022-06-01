#!/bin/bash

ffmpeg -re -stream_loop -1 \
    -i ../files/bike.webm \
    -c:v libvpx \
    -keyint_min 120 \
    -qmax 50 \
    -maxrate 2M \
    -b:v 200K \
    -an \
    -sdp_file vp8_rtp.sdp \
    -f rtp "rtp://127.0.0.1:5500"