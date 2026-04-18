#!/bin/sh

install_bin_path=/usr/local/bin
install_etc_path=/etc/supd

case "$1" in
    "install")
        if [ ! -f "../bin/supd" ]; then
            echo "bin/supd is not exist."
            exit 1
        fi

        mkdir -p /usr/bin
        cp -rf ../bin/supd $install_bin_path
        cp -rf ../bin/supc $install_bin_path
        mkdir -p /etc/supd/conf.d
        cp -rf ../etc/supd/supd.ini $install_etc_path
        echo "Install supd done."
    ;;
    "upgrade")
        if [ ! -f "../bin/supd" ]; then
            echo "bin/supd is not exist."
            exit 1
        fi
        cp -rf ../bin/supd $install_bin_path
        cp -rf ../bin/supc $install_bin_path
        echo "Upgrade bin done, need restart supd by manually."
    ;;
    "clean")
        if [ -f "$install_bin_path/supd" ];then
            rm $install_bin_path/supd
            rm $install_bin_path/supc
            echo "Clean bin done, need remove '$install_etc_path' by manully."
        fi
    ;;
esac
