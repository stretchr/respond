package respond

import (
	"encoding/json"
	"net/http"
)

// With describes the kind of response that will be made.
// respond.With{ [options] }.To(w, r)
type With struct {
	// Data is the data object to respond with.
	Data interface{}
	// Status is the HTTP status to respond with.  Use http.Status*
	Status int
}

// To writes the response from R to the specified http.ResponseWriter,
// referring to details in the http.Request where appropriate.
func (with With) To(w http.ResponseWriter, r *http.Request) error {
	pWith := &with
	w.WriteHeader(status(pWith))
	if pWith.Data != nil {
		if err := json.NewEncoder(w).Encode(pWith.Data); err != nil {
			return err
		}
	}
	return nil
}

// status gets the status code that should be written
// for the specified With data.
func status(with *With) int {
	const NoStatus = 0
	if with.Status == NoStatus {
		return http.StatusOK // TODO: put in default config
	}
	return with.Status
}
