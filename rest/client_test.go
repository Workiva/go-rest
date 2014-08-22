package rest

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

type testObj struct{ Prop string }

type testAuthorizer struct{ key, secret string }

// Authorize will provide a testing specific authorization without
// actually signing the request.
func (tc testAuthorizer) Authorize(urlStr string, requestType string, form url.Values) url.Values {
	baseParams := map[string]string{
		"oauth_consumer_key":     tc.key,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_version":          "1.0",
		"oauth_nonce":            strconv.FormatInt(rand.Int63(), 10),
	}
	for param, value := range baseParams {
		form.Set(param, value)
	}
	result := requestType + "&" + url.QueryEscape(urlStr)
	baseString := result + "&" + url.QueryEscape(form.Encode())
	form.Set("oauth_signature", baseString)
	return form
}

func TestEverything(t *testing.T) {
	returnJson := `{"Status": 200, "Reason": "", "Messages": [], "Next": "", "Results": {"Prop": "a"}}`

	formData := map[string]string{
		"oauth_consumer_key":     "token",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_version":          "1.0",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter, r *http.Request) {

		r.ParseForm()

		for _, prop := range []string{"oauth_nonce", "oauth_timestamp"} {
			if _, ok := r.Form[prop]; !ok {
				t.Errorf("Form property %s not found", prop)
			}
		}

		for key, value := range formData {
			if r.Form[key][0] != value {
				t.Errorf("Form data value %s, want %s", r.Form[key][0], value)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, returnJson)
	}))

	defer ts.Close()

	var c = Client{testAuthorizer{
		"token",
		"value",
	}}

	testObj := testObj{}
	fs := []func() (*BaseResponse, error){
		func() (*BaseResponse, error) {
			return c.GetJson(ts.URL, nil, &testObj)
		},
		func() (*BaseResponse, error) {
			return c.DeleteJson(ts.URL, nil, &testObj)
		},
		func() (*BaseResponse, error) {
			return c.PutJson(ts.URL, nil, &testObj)
		},
		func() (*BaseResponse, error) {
			return c.PostJson(ts.URL, nil, &testObj)
		},
	}

	for _, f := range fs {
		resp, err := f()
		if err != nil {
			t.Errorf("Error in request: %s\n", err)
		}

		if resp.Status != 200 {
			t.Errorf("REST response status %d, want %d", resp.Status, 200)
		}

		if resp.Reason != "" {
			t.Errorf("REST response reason %s, want %s", resp.Reason, "")
		}

		if len(resp.Messages) != 0 {
			t.Errorf("REST response message count %d, want %d", len(resp.Messages), 0)
		}

		if resp.Next != "" {
			t.Errorf("REST response next %s, want %s", resp.Next, "")
		}

		if testObj.Prop != "a" {
			t.Errorf("testObj prop %s, want %s", testObj.Prop, "a")
		}
	}
}
