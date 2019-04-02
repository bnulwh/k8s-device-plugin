package nvidia

import (
	"io/ioutil"
	"runtime"

	log "github.com/astaxie/beego/logs"
)

func StackTrace(all bool) string {
	buf := make([]byte, 10240)

	for {
		size := runtime.Stack(buf, all)

		if size == len(buf) {
			buf = make([]byte, len(buf)<<1)
			continue
		}
		break

	}

	return string(buf)
}

func coreDump(fileName string) {
	log.Info("Dump stacktrace to ", fileName)
	err := ioutil.WriteFile(fileName, []byte(StackTrace(true)), 0644)
	if err != nil {
		log.Error("Write file %s error: %s", fileName, err)
	}
}
