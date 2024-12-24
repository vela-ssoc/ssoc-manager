// MIT License

// Copyright (c) 2019 Muhammad Muzzammil

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package jsonc

const (
	charESCAPE   = 92
	charQUOTE    = 34
	charSPACE    = 32
	charTAB      = 9
	charNEWLINE  = 10
	charASTERISK = 42
	charSLASH    = 47
	charHASH     = 35
)

func Translate(s []byte) []byte {
	var i int
	var quote bool
	var escaped bool

	j := make([]byte, len(s))
	comment := &commentData{}
	for _, ch := range s {
		if ch == charESCAPE || escaped {
			if !comment.startted {
				j[i] = ch
				i++
			}
			escaped = !escaped
			continue
		}
		if ch == charQUOTE && !comment.startted {
			quote = !quote
		}
		if (ch == charSPACE || ch == charTAB) && !quote {
			continue
		}
		if ch == charNEWLINE {
			if comment.isSingleLined {
				comment.stop()
			}
			continue
		}
		if quote && !comment.startted {
			j[i] = ch
			i++
			continue
		}
		if comment.startted {
			if ch == charASTERISK && !comment.isSingleLined {
				comment.canEnd = true
				continue
			}
			if comment.canEnd && ch == charSLASH && !comment.isSingleLined {
				comment.stop()
				continue
			}
			comment.canEnd = false
			continue
		}
		if comment.canStart && (ch == charASTERISK || ch == charSLASH) {
			comment.start(ch)
			continue
		}
		if ch == charSLASH {
			comment.canStart = true
			continue
		}
		if ch == charHASH {
			comment.start(ch)
			continue
		}
		j[i] = ch
		i++
	}
	return j[:i]
}

type commentData struct {
	canStart      bool
	canEnd        bool
	startted      bool
	isSingleLined bool
	endLine       int
}

func (c *commentData) stop() {
	c.startted = false
	c.canStart = false
}

func (c *commentData) start(ch byte) {
	c.startted = true
	c.isSingleLined = ch == charSLASH || ch == charHASH
}
