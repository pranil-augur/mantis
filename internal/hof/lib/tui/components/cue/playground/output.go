/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package playground

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"github.com/parnurzeal/gorequest"

	"github.com/opentofu/opentofu/internal/hof/lib/tui"
	"github.com/opentofu/opentofu/internal/hof/lib/yagu"
)

const HTTP2_GOAWAY_CHECK = "http2: server sent GOAWAY and closed the connection"

func (C *Playground) PushToPlayground() (string, error) {
	src := C.edit.GetText()

	url := "https://cuelang.org/.netlify/functions/snippets"
	req := gorequest.New().Post(url)
	req.Set("Content-Type", "text/plain")
	req.Send(src)

	resp, body, errs := req.End()

	if len(errs) != 0 && !strings.Contains(errs[0].Error(), HTTP2_GOAWAY_CHECK) {
		fmt.Println("errs:", errs)
		fmt.Println("resp:", resp)
		fmt.Println("body:", body)
		return body, errs[0]
	}

	if len(errs) != 0 || resp.StatusCode >= 500 {
		return body, fmt.Errorf("Internal Error: " + body)
	}
	if resp.StatusCode >= 400 {
		return body, fmt.Errorf("Bad Request: " + body)
	}

	return body, nil
}

func (C *Playground) WriteEditToFile(filename string) (error) {
	src := C.edit.GetText()

	return os.WriteFile(filename, []byte(src), 0644)
}

func (C *Playground) ExportFinalToFile(filename string) (error) {
	ext := filepath.Ext(filename)
	ext = strings.TrimPrefix(ext, ".")
	src, err := C.final.GetValueText(ext)
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)
	err = yagu.Mkdir(dir)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, []byte(src), 0644)
}

func (C *Playground) GetValue() cue.Value {
	tui.Log("trace", fmt.Sprintf("Play.GetConnValue from: %s/%s", C.Id(), C.Name()))

	return C.final.GetValue()
}

func (C *Playground) GetValueExpr(expr string) func () cue.Value {
	tui.Log("trace", fmt.Sprintf("Play.GetConnValueExpr from: %s/%s %s", C.Id(), C.Name(), expr))
	p := cue.ParsePath(expr)

	return func() cue.Value {
		return C.final.GetValue().LookupPath(p)
	}
}

