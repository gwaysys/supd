#/bin/bash

bootName=""
case `uname` in
    "Linux"|"linux")
        if [ -d "/lib/systemd/system" ]; then
            bootName="boot_systemd.sh "
        fi
    ;;
    "Darwin")
        bootName="boot_darwin.sh"
    ;;
    # TODO: more system support
esac

if [ -z "${bootName}" ]; then
    echo "system not support, run supd by command './bin/supd -c ./etc/supd/supd.ini'"
    exit 0
fi

case "$1" in 
    "install")
        cd script
        ./${bootName} install
        cd ..
        ;;
    "upgrade")
        cd script
        ./${bootName} upgrade
        cd ..
        ;;
    "clean")
        cd script
        ./${bootName} clean
        cd ..
        ;;
    *)
        echo "install -- install to system."
        echo "upgrade -- reinstall supd binary and restart supd."
        echo "clean -- remove the installed."
        ;;
esac


