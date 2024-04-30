/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package panel


func (P* Panel) RangeItems(fn func (PanelItem)) {

	for _, item := range P.GetItems() {
		fi := item.Item // flexItem.Item
		switch t := fi.(type) {
		case *Panel:
			t.RangeItems(fn)
		case PanelItem:
			fn(t)
		}
	}
}
