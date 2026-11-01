package resolver

import (
	"context"
	"fmt"
	"strings"

	graphql "github.com/99designs/gqlgen/graphql"
	gqlerror "github.com/vektah/gqlparser/v2/gqlerror"
)

var (
	ErrIdentityNotFound = fmt.Errorf("identity not found")

	errDryrun = fmt.Errorf("dry run")
)

type ErrDBConnection struct {
	err error
}

func (e ErrDBConnection) Error() string {
	return e.err.Error()
}

func IsDBError(e error) bool {
	_, ok := e.(*ErrDBConnection)
	return ok
}

func defaultGQLErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	if gqlerr, ok := err.(*gqlerror.Error); ok {
		if gqlerr.Path == nil {
			gqlerr.Path = graphql.GetFieldContext(ctx).Path()
		}
		return gqlerr
	}

	// it seems there is no way to obtain extensions from context directly.
	// We simply use a nil map.
	return &gqlerror.Error{
		Message: err.Error(),
		Path:    graphql.GetFieldContext(ctx).Path(),
	}
}

type ErrCode uint64

const (
	/**************** Client Error ****************/
	CodeInvalidRequestError ErrCode = 0x20000000 + iota
	CodeInsufficientPrivileges
	CodeObjectNotFound
	CodeObjectDuplicated
	CodeObjectIsBusy
	CodeQuotaExceeded
	CodeNoLongerSupport
)

const (
	/**************** Server Error ****************/
	CodeInternalServerError ErrCode = 0x40000000 + iota
	CodeInternalDatabaseError
)

const (
	/**************** External Error ****************/
	CodeGatewayError ErrCode = 0x60000000 + iota
)

func newGQLError(ctx context.Context, code ErrCode, err error) *gqlerror.Error {
	e := defaultGQLErrorPresenter(ctx, err)
	e.Extensions = map[string]interface{}{
		"code": "0x" + fmt.Sprintf("%08X", code),
	}
	return e
}

func newGQLErrorf(ctx context.Context, code ErrCode, msg string, args ...interface{}) *gqlerror.Error {
	e := defaultGQLErrorPresenter(ctx, fmt.Errorf(msg, args...))
	e.Extensions = map[string]interface{}{
		"code": "0x" + fmt.Sprintf("%08X", code),
	}
	return e
}

func IsGQLError(e error) bool {
	_, ok := e.(*gqlerror.Error)
	return ok
}

func isObjectNotFound(e error) bool {
	gqlerr, ok := e.(*gqlerror.Error)
	return ok && gqlerr.Extensions["code"] == fmt.Sprintf("0x%08X", CodeObjectNotFound)
}

func errInvalidRequest(ctx context.Context, fieldName, reason string, fieldValue ...interface{}) *gqlerror.Error {
	err := newGQLErrorf(ctx, CodeInvalidRequestError, "invalid request")
	if fieldName != "" {
		err.Extensions["variable"] = fieldName
	}
	if len(fieldValue) > 0 {
		err.Extensions["value"] = fieldValue
	}
	if reason != "" {
		err.Extensions["reason"] = reason
	}
	return err
}

func errInsufficientPrivileges(ctx context.Context) *gqlerror.Error {
	return newGQLErrorf(ctx, CodeInsufficientPrivileges, "insufficient privileges")
}

func errObjectNotFound(ctx context.Context, fieldName string, fieldValue ...interface{}) *gqlerror.Error {
	err := newGQLErrorf(ctx, CodeObjectNotFound, "object not found")
	if fieldName != "" {
		err.Extensions["variable"] = fieldName
	}
	if len(fieldValue) > 0 {
		err.Extensions["value"] = fieldValue
	}
	return err
}

func errObjectDuplicated(ctx context.Context, fieldName string, fieldValue ...interface{}) *gqlerror.Error {
	err := newGQLErrorf(ctx, CodeObjectDuplicated, "object duplicated")
	if fieldName != "" {
		err.Extensions["variable"] = fieldName
	}
	if len(fieldValue) > 0 {
		err.Extensions["value"] = fieldValue
	}
	return err
}

func errObjectIsBusy(ctx context.Context, fieldName, refererType string, refererID interface{}, refererName string, fieldValue ...interface{}) *gqlerror.Error {
	err := newGQLErrorf(ctx, CodeObjectIsBusy, "object is busy")
	if fieldName != "" {
		err.Extensions["variable"] = fieldName
	}
	if len(fieldValue) > 0 {
		err.Extensions["value"] = fieldValue
	}
	err.Extensions["reason"] = &objectIsBusy{
		Type: refererType,
		ID:   refererID,
		Name: refererName,
	}
	return err
}

func errQuotaExceeded(ctx context.Context, fieldName string, fieldValue ...interface{}) *gqlerror.Error {
	err := newGQLErrorf(ctx, CodeQuotaExceeded, "quota exceeded")
	if fieldName != "" {
		err.Extensions["variable"] = fieldName
	}
	if len(fieldValue) > 0 {
		err.Extensions["value"] = fieldValue
	}
	return err
}

func errNoLongerSupport(ctx context.Context, reason string) *gqlerror.Error {
	err := newGQLErrorf(ctx, CodeNoLongerSupport, "no longer support")
	if reason != "" {
		err.Extensions["reason"] = reason
	}
	return err
}

func errGatewayError(ctx context.Context, e error) *gqlerror.Error {
	code, message := gatewayErrorMessage(e)
	err := newGQLErrorf(ctx, code, "%s", message)
	err.Extensions["reason"] = message
	return err
}

func errClusterGatewayError(ctx context.Context, clusterName string, e error) *gqlerror.Error {
	err := errGatewayError(ctx, e)
	err.Extensions["cluster"] = clusterName
	return err
}

func gatewayErrorMessage(e error) (ErrCode, string) {
	reason := strings.TrimSpace(e.Error())
	if strings.Contains(reason, "the server could not find the requested resource") {
		switch {
		case strings.Contains(reason, "builds.kpack.io"):
			return CodeInvalidRequestError, "source builds are not enabled on this cluster"
		case strings.Contains(reason, ".velero.io"):
			return CodeInvalidRequestError, "backup and restore are not enabled on this cluster"
		}
	}
	switch {
	case strings.Contains(reason, "connection refused"):
		return CodeGatewayError, "gateway error: upstream connection refused"
	case strings.Contains(reason, "context deadline exceeded"), strings.Contains(reason, "i/o timeout"):
		return CodeGatewayError, "gateway error: upstream request timed out"
	case strings.Contains(strings.ToLower(reason), "unauthorized"), strings.Contains(reason, "provide credentials"):
		return CodeGatewayError, "gateway error: upstream authentication failed"
	case strings.Contains(strings.ToLower(reason), "forbidden"):
		return CodeGatewayError, "gateway error: upstream access denied"
	case strings.Contains(reason, "x509:"):
		return CodeGatewayError, "gateway error: upstream TLS certificate validation failed"
	case strings.Contains(reason, "the server could not find the requested resource"):
		return CodeGatewayError, "gateway error: requested Kubernetes resource is unavailable"
	}
	return CodeGatewayError, "gateway error: upstream request failed"
}

type objectIsBusy struct {
	Type string      `json:"type,omitempty"`
	ID   interface{} `json:"id,omitempty"`
	Name string      `json:"name,omitempty"`
}
