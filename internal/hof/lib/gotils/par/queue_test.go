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

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package par

import (
	"sync"
	"testing"
)

func TestQueueIdle(t *testing.T) {
	q := NewQueue(1)
	select {
	case <-q.Idle():
	default:
		t.Errorf("NewQueue(1) is not initially idle.")
	}

	started := make(chan struct{})
	unblock := make(chan struct{})
	q.Add(func() {
		close(started)
		<-unblock
	})

	<-started
	idle := q.Idle()
	select {
	case <-idle:
		t.Errorf("NewQueue(1) is marked idle while processing work.")
	default:
	}

	close(unblock)
	<-idle // Should be closed as soon as the Add callback returns.
}

func TestQueueBacklog(t *testing.T) {
	const (
		maxActive = 2
		totalWork = 3 * maxActive
	)

	q := NewQueue(maxActive)
	t.Logf("q = NewQueue(%d)", maxActive)

	var wg sync.WaitGroup
	wg.Add(totalWork)
	started := make([]chan struct{}, totalWork)
	unblock := make(chan struct{})
	for i := range started {
		started[i] = make(chan struct{})
		i := i
		q.Add(func() {
			close(started[i])
			<-unblock
			wg.Done()
		})
	}

	for i, c := range started {
		if i < maxActive {
			<-c // Work item i should be started immediately.
		} else {
			select {
			case <-c:
				t.Errorf("Work item %d started before previous items finished.", i)
			default:
			}
		}
	}

	close(unblock)
	for _, c := range started[maxActive:] {
		<-c
	}
	wg.Wait()
}
