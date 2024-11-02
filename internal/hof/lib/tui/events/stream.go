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

package events

import (
	"sync"
)

type EventStream struct {
	stream      chan Event
	wg          sync.WaitGroup
	sigStopLoop chan Event
	hook        func(Event)
	sources     sync.Map // chan Event
	handlers    sync.Map // func(Event)
}

func NewEventStream() *EventStream {
	return &EventStream{
		stream:      make(chan Event, 256),
		sigStopLoop: make(chan Event),
	}
}

func (es *EventStream) Init() {
	es.Merge("internal", es.sigStopLoop)
	go func() {
		es.wg.Wait()
		close(es.stream)
	}()
}

func (es *EventStream) Merge(name string, ec chan Event) {

	es.wg.Add(1)

	es.sources.Store(name, ec)

	go func(a chan Event) {
		for n := range a {
			n.From = name
			es.stream <- n
		}
		es.wg.Done()
	}(ec)
}

func (es *EventStream) AddHandler(path string, handler func(Event)) {
	n := cleanPath(path)
	es.handlers.Store(n, handler)
}

func (es *EventStream) RemoveHandle(path string) {
	n := cleanPath(path)
	es.handlers.Delete(n)
}

func (es *EventStream) match(path string) string {
	n := -1
	pattern := ""
	es.handlers.Range(func(key, value any) bool {
		m := key.(string)
		if !isPathMatch(m, path) {
			return true
		}
		if len(m) > n {
			pattern = m
			n = len(m)
		}
		return false
	}) 
	return pattern
}

func (es *EventStream) Hook(f func(Event)) {
	es.hook = f
}

func (es *EventStream) Loop() {
	for e := range es.stream {
		switch e.Path {
		case "/sig/stoploop":
			return
		}
		func(a Event) {
			if pattern := es.match(a.Path); pattern != "" {
				h, ok := es.handlers.Load(pattern)
				if ok {
					fn := h.(func(Event) )
					fn(a)
				}
			}
		}(e)

		if es.hook != nil {
			es.hook(e)
		}
	}

}

func (es *EventStream) StopLoop() {
	go func() {
		e := Event{
			Path: "/sig/stoploop",
		}
		es.sigStopLoop <- e
	}()
}


