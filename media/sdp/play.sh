#!/bin/bash

ffplay -protocol_whitelist file,rtp,udp -i from_restreamer.sdp & ffplay udp://238.1.1.1:5500