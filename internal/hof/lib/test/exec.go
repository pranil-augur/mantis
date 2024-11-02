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

package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type BaseTester struct {
	Dir    string
	Env    map[string]string
	Sysenv bool
}

type BashTester struct {
	BaseTester

	Script string
}

func RunBash(T *Tester, verbose int) (err error) {
	// Decode our BT
	var BT BashTester
	err = T.Value.Decode(&BT)

	// Check for errors and validate
	if err != nil {
		return err
	}
	if BT.Script == "" {
		return fmt.Errorf("Bash tester %q has empty script field", T.Name)
	}

	// Prep our command
	cmd := exec.Command("bash", "-p", "-c", BT.Script)
	cmd.Dir = BT.Dir

	// add env vars if needed
	if len(BT.Env) > 0 {
		for k, v := range BT.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Setup io streams
	cmd.Stdin = os.Stdin

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &outBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)

	// Run and save output
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	T.Output = outBuf.String() + errBuf.String()

	return err
}

type ExecTester struct {
	BaseTester

	Command string
}

func RunExec(T *Tester, verbose int) (err error) {
	// Decode our ET
	var ET ExecTester
	err = T.Value.Decode(&ET)

	// Check for errors and validate
	if err != nil {
		return err
	}
	if ET.Command == "" {
		return fmt.Errorf("Bash tester %q has empty script field", T.Name)
	}

	args := strings.Fields(ET.Command)

	// Prep our command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = ET.Dir

	// add env vars if needed
	if len(ET.Env) > 0 {
		for k, v := range ET.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Setup io streams
	cmd.Stdin = os.Stdin

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &outBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)

	// Run and save output
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	T.Output = outBuf.String() + errBuf.String()

	return err
}
