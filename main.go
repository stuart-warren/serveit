package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

func init() {
}

// Run runs the command and returns nill if succesfull, otherwise an error
func Run(in io.Reader, out io.Writer, args []string) error {

	// flags
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		listen = flags.String("l", ":5000", "HOST:PORT to listen on")
	)
	flags.Parse(args[1:])

	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = &logrus.TextFormatter{}

	// Output to stderr instead of stdout, could also be a file.
	log.Out = out

	// Only log the warning severity or above.
	log.Level = logrus.InfoLevel

	//	s := bufio.NewScanner(in)
	//	for s.Scan() {
	//		out.Write(s.Bytes())
	//		io.WriteString(out, " "+*listen)
	//	}
	//	if err := s.Err(); err != nil {
	//		return err
	//	}
	conf, err := NewConfig(*listen)
	if err != nil {
		return fmt.Errorf("Error with configuration: %s", err)
	}
	server := NewServer(conf)
	server.SetLogger(log)

	// FIXME
	if len(flags.Args()) > 0 {
		cmd := exec.Command("child", flags.Args()...)
		cmd.Stdout = out
		cmd.Stderr = out
		defer log.Fatalln(cmd.Wait())
		cmd.Start()
	}
	return server.Run()
}

func main() {
	if err := Run(os.Stdin, os.Stdout, os.Args); err != nil {
		log.Fatalln(err)
	}
}
