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

# Supervisord configuration file parameters
## --------------------START-------------------
## The following is the default configuration data for the supervisor during deployment. If not configured, the following default data will be used
## Configure the user for running the supervisor, default is the current user
#export SUP_USER=$USER
## Configure the directory of the supervisor's configuration files
#export SUP_ETC_DIR="/etc/supervisor/conf.d/"
## The maximum size of a single file for configuring the subroutine log of the supervisor
#export SUP_LOG_SIZE="10MB"
## The maximum number of files for configuring the subroutine log of the supervisor
#export SUP_LOG_BAK="10"
## Configure the environment variable in the supervisor configuration
#export SUP_APP_ENV="PRJ_ROOT=\\\"$PRJ_ROOT\\\",GIN_MODE=\\\"release\\\",LD_LIBRARY_PATH=\\\"$LD_LIBRARY_PATH\\\""

# Set the folder environment variables that need to be included when packaging the publish command
# -------------------------------------------------
# export BUILD_GIT_COMMIT="github.com/gwaysys/supd/version.GitCommit" # it will be set when go build ldflags if set
# export BUILD_LDFLAGS="" # Extend go build flag in sup build
export PUB_ROOT_RES="etc script setup.sh" # The folders that need to be packaged in the root directory, such as "etc", using default is not set
# export PUB_APP_RES="public" # The folders under the app, such as "public", using default is not set
#

# Set common args for program ran
# export APP_ARGS="serve --etc-root=$PRJ_ROOT" # sup start with args when \$APP_ARGS is set

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

