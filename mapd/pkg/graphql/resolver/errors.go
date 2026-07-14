package resolver

import (
	"context"
	"fmt"

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
	err := newGQLErrorf(ctx, CodeGatewayError, "gateway error")
	err.Extensions["reason"] = e.Error()
	return err
}

type objectIsBusy struct {
	Type string      `json:"type,omitempty"`
	ID   interface{} `json:"id,omitempty"`
	Name string      `json:"name,omitempty"`
}
