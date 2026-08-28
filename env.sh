#!/bin/bash

# Template from https://github.com/gwaysys/goapp

# The program environment variables that need to be exported for the project
# -------------------------------------------------
export PRJ_ROOT=`pwd`
export PRJ_NAME="supd"
export GOBIN=$PRJ_ROOT/bin
export GO111MODULE=on

# Set the sup [command] all directories to be traversed
export BUILD_ALL_PATH="cmd/supd" # for build
export INSTALL_ALL_PATH="" # for running, using $BUILD_ALL_PATH when not set

# Changing the path can change the version number of the compiler. If not specified, the system's default configuration will be used
go_root="/usr/local/go"
if [ -d "$go_root" ]; then
    export GOROOT="$go_root"
fi

# Add GOBIN to PATH
bin_path=$GOBIN:$GOROOT/bin:

rep=${PATH/bin_path/""}
if [ ! "$PATH" = "$rep" ]; then
    PATH=$rep # Reset the position of the original value
fi
export PATH=$bin_path$PATH

# Download the sup management tool
if [ ! -f $GOBIN/sup ]; then
    type curl >/dev/null 2>&1||{ echo -e >&2 "curl not found, need install at first."; exit 0; }
    echo "Download sup to bin."
    mkdir -p $GOBIN&& \
    curl https://raw.githubusercontent.com/gwaysys/supd/v1/bin/sup -o $GOBIN/sup && \
    chmod +x $GOBIN/sup&&echo "Download sup done."|| exit 0
fi
# --------------------END--------------------

echo "Env have changed to \"$PRJ_NAME\""
echo "Using \"sup help\" to manage project"
# -------------------------------------------------