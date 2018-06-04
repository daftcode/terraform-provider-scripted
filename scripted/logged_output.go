package scripted

import (
	"fmt"
	"github.com/armon/circbuf"
	"github.com/mitchellh/go-linereader"
	"io"
	"os"
)

type LoggedOutput struct {
	s      *Scripted
	tag    string
	pw     *os.File
	pr     *os.File
	tee    io.Reader
	buf    *circbuf.Buffer
	doneCh chan struct{}
}

func newLoggedOutput(s *Scripted, tag string) (*LoggedOutput, error) {
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logging pipe: %s", err)
	}
	buf, err := circbuf.NewBuffer(8 * 1024)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logging buffer: %s", err)
	}
	lo := &LoggedOutput{
		s:      s,
		tag:    tag,
		pw:     pw,
		pr:     pr,
		buf:    buf,
		tee:    io.TeeReader(pr, buf),
		doneCh: make(chan struct{}),
	}

	return lo, nil
}

func (lo *LoggedOutput) Start() *os.File {
	go lo.logOutput()
	return lo.pw
}

func (lo *LoggedOutput) Close() {
	lo.pw.Close()
	select {
	case <-lo.doneCh:
	}
}

func (lo *LoggedOutput) logOutput() {
	defer close(lo.doneCh)
	lr := linereader.New(lo.tee)
	for line := range lr.Ch {
		format := fmt.Sprintf("<%[1]s ppid=%-5[3]d pid=%-5[4]d>%%-%[2]ds</%[1]s>", lo.tag, lo.s.pc.Commands.Output.LineWidth, os.Getppid(), os.Getpid())
		lo.s.log(lo.s.pc.Commands.Output.LogLevel, fmt.Sprintf(format, line))
	}
}
