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
	// Headers represents explicit HTTP headers that will be send
	// along with the response.
	Headers http.Header
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
	// SetHeaders will set the headers on the ResponseWriter using the
	// DefaultHeaders first, followed by any explicit headers.
	SetHeaders func(c *Ctx)
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
	// DefaultHeaders represents the default HTTP headers to send
	// with each request. DefaultHeaders will be merged with the
	// explicit Headers when responding.
	// DefaultHeaders will be overridden by any explicit headers
	// if the keys match by default, do SetHeaders = SetHeadersAggregate
	// to add the values instead.
	DefaultHeaders http.Header
}

// Copy makes a copy of the options allowing the returning object
// to be modified without affecting the original.
func (o *Options) Copy() *Options {
	return &Options{
		Encoders:       o.Encoders,
		SetHeaders:     o.SetHeaders,
		WriteHeader:    o.WriteHeader,
		WriteData:      o.WriteData,
		Encoder:        o.Encoder,
		DefaultStatus:  o.DefaultStatus,
		DefaultEncoder: o.DefaultEncoder,
		DefaultHeaders: o.DefaultHeaders,
	}
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
		SetHeaders:    SetHeadersOverride,
		WriteHeader: func(c *Ctx, code int) {
			options(c.With).SetHeaders(c)
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

// SetHeadersOverride is a SetHeaders func that overrides
// DefaultHeaders with explicit ones.
func SetHeadersOverride(c *Ctx) {
	setHeaders(c, true)
}

// SetHeadersAggregate is a SetHeaders func that adds
// explicit headers to DefaultHeaders.
func SetHeadersAggregate(c *Ctx) {
	setHeaders(c, false)
}

// setHeaders sets the headers, optionally overriding the
// defaults or not.
func setHeaders(c *Ctx, override bool) {
	for key, vals := range options(c.With).DefaultHeaders {
		for _, val := range vals {
			c.W.Header().Add(key, val)
		}
	}
	for key, vals := range c.With.Headers {
		if override {
			c.W.Header().Del(key)
		}
		for _, val := range vals {
			c.W.Header().Add(key, val)
		}
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
