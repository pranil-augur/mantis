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

import (
	"github.com/opentofu/opentofu/internal/hof/lib/tui/components/widget"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/tview"
)

func (P *Panel) Encode() (map[string]any, error) {
	m := make(map[string]any)

	m["id"] = P._cnt
	m["name"] = P._name
	m["typename"] = "panel"
	m["direction"] = P.GetDirection()

	items := []map[string]any{}

	for _, item := range P.GetItems() {
		var (
			d map[string]any
			i map[string]any
			err error
		)
		i = make(map[string]any)
		i["flexFixedSize"] = item.FixedSize
		i["flexProportion"] = item.Proportion

		switch item := item.Item.(type) {
		case *Panel:
			// recursion, within the current panel stack
			d, err = item.Encode()
		case PanelItem:
			// leaf, but may also recurse if a widget uses panels
			d, err = item.Encode()

		default:
			panic("unhandled item type in panel")	
		}

		if err != nil {
			return nil, err
		}

		i["flexItem"] = d

		// add the item to output
		items = append(items, i)
	}

	m["items"] = items

	return m, nil
}

// dummy
func (I *Panel) Decode(map[string]any) (widget.Widget, error) {
	return nil, nil
}


func PanelDecodeMap(input map[string]any, parent *Panel, creator ItemCreator) (*Panel, error) {
	// tui.Log("extra", fmt.Sprintf("Panel.Decode: %# v", input))
	P := &Panel{
		Flex: tview.NewFlex(),
		_creator: creator,
		_parent: parent,
		_cnt: input["id"].(int),
		_name: input["name"].(string),
	}

	if items, ok := input["items"]; ok {
		for _, idata := range items.([]any) {
			imap := idata.(map[string]any)
			fsize, _ := imap["flexFixedSize"].(int)
			fprop, _ := imap["flexProportion"].(int)
			fmap, _ := imap["flexItem"].(map[string]any)

			I, err := ItemDecodeMap(fmap, P, creator)
			if err != nil {
				return P, err
			}
			P.AddItem(I, fsize, fprop, true)
		}
	}

	// do layout setup here, once some children have been instantiated
	P.SetDirection(input["direction"].(int))
	P.SetBorder(true)
	P.SetTitle(P.TitleString())

	// tui.Log("trace", fmt.Sprintf("panel... %v %v", P.Id(), P.TitleString()))

	return P, nil
}


