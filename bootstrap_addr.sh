#!/bin/bash
getent hosts kademliaBootstrap | awk '{ print $1 }' > bootstrap_host
# wget $(cat bootstrap_host)
