package celeritas

import (
	"regexp"
	"runtime"
	"time"
)

func (c *Celeritas) LoadTime(start time.Time) {
	elapsed := time.Since(start)
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	regexpFn := regexp.MustCompile(`^.*\.(.*)$`)
	caller := regexpFn.ReplaceAllString(fn.Name(), "$1")
	c.InfoLog.Printf("Load Time: %s took %s\n", caller, elapsed)
}
