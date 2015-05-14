package web2

import "net/http"

var RenderJSON = RenderFunc(func(res *Response, w http.ResponseWriter, r *http.Request) error {

	if err := writeResonse(w, res, FormValueDefault(r, "callback", "")); err != nil {

		WriteError(w, "Error sending response", FormValueDefault(r, "callback", ""))
	}

	return nil

})
