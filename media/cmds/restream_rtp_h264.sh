#!/bin/bash

ffmpeg -re \
    -i srt://176.113.118.51:25360 \
    -map 0:1 \
    -c copy \
    -f rtp rtp://127.0.0.1:5500?pkt_size=1200