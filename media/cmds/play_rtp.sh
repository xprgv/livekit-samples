#!/bin/bash

ffplay -protocol_whitelist file,rtp,udp -i video.sdp & ffplay udp://127.0.0.1:5500