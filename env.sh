#!/bin/bash
# iperf-rpc deployment configuration
# Source this file before deploying: source env.sh

export IR_IMAGE=ghcr.io/wold9168/iperf-rpc
export IR_IMAGE_TAG=latest

export IR_NODEPORT_API=30080
export IR_NODEPORT_IPERF=30081
