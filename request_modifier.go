package rzluramartian

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/clbanning/mxj"
	"github.com/google/martian/parse"
	"golang.org/x/net/html/charset"
)

const (
	_contentType_JSON = "json"
	_contentType_XML  = "xml"
)

type (
	RequestBodyModifierConfig struct {
		OriginalContentType string `json:"original_content_type"`
		ExpectedContentType string `json:"expected_content_type"`
	}

	RequestBodyModifier struct {
		originalContentType string
		expectedContentType string
	}
)

func (m *RequestBodyModifier) ModifyRequest(req *http.Request) error {
	if req.Body == nil {
		return nil
	}

	var v map[string]interface{}
	switch m.expectedContentType {
	case _contentType_XML:
		mxj.XmlCharsetReader = charset.NewReaderLabel
		mv, err := mxj.NewMapXmlReader(xmlReader{r: req.Body})
		if err != nil {
			return err
		}
		v = mv
	default:
		// Default is JSON
		payloadBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}
		req.Body.Close()
		err = json.Unmarshal(payloadBytes, &v)
		if err != nil {
			return err
		}
	}

	switch m.originalContentType {
	case _contentType_XML:
		transformedDataBytes, err := mxj.AnyXml(v)
		if err != nil {
			return err
		}

		req.Header.Del("Content-Type")
		req.Header.Set("Content-Type", "application/xml")
		req.ContentLength = int64(len(transformedDataBytes))
		req.Body = ioutil.NopCloser(bytes.NewReader(transformedDataBytes))
	default:
		transformedDataBytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		req.Header.Del("Content-Type")
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(transformedDataBytes))
		req.Body = ioutil.NopCloser(bytes.NewReader(transformedDataBytes))
	}

	return nil
}

type xmlReader struct {
	r io.Reader
}

func (x xmlReader) Read(p []byte) (n int, err error) {
	n, err = x.r.Read(p)

	if err != io.EOF {
		return n, err
	}

	if len(p) == n {
		return n, nil
	}

	p[n] = ([]byte("\n"))[0]
	return n + 1, err
}

func requestTransformerFromJSON(b []byte) (*parse.Result, error) {
	cfg := &RequestBodyModifierConfig{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	mod := &RequestBodyModifier{
		expectedContentType: cfg.ExpectedContentType,
		originalContentType: cfg.OriginalContentType,
	}
	return parse.NewResult(mod, []parse.ModifierType{parse.Request})
}
