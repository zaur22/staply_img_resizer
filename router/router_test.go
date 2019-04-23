package router

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type ResizerMock struct {
	Err     error
	Entered string
}

func (r *ResizerMock) FromUrl(url string) error {
	r.Entered = url
	return r.Err
}

func (r *ResizerMock) ResizeImg(img []byte) error {
	r.Entered = string(img)
	return r.Err
}

func TestRouterGet(t *testing.T) {
	testCases := []struct {
		URLValues          url.Values
		Resizer            ResizerMock
		ExpectedStatusCode int
		ExpectedEnter      string
		ExpectedBody       string
	}{
		{
			URLValues:          createVals("url", "someUrl"),
			Resizer:            ResizerMock{},
			ExpectedStatusCode: http.StatusOK,
			ExpectedEnter:      "someUrl",
		},
		{
			URLValues:          url.Values{},
			Resizer:            ResizerMock{},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedBody:       "Missing required parameter 'url'",
		},
		{
			URLValues: createVals("url", "anotherUrl"),
			Resizer: ResizerMock{
				Err: fmt.Errorf("some err"),
			},
			ExpectedEnter:      "anotherUrl",
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedBody:       "some err",
		},
	}

	for _, tCase := range testCases {
		v := tCase.URLValues.Encode()
		u, _ := url.Parse("https://example.org?" + v)
		req := httptest.NewRequest(http.MethodGet, u.String(), nil)
		router := NewRouter(&tCase.Resizer)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != tCase.ExpectedStatusCode {
			t.Errorf("Bad status code. Expected '%v', got '%v'", tCase.ExpectedStatusCode, w.Code)
		}

		if w.Body.String() != tCase.ExpectedBody {
			t.Errorf("Bad body value. Expected '%v', got '%v'", tCase.ExpectedBody, w.Body.String())
		}

		if tCase.Resizer.Entered != tCase.ExpectedEnter {
			t.Errorf("Bad value entered to resizer method. Expected '%v', got '%v'", tCase.ExpectedEnter, tCase.Resizer.Entered)
		}

	}
}

func TestMultiPartFormImg(t *testing.T) {
	testCases := []struct {
		fieldName          string
		fileVal            string
		Resizer            ResizerMock
		ExpectedStatusCode int
		ExpectedEnter      string
		ExpectedBody       string
	}{
		{
			fieldName:          "image",
			fileVal:            "some bytes",
			Resizer:            ResizerMock{},
			ExpectedStatusCode: http.StatusOK,
			ExpectedEnter:      "some bytes",
		},
		{
			fieldName:          "noImage",
			fileVal:            "some bytes",
			Resizer:            ResizerMock{},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedEnter:      "",
			ExpectedBody:       "Can't take image: http: no such file",
		},
		{
			fieldName: "image",
			fileVal:   "some bytes",
			Resizer: ResizerMock{
				Err: fmt.Errorf("some error"),
			},
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedEnter:      "some bytes",
			ExpectedBody:       "some error",
		},
	}

	for _, tCase := range testCases {

		b, bodyWriter := multipartBody(tCase.fieldName, tCase.fileVal)

		req := httptest.NewRequest(
			http.MethodPost,
			"https://example.org",
			b,
		)

		req.Header.Set("Content-Type",
			bodyWriter.FormDataContentType())
		router := NewRouter(&tCase.Resizer)
		w := httptest.NewRecorder()

		err := bodyWriter.Close()
		if err != nil {
			fmt.Print(err.Error())
		}

		router.ServeHTTP(w, req)

		if w.Code != tCase.ExpectedStatusCode {
			t.Errorf("Bad status code. Expected '%v', got '%v'", tCase.ExpectedStatusCode, w.Code)
		}

		if w.Body.String() != tCase.ExpectedBody {
			t.Errorf("Bad body value. Expected '%v', got '%v'", tCase.ExpectedBody, w.Body.String())
		}

		if tCase.Resizer.Entered != tCase.ExpectedEnter {
			t.Errorf("Bad value entered to resizer method. Expected '%v', got '%v'", tCase.ExpectedEnter, tCase.Resizer.Entered)
		}

	}
}

func TestJSONImg(t *testing.T) {
	testCases := []struct {
		fieldName          string
		fieldVal           string
		Resizer            ResizerMock
		ExpectedStatusCode int
		ExpectedEnter      string
		ExpectedBody       string
	}{
		{
			fieldName:          "image",
			fieldVal:           "some bytes",
			Resizer:            ResizerMock{},
			ExpectedStatusCode: http.StatusOK,
			ExpectedEnter:      "some bytes",
		},
		{
			fieldName:          "anotherField",
			fieldVal:           "some bytes",
			Resizer:            ResizerMock{},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedBody:       "Field 'image' cannot be empty\n",
		},
		{
			fieldName: "image",
			fieldVal:  "some bytes",
			Resizer: ResizerMock{
				Err: fmt.Errorf("Bad val"),
			},
			ExpectedEnter:      "some bytes",
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedBody:       "Bad val",
		},
	}

	for _, tCase := range testCases {

		jsonVal := jsonBody(tCase.fieldName, tCase.fieldVal)

		req := httptest.NewRequest(
			http.MethodPost,
			"https://example.org",
			bytes.NewReader(jsonVal),
		)
		req.Header.Set("Content-Type", "application/json")
		router := NewRouter(&tCase.Resizer)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != tCase.ExpectedStatusCode {
			t.Errorf("Bad status code. Expected '%v', got '%v'", tCase.ExpectedStatusCode, w.Code)
		}

		if w.Body.String() != tCase.ExpectedBody {
			t.Errorf("Bad body value. Expected '%v', got '%v'", tCase.ExpectedBody, w.Body.String())
		}

		if tCase.Resizer.Entered != tCase.ExpectedEnter {
			t.Errorf("Bad value entered to resizer method. Expected '%v', got '%v'", tCase.ExpectedEnter, tCase.Resizer.Entered)
		}

	}
}

func createVals(name string, vals ...string) url.Values {
	var urlVals = url.Values{}
	for _, v := range vals {
		urlVals.Add(name, v)
	}
	return urlVals
}

func multipartBody(formName string, val string) (*bytes.Buffer, *multipart.Writer) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile(formName, "filename.jpeg")
	if err != nil {
		fmt.Print(err.Error())
	}

	br := bytes.NewReader([]byte(val))
	io.Copy(fileWriter, br)

	return bodyBuf, bodyWriter
}

func jsonBody(formName string, val string) []byte {
	jsonVal, err := json.Marshal(map[string]string{
		formName: base64.StdEncoding.
			EncodeToString([]byte(val)),
	})
	if err != nil {
		fmt.Print(err.Error())
	}
	return jsonVal
}
