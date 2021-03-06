package xmlpull

// xgo/xml/xmlpull/parseComment.go

import (
	"fmt"
	"strings"
)

var _ = fmt.Print

// XML 1.0 Section 2.5 Comments
//
// [15] Comment :== '<!--' ((Char - '-') | ('-' (Char - '-')))* '-->'
//
// Grammer does not allow ---> at end.
//
// Enter having seen "<!-".  Expect to see another '-' and then zero
// or more Char, possibly including zero or more '-' but no longer
// sequences of '-'.  Ends with '-->'.
//
// On return, p.commentChars contains the body of the comment.
//
func (p *Parser) parseComment() (err error) {

	// Enter having seen "<!-"

	var (
		commentChars            []rune
		endSeen                 bool
		haveDash, haveTwoDashes bool
	)
	ch, err := p.NextCh()
	if ch != '-' {
		// XXX Possibly ill-considered.  It would make more sense to
		// just give up, probably treating what's been seen as simple
		// text.
		err = p.NewXmlPullError("comment must start with <!--")
	}
	if err == nil {
		p.start()
		for err == nil {
			ch, err = p.NextCh()
			if err == nil {
				if haveTwoDashes && ch != '>' {
					err = p.NewXmlPullError(
						"cannot have two dashes within comment")
					break
				}
				if ch == '-' {
					if !haveDash {
						haveDash = true
					} else {
						haveTwoDashes = true
						haveDash = false
					}
				} else if ch == '>' {
					if haveTwoDashes {
						endSeen = true
						break // end of comment
					} else {
						haveTwoDashes = false
					}
					haveDash = false
				} else {
					if haveDash {
						commentChars = append(commentChars, '-')
						haveDash = false
					}
					// \r, \n handled the same way
					commentChars = append(commentChars, ch)
				}
			}
		}
	}
	if (err == nil && !endSeen) ||
		// not a terribly robust test
		(err != nil && strings.HasSuffix(err.Error(), ": EOF")) {
		err = p.NotClosedErr("comment")
	}
	if err == nil {
		p.commentChars = make([]rune, len(commentChars))
		copy(p.commentChars, commentChars)
	}
	return
}
