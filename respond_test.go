package respond_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/respond"
	"github.com/stretchr/testify/require"
)

var testObject1 = map[string]interface{}{"one": 1, "yes": true, "name": "Stretchr"}

func TestRespond(t *testing.T) {

	for i, test := range []struct {
		with       respond.With
		expStatus  int
		expHeaders http.Header
		r          *http.Request
		assertions func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			with:      respond.With{},
			expStatus: 200,
		},
		{
			with: respond.With{
				Data: testObject1,
			},
			expStatus: 200,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				obj := jsonobj(w.Body.Bytes()).(map[string]interface{})
				require.Equal(t, obj["name"], testObject1["name"])
				require.Equal(t, obj["yes"], testObject1["yes"])
				require.Equal(t, obj["one"], testObject1["one"])
			},
		},
		{
			with: respond.With{
				Status: http.StatusCreated,
			},
			expStatus: http.StatusCreated,
		},
	} {

		// make sure we have a request
		if test.r == nil {
			test.r, _ = http.NewRequest("GET", "/", nil)
			test.r.Header.Set("Content-Type", "application/json; charset=utf8")
		}
		w := httptest.NewRecorder()
		err := test.with.To(w, test.r)
		require.NoError(t, err, "Error (test %d)", i) // TODO: fix
		if !w.Flushed {
			w.Flush()
		}

		// status
		if test.expStatus > 0 {
			require.Equal(t, test.expStatus, w.Code, "StatusCode (test %d)", i)
		}

		// headers
		if test.expHeaders != nil {
			for header, expValues := range test.expHeaders {
				require.Equal(t, expValues, w.HeaderMap[header], "Header %s (test %d)", header, i)
			}
		}

		// additional assertions
		if test.assertions != nil {
			test.assertions(t, w)
		}

	}

}

func jsonobj(b []byte) interface{} {
	var v interface{}
	json.NewDecoder(bytes.NewReader(b)).Decode(&v)
	return v
}

func TestDefaultWriteHeader(t *testing.T) {

	w := httptest.NewRecorder()
	respond.DefaultOptions.WriteHeader(&respond.Ctx{W: w, With: &respond.With{}}, http.StatusTeapot)
	require.Equal(t, http.StatusTeapot, w.Code)

}

func TestDefaultWriteData(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	data := map[string]interface{}{"one": 1}
	err := respond.DefaultOptions.WriteData(&respond.Ctx{W: w, R: r, With: &respond.With{}}, data)
	require.NoError(t, err)
	require.Equal(t, w.Body.String(), "{\"one\":1}\n")

}

func TestDefaultEncoder(t *testing.T) {

	w := httptest.NewRecorder()

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept", "application/json")
	e, err := respond.DefaultOptions.Encoder(&respond.Ctx{W: w, R: r, With: &respond.With{}})
	require.NoError(t, err)
	require.Equal(t, e, respond.JSONEncoder)

	r, _ = http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept", "")
	e, err = respond.DefaultOptions.Encoder(&respond.Ctx{W: w, R: r, With: &respond.With{}})
	require.NoError(t, err)
	require.Equal(t, e, respond.JSONEncoder)

}

func TestJSONEncoder(t *testing.T) {

	var buf bytes.Buffer
	data := map[string]interface{}{"one": 1}
	respond.JSONEncoder.Encode(&buf, data)
	require.Equal(t, buf.String(), "{\"one\":1}\n")

}

func TestSetHeadersOverride(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	opts := respond.DefaultOptions.Copy()
	opts.DefaultHeaders = map[string][]string{"X-App-Source": []string{"default"}}
	opts.SetHeaders = respond.SetHeadersOverride
	respond.With{
		Options: opts,
		Headers: map[string][]string{"X-App-Source": []string{"explicit"}},
	}.To(w, r)

	require.Equal(t, 1, len(w.HeaderMap["X-App-Source"]))
	require.Equal(t, "explicit", w.HeaderMap["X-App-Source"][0])

}

func TestSetHeadersAggregate(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	opts := respond.DefaultOptions.Copy()
	opts.DefaultHeaders = map[string][]string{"X-App-Source": []string{"default"}}
	opts.SetHeaders = respond.SetHeadersAggregate
	respond.With{
		Options: opts,
		Headers: map[string][]string{"X-App-Source": []string{"explicit"}},
	}.To(w, r)

	require.Equal(t, 2, len(w.HeaderMap["X-App-Source"]))
	require.Equal(t, "default", w.HeaderMap["X-App-Source"][0])
	require.Equal(t, "explicit", w.HeaderMap["X-App-Source"][1])

}

func TestDefaultHeaders(t *testing.T) {

	respond.DefaultOptions.DefaultHeaders = map[string][]string{"X-App-Version": []string{"1.0"}}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	respond.With{}.To(w, r)

	require.Equal(t, "1.0", w.HeaderMap.Get("X-App-Version"))

}
