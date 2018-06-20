package scripted

import (
	"fmt"
	"io"
	"os"
)

type LoggedOutput struct {
	s      *Scripted
	tag    string
	pw     *io.PipeWriter
	pr     *io.PipeReader
	doneCh chan struct{}
}

func newLoggedOutput(s *Scripted, tag string) *LoggedOutput {
	pr, pw := io.Pipe()
	return &LoggedOutput{
		s:      s,
		tag:    tag,
		pw:     pw,
		pr:     pr,
		doneCh: make(chan struct{}),
	}
}

func (lo *LoggedOutput) Start() *io.PipeWriter {
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
	lines := make(chan string)
	go lo.s.scanLines(lines, lo.pr)
	extra := ""
	if lo.s.pc.Commands.Output.LogPids {
		extra += fmt.Sprintf(" ppid=%-5[1]d pid=%-5[2]d", os.Getppid(), os.Getpid())
	}
	for line := range lines {
		format := fmt.Sprintf("<%[1]s%[3]s>%%-%[2]ds</%[1]s>", lo.tag, lo.s.pc.Commands.Output.LineWidth, extra)

		lo.s.log(lo.s.pc.Commands.Output.LogLevel, fmt.Sprintf(format, line))
	}
}
