package util

import (
	"github.com/fsnotify/fsnotify"
)

var (
	fsWatcher   *fsnotify.Watcher
	watchedFile = &StringInterfaceMap{}
)

type FileEventChan struct {
	Write chan struct{}
}

func init() {
	var err error
	fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
	} else {
		go func() {
			defer func() { _ = fsWatcher.Close() }()
			for {
				select {
				case err := <-fsWatcher.Errors:
					// TODO: log this error and continue
					_ = err
				case event, more := <-fsWatcher.Events:
					if !more {
						return
					}
					if v, ok := watchedFile.Get(event.Name); ok && v != nil {
						fevtCh := v.(*FileEventChan)
						// write event
						if event.Op&fsnotify.Write == fsnotify.Write {
							if fevtCh.Write != nil {
								fevtCh.Write <- struct{}{}
							}
						}
					}
				}
			}
		}()
	}
}

func WatchFileWrite(file string) (*FileEventChan, error) {
	if err := fsWatcher.Add(file); err != nil {
		return nil, err
	}

	var fevtCh *FileEventChan
	if v, ok := watchedFile.Get(file); ok && v != nil {
		fevtCh = v.(*FileEventChan)
		fevtCh.Write = make(chan struct{})
	} else {
		fevtCh = &FileEventChan{Write: make(chan struct{})}
	}

	watchedFile.Set(file, fevtCh)
	return fevtCh, nil
}

func UnWatchFileWrite(file string) error {
	if err := fsWatcher.Remove(file); err != nil {
		return err
	}

	watchedFile.Del(file)
	return nil
}
