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

package db

import (
	"database/sql"

	"cuelang.org/go/cue"

	_ "github.com/lib/pq"
)

func handlePostgresExec(dbname, query string, args []interface{}) (string, error) {
	db, err := sql.Open("postgres", dbname)
	if err != nil {
		return "", err
	}
	return handleExec(db, query, args)
}

func handlePostgresQuery(dbname, query string, args []interface{}) (*sql.Rows, error) {
	db, err := sql.Open("postgres", dbname)
	if err != nil {
		return nil, err
	}
	return handleQuery(db, query, args)
}
func handlePostgresStmts(dbname string, stmts cue.Value, args []interface{}) (cue.Value, error) {
	db, err := sql.Open("postgres", dbname)
	if err != nil {
		return stmts, err
	}
	return handleStmts(db, stmts, args)
}
