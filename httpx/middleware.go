package httpx

import (
	"context"

	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func AddComputeContentLength(stack *middleware.Stack) error {
	return stack.Build.Add(&smithyhttp.ComputeContentLength{}, middleware.After)
}

// rawResponseKey is the accessor key used to store and access the
// raw response within the response metadata.
type rawResponseKey struct{}

// AddRawResponse middleware adds raw response on to the metadata
type AddRawResponse struct{}

// ID the identifier for the ClientRequestID
func (m *AddRawResponse) ID() string {
	return "AddRawResponseToMetadata"
}

// HandleDeserialize adds raw response on the middleware metadata
func (m AddRawResponse) HandleDeserialize(ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler) (
	out middleware.DeserializeOutput, metadata middleware.Metadata, err error,
) {
	out, metadata, err = next.HandleDeserialize(ctx, in)
	metadata.Set(rawResponseKey{}, out.RawResponse)
	return out, metadata, err
}

// AddRawResponseToMetadata adds middleware to the middleware stack that
// store raw response on to the metadata.
func AddRawResponseToMetadata(stack *middleware.Stack) error {
	return stack.Deserialize.Add(&AddRawResponse{}, middleware.Before)
}

// GetRawResponse returns raw response set on metadata
func GetRawResponse(metadata middleware.Metadata) interface{} {
	return metadata.Get(rawResponseKey{})
}
