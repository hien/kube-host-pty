package util

import (
	"io/ioutil"
)

func Unmarshal(file string, out interface{}, unmarshalFunc func([]byte, interface{}) error) error {
	config, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = unmarshalFunc(config, out)
	if err != nil {
		return err
	}

	return nil
}

func NotifyWhenConfigChanged(file string, out interface{}, unmarshalFunc func([]byte, interface{}) error) <-chan struct{} {
	Workers.Add(func(sigContinue func()) (interface{}, error) {
		return nil, nil
	})

	fileChangedCh := make(chan struct{}, 1)
	if ch, err := WatchFileWrite(file); err == nil {
		Workers.Add(func(continueFunc func()) (val interface{}, err error) {
			for {
				select {
				case _, more := <-ch.Write:
					if !more {
						close(fileChangedCh)
						return
					}

					if err := Unmarshal(file, out, unmarshalFunc); err != nil {
						continue
					}

					fileChangedCh <- struct{}{}
				}
			}
		})
	}
	return fileChangedCh
}
