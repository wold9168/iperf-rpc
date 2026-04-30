#!/bin/bash
# iperf-rpc deployment configuration
# Source this file before deploying: source env.sh

export IR_IMAGE=ghcr.io/wold9168/iperf-rpc
export IR_IMAGE_TAG=latest

export IR_NODEPORT_SERVER_API=30080
export IR_NODEPORT_SERVER_IPERF=30081
export IR_NODEPORT_CLIENT_API=30082
