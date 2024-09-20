package httpx

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/aws/smithy-go/middleware"

	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func checkPathTemplate(pathTemplate string) bool {
	return len(pathTemplate) == 0 || pathTemplate[0] == '/'
}

// TODO...
type HaveEndpointOption interface {
	Get() *EndpointOptions
}

type EndpointOptions struct {
	method       string
	baseURL      *url.URL
	pathTemplate string
	path         string
}

type ResolveEndpoint struct {
	options *EndpointOptions
}

func (*ResolveEndpoint) ID() string {
	return "ResolveEndpoint"
}

func (m *ResolveEndpoint) HandleSerialize(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	if !checkPathTemplate(m.options.pathTemplate) {
		return out, metadata, fmt.Errorf("bad path template: %s", m.options.pathTemplate)
	}

	tmpl, err := template.New("ResolveEndpointPath").Parse(m.options.pathTemplate)
	if err != nil {
		return out, metadata, fmt.Errorf("bad path template: %s", m.options.pathTemplate)
	}

	var builder strings.Builder
	err = tmpl.Execute(&builder, in.Parameters)
	if err != nil {
		return out, metadata, fmt.Errorf("failed rander path template: %s", m.options.pathTemplate)
	}
	m.options.path = builder.String()

	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown transport type %T", in.Request)
	}

	req.Method = m.options.method

	m.options.baseURL.Path = m.options.path
	req.URL.Scheme = m.options.baseURL.Scheme
	req.URL.Host = m.options.baseURL.Host
	req.URL.Path = m.options.path

	return next.HandleSerialize(ctx, in)
}

func NewEndpointOptions(method, baseUrl, pathTemp string) (*EndpointOptions, error) {
	url, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	return &EndpointOptions{
		method:       method,
		baseURL:      url,
		pathTemplate: pathTemp,
	}, nil
}

func NewResolveEndpointMiddleware(options *EndpointOptions) *ResolveEndpoint {
	return &ResolveEndpoint{options: options}
}
