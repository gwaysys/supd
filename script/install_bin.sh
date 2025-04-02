#!/bin/sh

case "$1" in
    "install")
        if [ ! -f "../bin/supd" ]; then
            echo "bin/supd is not exist."
            exit 1
        fi

        mkdir -p /usr/bin
        cp -rf ../bin/supd /usr/bin
        cp -rf ../bin/supc /usr/bin
        mkdir -p /etc/supd/conf.d
        cp -rf ../etc/supd/supd.ini /etc/supd
    ;;
    "upgrade")
        if [ ! -f "../bin/supd" ]; then
            echo "bin/supd is not exist."
            exit 1
        fi
        cp -rf ../bin/supd /usr/bin
        cp -rf ../bin/supc /usr/bin
    ;;
    "clean")
        if [ -f "/usr/bin/supd" ];then
            rm /usr/bin/supd
            rm /usr/bin/supc
        fi
    ;;
esac
