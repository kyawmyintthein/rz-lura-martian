package rzluramartian

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/google/martian/parse"
)

type (
	QueryModifierConfig struct {
		KeysToExtract []string `json:"keys_to_extract"`
		Template      string   `json:"template"`
	}
	Query2BodyModifier struct {
		keysToExtract []string
		template      *template.Template
	}
)

func queryModifierFromJSON(b []byte) (*parse.Result, error) {
	cfg := &QueryModifierConfig{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	tmpl, err := template.New("query2body_modifier").Parse(cfg.Template)
	if err != nil {
		return nil, err
	}

	mod := &Query2BodyModifier{
		keysToExtract: cfg.KeysToExtract,
		template:      tmpl,
	}
	return parse.NewResult(mod, []parse.ModifierType{parse.Request})
}

func (m *Query2BodyModifier) ModifyRequest(req *http.Request) error {
	query := req.URL.Query()

	buf := new(bytes.Buffer)
	if err := m.template.Execute(buf, query); err != nil {
		return err
	}

	for _, k := range m.keysToExtract {
		query.Del(k)
	}

	req.ContentLength = int64(buf.Len())
	req.Body = ioutil.NopCloser(buf)
	req.URL.RawQuery = query.Encode()
	return nil
}
