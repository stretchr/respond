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
