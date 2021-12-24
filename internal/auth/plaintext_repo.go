package auth

import (
	"code.cloudfoundry.org/lager"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/timotto/ardumower-relay/internal/model"
	"os"
	"strings"
	"sync"
	"time"
)

type plaintextRepo struct {
	logger   lager.Logger
	filename string

	data  map[string]string
	lock  *sync.RWMutex
	watch *fsnotify.Watcher
}

func newPlaintextRepo(logger lager.Logger, filename string) *plaintextRepo {
	return &plaintextRepo{
		logger:   logger.Session("plaintext-repo"),
		filename: filename,

		data: make(map[string]string),
		lock: &sync.RWMutex{},
	}
}

func (r *plaintextRepo) Lookup(username, password string) (model.User, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	pw, ok := r.data[username]
	if !ok {
		return nil, ErrUnauthorized
	}

	if pw != password {
		return nil, ErrUnauthorized
	}

	return NewUser(username), nil
}

func (r *plaintextRepo) Start() error {
	if err := r.readFile(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if err := r.watcher(); err != nil {
		return fmt.Errorf("failed to watch config: %w", err)
	}

	return nil
}

func (r *plaintextRepo) Stop() {
	if r.watch != nil {
		_ = r.watch.Close()
	}
}

func (r *plaintextRepo) readFile() error {
	file, err := os.ReadFile(r.filename)
	if err != nil {
		return err
	}

	data := make(map[string]string)

	lines := strings.Split(string(file), "\n")
	for i, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			return fmt.Errorf("invalid entry in line %v", i+1)
		}

		data[parts[0]] = strings.Join(parts[1:], ":")
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	r.data = data

	r.logger.Info("read-file", lager.Data{"items": len(data)})

	return nil
}

func (r *plaintextRepo) watcher() error {
	var err error
	r.watch, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := r.watch.Add(r.filename); err != nil {
		_ = r.watch.Close()

		return err
	}

	go r.watchWorker()

	return nil
}

func (r *plaintextRepo) watchWorker() {
	l := r.logger.Session("watcher")
	l.Debug("starting")

	defer func() {
		_ = r.watch.Close()
		l.Debug("stopped")
	}()

	dirty := false
	debounce := make(chan bool)

	for {
		select {
		case <-debounce:
			if !dirty {
				continue
			}
			dirty = false

			if err := r.readFile(); err != nil {
				l.Error("read-config", err)
			}

		case event, ok := <-r.watch.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write != fsnotify.Write {
				continue
			}

			dirty = true
			time.AfterFunc(100*time.Millisecond, func() { debounce <- true })

		case err, ok := <-r.watch.Errors:
			if !ok {
				return
			}
			l.Error("watch", err)
		}
	}
}
