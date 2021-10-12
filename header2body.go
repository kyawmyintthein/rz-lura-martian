package rzluramartian

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/martian/parse"
)

type (
	HeaderModifierConfig struct {
		KeysToExtract []string `json:"keys_to_extract"`
	}

	Header2BodyModifier struct {
		keysToExtract []string
	}
)

func headerModifierFromJSON(b []byte) (*parse.Result, error) {
	cfg := &HeaderModifierConfig{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	mod := &Header2BodyModifier{
		keysToExtract: cfg.KeysToExtract,
	}
	return parse.NewResult(mod, []parse.ModifierType{parse.Request})
}

func (m *Header2BodyModifier) ModifyRequest(req *http.Request) error {
	var buf []byte
	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		data := make(map[string]interface{})
		err = json.Unmarshal(body, &data)
		if err != nil {
			return err
		}

		for _, k := range m.keysToExtract {
			data[k] = req.Header.Get(k)
			req.Header.Del(k)
		}

		buf, err = json.Marshal(data)
		if err != nil {
			return err
		}
	} else {
		data := make(map[string]interface{})
		for _, k := range m.keysToExtract {
			data[k] = req.Header.Get(k)
			req.Header.Del(k)
		}

		var err error
		buf, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	req.ContentLength = int64(len(buf))
	req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	return nil
}
