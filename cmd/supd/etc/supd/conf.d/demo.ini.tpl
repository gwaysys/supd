# rename this tpl to demo.ini for running
[program:demo]
#user=root # run user
#environment=GIN_MODE="release"
#directory= # current run directory
command=ping 163.com # run program
autostart=true
startsecs=3
startretries=3
autorestart=true
exitcodes=0,2
stopsignal=TERM
stopwaitsecs=10
#stopasgroup=true
#killasgroup=true
stdout_logfile=$PRJ_ROOT/var/log/demo.logfile.stdout # if PRJ_ROOT not set, using the root directory
stdout_logfile_maxbytes=1MB
stdout_logfile_backups=10
stderr_logfile=$PRJ_ROOT/var/log/demo.logfile.stderr
stderr_logfile_maxbytes=1MB
stderr_logfile_backups=10
