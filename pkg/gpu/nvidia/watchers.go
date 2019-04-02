package nvidia

import (
	log "github.com/astaxie/beego/logs"

	"github.com/fsnotify/fsnotify"
	//"k8s.io/kubernetes/pkg/kubelet/kubeletconfig/util/log"
	"os"
	"os/signal"
)

func newFSWatcher(files ...string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("new fs watcher failed: %s", err)
		return nil, err
	}

	for _, f := range files {
		err = watcher.Add(f)
		if err != nil {
			log.Error("add %s to watcher failed: %s", f, err)
			err2 := watcher.Close()
			if err2 != nil {
				log.Warning("close watcher error: %v", err2)
			}
			return nil, err
		}
	}

	return watcher, nil
}

func newOSWatcher(sigs ...os.Signal) chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)

	return sigChan
}
