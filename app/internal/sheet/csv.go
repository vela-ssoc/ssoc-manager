package sheet

import (
	"bytes"
	"encoding/csv"
	"io"
	"mime"
)

type CSVReader interface {
	UTF8BOM() bool
	Filename() string
	Header() []string
	Next() ([][]string, error)
}

type CSVStreamer interface {
	io.Reader
	MIME() string
	Disposition() string
}

func NewCSV(rd CSVReader) CSVStreamer {
	var bom bool
	if rd != nil {
		bom = rd.UTF8BOM()
	}

	buf := new(bytes.Buffer)
	if bom {
		// UTF-8 BOM: https://learn.microsoft.com/en-us/globalization/encoding/byte-order-mark
		buf.Write([]byte{0xef, 0xbb, 0xbf})
	}
	cvw := csv.NewWriter(buf)

	return &csvStream{
		read: rd,
		buf:  buf,
		cvw:  cvw,
	}
}

type csvStream struct {
	read  CSVReader
	wrote bool
	eof   bool
	buf   *bytes.Buffer
	cvw   *csv.Writer
}

func (cs *csvStream) Read(p []byte) (int, error) {
	if cs.read == nil || cs.eof {
		return 0, io.EOF
	}
	if !cs.wrote {
		cs.wrote = true
		header := cs.read.Header()
		if err := cs.cvw.Write(header); err != nil {
			return 0, err
		}
	}

	psz, wsz := len(p), 0
	for wsz < psz {
		if cs.buf.Len() == 0 {
			next, err := cs.read.Next()
			if err != nil {
				cs.eof = true
				if wsz != 0 {
					return wsz, nil
				}
				return 0, err
			}
			if len(next) == 0 {
				continue
			}
			if err = cs.cvw.WriteAll(next); err != nil {
				return 0, err
			}
		}
		n, err := cs.buf.Read(p[wsz:])
		if err != nil {
			return wsz, err
		}
		wsz += n
	}

	return wsz, nil
}

func (cs *csvStream) MIME() string {
	return "text/csv; charset=utf-8"
}

func (cs *csvStream) Disposition() string {
	name := "unnamed.csv"
	if r := cs.read; r != nil {
		if fn := r.Filename(); fn != "" {
			name = fn
		}
	}
	return mime.FormatMediaType("attachment", map[string]string{"filename": name})
}
