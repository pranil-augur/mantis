/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package runtime

// scriptCmds are the script command implementations.
var scriptCmds = map[string]func(*Script, int, []string){

	// cmd_call.go
	"call": (*Script).CmdCall,

	// cmd_checks.go
	"status": (*Script).CmdStatus,

	// cmd_cmp.go
	"cmp":    (*Script).CmdCmp,
	"cmpenv": (*Script).CmdCmpenv,

	// cmd_env.go
	"env":    (*Script).CmdEnv,
	"envsub": (*Script).CmdEnvsub,

	// cmd_exec.go
	"exec": (*Script).CmdExec,
	"wait": (*Script).CmdWait,
	"skip": (*Script).CmdSkip,
	"stop": (*Script).CmdStop,

	// cmd_fs.go
	"cd":      (*Script).CmdCd,
	"chmod":   (*Script).CmdChmod,
	"cp":      (*Script).CmdCp,
	"exists":  (*Script).CmdExists,
	"mkdir":   (*Script).CmdMkdir,
	"rm":      (*Script).CmdRm,
	"symlink": (*Script).CmdSymlink,

	// cmd_http.go
	"http": (*Script).CmdHttp,

	// cmd_log.go
	"log": (*Script).CmdLog,

	// cmd_stdio.go
	"stdin":  (*Script).CmdStdin,
	"stderr": (*Script).CmdStderr,
	"stdout": (*Script).CmdStdout,

	// cmd_str.go
	"grep":   (*Script).CmdGrep,
	"regexp": (*Script).CmdRegexp,
	"sed":    (*Script).CmdSed,

	// other
	"unquote": (*Script).CmdUnquote,
}
