package soap

import (
	"encoding/json"
	"fmt"
	"github.com/clbanning/mxj/x2j"
	"github.com/pkg/errors"
	"io"
)

type jsonDecoder struct {
	reader io.Reader
}

func (j *jsonDecoder) Decode(v interface{}) error {
	b, err := io.ReadAll(j.reader)
	if err != nil {
		return errors.Wrapf(err, "ReadAll")
	}
	jb, err := x2j.XmlToJson(b, false)
	if err != nil {
		return errors.Wrapf(err, "XmlToJson failed: %s", string(jb))
	}
	switch vv := v.(type) {
	case *[]uint8:
		*vv = jb
		return nil
	case []uint8:
		return fmt.Errorf("[]uint8 must be pointer")
	}

	if err = json.Unmarshal(jb, v); err != nil {
		return errors.Wrapf(err, "Unmarshal: %s", string(jb))
	}
	return nil
}

func newJsonDecoder(r io.Reader) *jsonDecoder {
	return &jsonDecoder{
		reader: r,
	}
}
