/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package yagu

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenBrowserCmd(url string) (*exec.Cmd, error) {
	var args []string

	switch runtime.GOOS {

	case "linux":
		args = []string{"xdg-open", url}
	case "windows":
		args = []string{"rundll32", "url.dll,FileProtocolHandler", url}
		// args = []string{"cmd", "/C", "start", url}
	case "darwin":
		args = []string{"open", url}

	default:
		return nil, fmt.Errorf("unsupported platform")

	}

	return exec.Command(args[0], args[1:]...), nil
}

func OpenBrowserCmdSafe(url string) (error) {
	cmd, err := OpenBrowserCmd(url)
	if err != nil {
		fmt.Println(url)
		return nil
	}

	return cmd.Run()
}
