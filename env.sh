#!/bin/bash

# Template from https://github.com/gwaysys/goapp

# Export environment for building or runing program
# -------------------------------------------------
export PRJ_ROOT=`pwd`
export PRJ_NAME="supd"
export GOBIN=$PRJ_ROOT/bin
export GO111MODULE=on

export BUILD_VER=""
# Env of the var path which the version will write to, example: github.com/gwaysys/supd/version.Version
export BUILD_VER_PATH="" 
# Env of the var path which git describe will write to, example: github.com/gwaysys/supd/version.GitCommit
export BUILD_GIT_PATH=""

export BUILD_LDFLAGS="" # set go ldflags if need
# export APP_ARGS="serve --etc-root=$PRJ_ROOT" # sup start with args when \$APP_ARGS is set
#
# Setting directory of sup [command] all, split with space for multiply directory.
export BUILD_ALL_PATH="." # for build
export INSTALL_ALL_PATH="" # for running, using $BUILD_ALL_PATH when not set

# Setting supd program params configuration
## --------------------START-------------------
export SUP_USER=$USER # default is $USER
export SUP_ETC_DIR="/etc/supd/conf.d" # default is /etc/supd/conf.d
export SUP_LOG_SIZE="10MB"
export SUP_LOG_BAK="10"
export SUP_APP_ENV="PRJ_ROOT=\\\"$PRJ_ROOT\\\",GWAYLIB_LOG_LEVEL=1"

# Seting collection to sup publish
# -------------------------------------------------
## Root directory to colect of project. default is etc, split with space for multiply directory.
export PUB_ROOT_RES="bin" 
## App directory to colect of application. default is public, split with space for multiply directory.
export PUB_APP_RES="" 

# Setting GOROOT
# -------------------------------------------------
if [ -z "$GOROOT" ]; then
    export GOROOT="/usr/local/go"
fi

# Setting GOBIN to PATH
# -------------------------------------------------
bin_path=$GOBIN:$GOROOT/bin:
rep=${PATH/bin_path/""}
if [ ! "$PATH" = "$rep" ]; then
    PATH=$rep 
fi
export PATH=$bin_path$PATH

main(){
    # Download sup to manage project
    # -------------------------------------------------
    if [ ! -f $GOBIN/sup ]; then
        type curl >/dev/null 2>&1||{ echo -e >&2 "curl not found"; return 1; }
        echo "Download sup to bin"
        mkdir -p $GOBIN&& \
        #curl https://raw.githubusercontent.com/gwaysys/supd/master/bin/sup -o $GOBIN/sup && \
        curl https://raw.githubusercontent.com/gwaysys/supd/v1.2.0/bin/sup -o $GOBIN/sup && \
        chmod +x $GOBIN/sup&&echo "Download sup done."|| return 1
    fi
    # --------------------END--------------------
    
    echo "Envs have set to \"$PRJ_NAME\""
    echo "Using \"sup help\" to manage project"
    # -------------------------------------------------
}

main||echo "Check env failed"
