package respond

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// With describes the kind of response that will be made.
// respond.With{ [options] }.To(w, r)
type With struct {
	// Data is the data object to respond with.
	Data interface{}
	// Status is the HTTP status to respond with.  Use http.Status*
	Status int
	// Options are the Options to use when responding.
	// DefaultOptions are used by default.
	Options *Options
}

// To writes the response from R to the specified http.ResponseWriter,
// referring to details in the http.Request where appropriate.
func (with With) To(w http.ResponseWriter, r *http.Request) error {
	pWith := &with
	opts := options(pWith)
	ctx := &Ctx{W: w, R: r, With: pWith}
	opts.WriteHeader(ctx, status(pWith))
	if pWith.Data != nil {
		if err := opts.WriteData(ctx, pWith.Data); err != nil {
			return err
		}
	}
	return nil
}

// status gets the status code that should be written
// for the specified With data.
func status(with *With) int {
	const NoStatus = 0
	if with.Status != NoStatus {
		return with.Status
	}
	return options(with).DefaultStatus
}

func options(with *With) *Options {
	if with.Options != nil {
		return with.Options
	}
	return DefaultOptions
}

// Options represents the options with which responses are
// handled.
type Options struct {
	// Encoders represents a map of content types to Encoder objects.
	Encoders map[string]Encoder
	// WriteHeader writes the header to the http.ResponseWriter.
	WriteHeader func(c *Ctx, code int)
	// WriteData writes the data to the http.ResponseWriter.
	WriteData func(c *Ctx, data interface{}) error
	// Encoder gets the Encoder to write the response data with.
	Encoder func(c *Ctx) (Encoder, error)
	// DefaultStatus is the default http.Status to use when none
	// is specified.
	DefaultStatus int
	// DefaultEncoder is the default Encoder to use when none
	// is specified.
	DefaultEncoder Encoder
}

// DefaultOptions represents the default options that will be
// used when responding.
// Properties of DefaultOptions may be changed directly, or else
// you can set specific options in each With object.
var DefaultOptions *Options

func init() {
	JSONEncoder = (*jsonEncoder)(nil)
	DefaultOptions = &Options{
		DefaultStatus: http.StatusOK,
		WriteHeader: func(c *Ctx, code int) {
			c.W.WriteHeader(code)
		},
		WriteData: func(c *Ctx, data interface{}) error {
			enc, err := options(c.With).Encoder(c)
			if err != nil {
				return err
			}
			return enc.Encode(c.W, data)
		},
		Encoders:       map[string]Encoder{"application/json": JSONEncoder},
		DefaultEncoder: JSONEncoder,
	}
	DefaultOptions.Encoder = func(c *Ctx) (Encoder, error) {
		opts := options(c.With)
		accept := c.R.Header.Get("Accept")
		for contentType, enc := range opts.Encoders {
			if strings.Contains(accept, contentType) {
				return enc, nil
			}
		}
		return DefaultOptions.DefaultEncoder, nil
	}
}

// Ctx wraps the http.ResponseWriter, http.Request and
// respond.With objects.
// Used when overriding DefaultOptions.
type Ctx struct {
	// W is the http.ResponseWriter.
	W http.ResponseWriter
	// R is the http.Request.
	R *http.Request
	// With is the With object.
	With *With
}

// Encoder represents an object capable of encoding data
// to an io.Writer.
type Encoder interface {
	// Encode writes the data to the writer, returns an error
	// if something went wrong, otherwise nil.
	Encode(w io.Writer, data interface{}) error
}

type jsonEncoder struct{}

func (_ *jsonEncoder) Encode(w io.Writer, data interface{}) error {
	return json.NewEncoder(w).Encode(data)
}

// JSONEncoder represents an Encoder capable of writing
// JSON.
var JSONEncoder Encoder // assigned to in init
