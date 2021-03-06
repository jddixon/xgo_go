package xmlpull

import (
	"fmt"
	"io"
)

// xgo/xml/xmlpull/parseMisc.go

// Accept zero or one Misc productions, returning miscFound == true if one
// is found.
//
func (p *Parser) acceptMisc() (miscFound bool, curEvent PullEvent, err error) {

	var found bool

	// handle for comment is '<!-' --------------------------
	found, err = p.AcceptStr("<!-")
	if err == nil && found {
		// DEBUG
		fmt.Println("state XML_DECL_SEEN: found COMMENT")
		// END
		err = p.parseComment()
		if err == nil {
			curEvent = COMMENT
			miscFound = true
		}
	}
	// handle for PI is '<?' --------------------------------
	if !miscFound && err == nil {
		found, err = p.AcceptStr("<?")
		if err == nil && found {
			// DEBUG
			fmt.Println("found PROCESSING_INSTRUCTION")
			// END
			found, err = p.parsePI()
			if err == nil && found {
				curEvent = PROCESSING_INSTRUCTION
				miscFound = true
			}
		}
	}
	if !miscFound && err == nil {
		p.text = p.text[:0] // clear the slice

		// handle for S is IsS() --------------------------------
		var ch rune
		ch, err = p.NextCh()
		chIsS := p.IsS(ch)
		for err == nil && chIsS {
			// DEBUG
			fmt.Printf("    got S 0x%02x, err %v\n", ch, err)
			// END
			p.text = append(p.text, ch) // ACCUMULATING WHITESPACE IN text
			ch, err = p.NextCh()
			chIsS = p.IsS(ch)
		}
		// DEBUG
		fmt.Printf("S-count = %d\n", len(p.text))
		// END
		if (err == nil) || (err == io.EOF) {
			if err == io.EOF {
				err = nil
			} else {
				// DEBUG
				fmt.Printf("  pushing back 0x%02x\n", ch)
				// END
				if !chIsS {
					p.PushBack(ch)
				}
			}
			if len(p.text) > 0 {
				curEvent = IGNORABLE_WHITESPACE
				miscFound = true
				// DEBUG
				fmt.Printf("  exiting acceptMisc(): IGNORABLE, len %d, '%s'\n",
					len(p.text), string(p.text))
				// END
			}
		}
	}
	// DEBUG
	fmt.Printf("acceptMisc returning curEvent %d, error %v\n",
		curEvent, err)
	// END
	return
}
