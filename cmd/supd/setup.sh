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
    echo "system not support, run supd by command './supd -c ./etc/supd/supd.ini'"
    exit 0
fi

case "$1" in 
    "install")
        if [ ! -f "./supd" ]; then
            echo "Program supd not found, need build first."
            exit 0
        fi
        cd script
        ./${bootName} install
        cd ..
        echo "Install supd done"
        ;;
    "upgrade")
        if [ ! -f "./supd" ]; then
            echo "Program supd not found, need build first."
            exit 0
        fi
        cd script
        ./${bootName} upgrade
        cd ..
        echo "Upgrade supd done"
        ;;
    "clean")
        cd script
        ./${bootName} clean
        cd ..
        echo "Clean supd done"
        ;;
    *)
        echo "install -- install to system."
        echo "upgrade -- reinstall supd binary and restart supd."
        echo "clean -- remove the installed."
        ;;
esac


