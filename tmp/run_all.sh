#!/bin/bash
    echo "Starting Common Service..."
    ./tmp/common &

    echo "Starting Server ..."
    ./tmp/server &

    wait
