/*
/*
 * Copyright (c) 2024 Augur AI, Inc.
 * This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. 
 * If a copy of the MPL was not distributed with this file, you can obtain one at https://mozilla.org/MPL/2.0/.
 *
 
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package watch

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Runner func() error

func Watch(Run Runner, files []string, label string, wait time.Duration, quit chan bool, verbose bool) (error) {
	// now loop
	debounce := NewDebouncer(wait)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		// this is to prevent more than one running at a time
		//var tellDone chan bool
		//tellDone = make(chan bool)

		// watching loop
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					if verbose {
						fmt.Println("event not ok", event)
					}
					break
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					// I think this is so we don't start a new
					// debounce while we are waiting for a code gen to run
					// (later) because CUE does not have a way to cancel yet
					//if tellDone != nil {
					//  tellDone <- true
					//}

					debounce(func() {
						// kill previous sub-spawn (xcue)
						if verbose {
							fmt.Printf("regen (%s)\n", label)
						}
						start := time.Now()

						// CORE OPERATION
						err = Run()

						end := time.Now()

						elapsed := end.Sub(start).Round(time.Millisecond)
						if verbose {
							fmt.Printf(" done (%s) %v\n", label, elapsed)
						}

						if verbose && err != nil {
							fmt.Println("error:", err)
						}

						//tellDone <- true
					})
				}

			case err, ok := <-watcher.Errors:
				if verbose {
					fmt.Println("error:", err)
				}
				if !ok {
					break
				}

			case <-quit:
				return
			}
		}
	}()

	for _, file := range files {
		err = watcher.Add(file)
		if err != nil {
			return err
		}
	}
	// fmt.Printf("watching (%s) %d files\n", label, len(files))

	wg.Wait()

	return nil
}

func NewDebouncer(after time.Duration) func(f func()) {
	d := &debouncer{after: after}

	return func(f func()) {
		d.add(f)
	}
}

type debouncer struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer
}

func (d *debouncer) add(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.after, f)
}

