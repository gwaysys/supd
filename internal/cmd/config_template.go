package cmd

import (
	"fmt"
)

var config_template = `[unix_http_server]
file=/tmp/supervisord.sock
#chmod=not support
#chown=not support
username=test1
password={SHA}82ab876d1387bfafe46cc1c8a2ef074eae50cb1d

[inet_http_server]
port=127.0.0.1:9002
username=test1
password=thepassword

[supervisord]
logfile=%(here)s/supervisord.log
logfile_maxbytes=50MB
logfile_backups=10
loglevel=info
pidfile=%(here)s/supervisord.pid
#umask=not support
#nodaemon=not support
#minfds=not support
#minprocs=not support
#nocleanup=not support
#childlogdir=not support
#user=not support
#directory=not support
#strip_ansi=not support
#environment=not support
identifier=supervisor

[program:x]
command=/bin/cat
process_name=%(program_name)s
numprocs=1
#numprocs_start=not support
autostart=true
startsecs=3
startretries=3
autorestart=true
exitcodes=0,2
stopsignal=TERM
stopwaitsecs=10
stopasgroup=true
killasgroup=true
user=user1
redirect_stderr=false
stdout_logfile=AUTO
stdout_logfile_maxbytes=50MB
stdout_logfile_backups=10
stdout_capture_maxbytes=0
stdout_events_enabled=true
stderr_logfile=AUTO
stderr_logfile_maxbytes=50MB
stderr_logfile_backups=10
stderr_capture_maxbytes=0
stderr_events_enabled=false
environment=KEY="val",KEY2="val2"
directory=/tmp
#umask=not support
serverurl=AUTO

[include]
files=/an/absolute/filename.conf /an/absolute/*.conf foo.conf config??.conf

[group:x]
programs=bar,baz
priority=999

[eventlistener:x]
command=/bin/eventlistener
process_name=%(program_name)s
numprocs=1
#numprocs_start=not support
autostart=true
startsecs=3
startretries=3
autorestart=true
exitcodes=0,2
stopsignal=TERM
stopwaitsecs=10
#stopasgroup=not support
#killasgroup=not support
user=user1
redirect_stderr=false
stdout_logfile=AUTO
stdout_logfile_maxbytes=50MB
stdout_logfile_backups=10
stdout_capture_maxbytes=0
stdout_events_enabled=true
stderr_logfile=AUTO
stderr_logfile_maxbytes=50MB
stderr_logfile_backups=10
stderr_capture_maxbytes=0
stderr_events_enabled=false
environment=KEY="val",KEY2="val2"
directory=/tmp
#umask=not support
serverurl=AUTO
buffer_size=10240
events=PROCESS_STATE
#result_handler=not support

[supervisorctl]
serverurl = unix:///tmp/supervisor.sock
username = chris
password = 123
#prompt = not support
`

// TemplateCommand prints the supported supd configuration template to
// stdout. Previously known as InitTemplateCommand, the rename reflects
// the fact that it no longer writes to a file -- it just dumps the
// template so users can pipe it to `tee` / redirect / grep as desired.
type TemplateCommand struct{}

var templateCommand TemplateCommand

// Execute prints the configuration template to stdout. With no flag or
// positional arguments the go-flags parser invokes Execute directly
// and the entire template is emitted on a single stdout write.
func (x *TemplateCommand) Execute(args []string) error {
	_, err := fmt.Print(config_template)
	return err
}

func init() {
	parser.AddCommand("template",
		"print supported configuration template to stdout",
		"The template subcommand prints the full supd configuration "+
			"template (every supported section) to stdout. Pipe or redirect "+
			"the output to create a starter supd.ini.",
		&templateCommand)
}