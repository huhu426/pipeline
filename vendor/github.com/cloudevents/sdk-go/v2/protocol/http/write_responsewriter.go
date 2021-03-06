package http

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/types"
)

// Write out to the the provided httpResponseWriter with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteResponseWriter(ctx context.Context, m binding.Message, status int, rw http.ResponseWriter, transformers ...binding.TransformerFactory) error {
	if status < 200 || status >= 600 {
		status = http.StatusOK
	}
	writer := &httpResponseWriter{rw: rw, status: status}

	_, err := binding.Write(
		ctx,
		m,
		writer,
		writer,
		transformers...,
	)
	return err
}

type httpResponseWriter struct {
	rw     http.ResponseWriter
	status int
}

func (b *httpResponseWriter) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	b.rw.Header().Set(ContentType, format.MediaType())
	return b.SetData(event)
}

func (b *httpResponseWriter) Start(ctx context.Context) error {
	return nil
}

func (b *httpResponseWriter) End(ctx context.Context) error {
	return nil
}

func (b *httpResponseWriter) SetData(reader io.Reader) error {
	// Finalize the headers.
	b.rw.WriteHeader(b.status)

	// Write body.
	copied, err := io.Copy(b.rw, reader)
	if err != nil {
		return err
	}
	b.rw.Header().Set("Content-Length", strconv.FormatInt(copied, 10))
	return nil
}

func (b *httpResponseWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}

	if attribute.Kind() == spec.DataContentType {
		b.rw.Header().Add(ContentType, s)
	} else {
		b.rw.Header().Add(prefix+attribute.Name(), s)
	}
	return nil
}

func (b *httpResponseWriter) SetExtension(name string, value interface{}) error {
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.rw.Header().Add(prefix+name, s)
	return nil
}

var _ binding.StructuredWriter = (*httpResponseWriter)(nil) // Test it conforms to the interface
var _ binding.BinaryWriter = (*httpResponseWriter)(nil)     // Test it conforms to the interface
