// Package svc owns the supervisor daemon implementation: the Supervisor
// type that loads configuration, manages processes and exposes the RPC
// service; the HTTP/RPC server that publishes it; the REST program API;
// and the daemon-mode wrapper that double-forks the process.
//
// All of these used to live in the supd root package; they are extracted
// here so the ctl/ package can import svc.Supervisor directly and so the
// root package can shrink to a thin entry point.
package svc

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"unicode"

	"github.com/gwaysys/supd/internal/logger"

	log "github.com/sirupsen/logrus"
)

// Options mirrors the root-package flag struct used by the daemon entry.
// Kept here so the daemon startup logic can be moved wholesale out of
// the root package.
type Options struct {
	Configuration string `short:"c" long:"configuration" description:"the configuration file"`
	Daemon        bool   `short:"d" long:"daemon" description:"run as daemon"`
	EnvFile       string `long:"env-file" description:"the environment file"`
}

// LoadEnvFile reads key=value pairs from path (one per line) and applies
// them to the current process environment. Lines starting with `#` are
// comments. Lines may optionally start with `export ` (matches the
// historical shell-source semantics). Equivalent to the function that
// used to live in main.go.
func LoadEnvFile(path string) {
	if len(path) <= 0 {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		log.WithFields(log.Fields{"file": path}).Error("Fail to open environment file")
		return
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		if strings.HasPrefix(line, "export") && len(line) > len("export") && unicode.IsSpace(rune(line[len("export")])) {
			line = strings.TrimSpace(line[len("export"):])
		}
		pos := strings.Index(line, "=")
		if pos != -1 {
			k := strings.TrimSpace(line[0:pos])
			v := strings.TrimSpace(line[pos+1:])
			if len(k) > 0 && len(v) > 0 {
				os.Setenv(k, v)
			}
		}
	}
}

// FindSupervisordConf searches well-known locations for the supervisor
// configuration file. Mirrors the legacy findSupervisordConf() in main.go
// 1. $CWD/supervisord.conf
// 2. $CWD/etc/supervisord.conf
// 3. /etc/supervisord.conf
// 4. /etc/supervisor/supervisord.conf (since Supervisor 3.3.0)
// 5. ../etc/supervisord.conf (relative to the executable)
// 6. ../supervisord.conf (relative to the executable)
//
// The hard-coded paths above are adjusted to the supd defaults:
//   $CWD/supd.ini, $CWD/etc/supd.ini, /etc/supd/supd.ini, ../etc/supd.ini,
//   ../supd.ini.
func FindSupervisordConf() (string, error) {
	candidates := []string{
		"./supd.ini",
		"./etc/supd.ini",
		"/etc/supd/supd.ini",
		"../etc/supd.ini",
		"../supd.ini",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return filepath.Abs(p)
		}
	}
	return "", fmt.Errorf("fail to find supd.ini (try -c <path>)")
}

// initSignals wires SIGINT / SIGTERM to the supervisor's graceful shutdown.
// Used by Run so the root package does not need to keep its own copy.
func initSignals(s *Supervisor) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.WithFields(log.Fields{"signal": sig}).Info("receive a signal to stop all process & exit")
		s.procMgr.StopAllProcesses()
		os.Exit(-1)
	}()
}

// Server is the public entry point used by the root supd package to start
// the daemon. It wraps the legacy free-function RunServer() from main.go
// behind a value so callers can configure supervisor / signal handling
// without exposing the internal Supervisor type at the root package level.
type Server struct {
	// EnvFile is the path to an optional env-file (equivalent to the
	// historical options.EnvFile root-package field). May be empty.
	EnvFile string

	// ConfigPath, when non-empty, is used as the configuration file
	// regardless of FindSupervisordConf() auto-detection. Equivalent to
	// options.Configuration in the legacy code.
	ConfigPath string
}

// NewServer returns a Server ready to have its Run method called.
func NewServer() *Server {
	return &Server{}
}

// Run executes the legacy RunServer flow: an infinite loop that reloads
// the supervisor configuration on each iteration, restarting processes
// when the binary is replaced. It blocks until the process is killed.
//
// Callers that need custom signal / zombie handling should construct a
// Server, set EnvFile / ConfigPath, then call Run.
func (s *Server) Run() error {
	LoadEnvFile(s.EnvFile)

	for {
		cfg := s.ConfigPath
		if cfg == "" {
			detected, err := FindSupervisordConf()
			if err != nil {
				return err
			}
			cfg = detected
		}
		sup := NewSupervisor(cfg)
		initSignals(sup)
		if sErr, _, _, _ := sup.reload(); sErr != nil {
			log.Fatal(sErr)
		}
		sup.waitForExit()
	}
}

// Daemonize is a thin wrapper that double-forks the current process and
// runs proc in the child. Equivalent to the historical Deamonize() in
// the root package's daemonize.go. The unusual spelling is preserved so
// the public symbol stays backwards compatible with the original shell
// wrappers.
func Daemonize(proc func()) {
	daemonize(proc)
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logger.CustomFormatter{FullTimestamp: true})
	log.SetLevel(log.DebugLevel)
}