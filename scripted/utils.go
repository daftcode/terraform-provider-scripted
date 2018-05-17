package scripted

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/go-linereader"
	"io"
	"os/exec"
	"reflect"
	"syscall"
)

func mergeMaps(maps ...map[string]string) map[string]string {
	ctx := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			ctx[k] = v
		}
	}
	return ctx
}

func castConfigList(v interface{}) []string {
	var ret []string
	for _, v := range v.([]interface{}) {
		ret = append(ret, v.(string))
	}
	return ret
}

func castConfigMap(v interface{}) map[string]string {
	ret := map[string]string{}
	if v == nil {
		return ret
	}
	for k, v := range v.(map[string]interface{}) {
		ret[k] = v.(string)
	}
	return ret
}

func castConfigChangeMap(o, n interface{}) *ChangeMap {
	return &ChangeMap{
		Old: castConfigMap(o),
		New: castConfigMap(n),
	}
}

func mapToEnv(env map[string]string) []string {
	var ret []string
	for key, value := range env {
		ret = append(ret, fmt.Sprintf("%s=%s", key, value))
	}
	return ret
}

func is(b, other interface{}) bool {
	x := reflect.ValueOf(b)
	y := reflect.ValueOf(other)
	return x.Pointer() == y.Pointer()
}

func getExitStatus(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return -1
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func copyOutput(s *State, r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)
	lr := linereader.New(r)
	for line := range lr.Ch {
		format := fmt.Sprintf("<LINE>%%-%ds</LINE>", s.pc.CommandLogWidth)
		s.log(s.pc.CommandLogLevel, fmt.Sprintf(format, line))
	}
}

func selectLogFunction(logger hclog.Logger, level hclog.Level) func(msg string, args ...interface{}) {
	switch level {
	case hclog.Trace:
		return logger.Trace
	case hclog.Debug:
		return logger.Debug
	case hclog.Info:
		return logger.Info
	case hclog.Warn:
		return logger.Warn
	case hclog.Error:
		return logger.Error
	default:
		return logger.Info
	}
}
