package custom_errors

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCustomErrorFromError(t *testing.T) {
	t.Parallel()
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want *CustomError
	}{
		{
			name: "is CustomError",
			args: args{
				err: NewNotFoundError(nil),
			},
			want: NewNotFoundError(nil),
		},
		{
			name: "is not CustomError",
			args: args{
				err: errors.New("some error"),
			},
			want: nil,
		},
		{
			name: "wrapped custom error",
			args: args{
				err: fmt.Errorf("some error: %w", NewNotFoundError(nil)),
			},
			want: NewNotFoundError(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CustomErrorFromError(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CustomErrorFromError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_AddDetails(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	type args struct {
		detail []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CustomError
	}{
		{
			name: "",
			fields: fields{
				code: codes.NotFound,
			},
			args: args{
				detail: []string{"smth went wrong"},
			},
			want: &CustomError{
				code:    codes.NotFound,
				details: []string{"smth went wrong"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.AddDetails(tt.args.detail...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_Copy(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	tests := []struct {
		name   string
		fields fields
		want   *CustomError
	}{
		{
			name: "copy",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			want: &CustomError{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.Copy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_Error(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			want: "error code = NotFound; details: = some detail; errTypeCode: = smth code",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_FormatMsg(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			want: `code = NotFound; base error: = some error; details: = some detail; errTypeCode: = smth code; description: = smth went wrong, resolve via contact centre`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.FormatMsg(); got != tt.want {
				t.Errorf("FormatMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_GRPCStatus(t *testing.T) {
	t.Parallel()
	t.Run("gRPCStatus", func(t *testing.T) {
		t.Parallel()
		c := &CustomError{
			errType:     "smth code",
			code:        codes.NotFound,
			base:        errors.New("some error"),
			description: "smth went wrong, resolve via contact centre",
			details:     []string{"some detail"},
		}

		grpcErr, _ := status.New(codes.NotFound, codes.NotFound.String()).WithDetails(
			&epb.ErrorInfo{
				Reason: "smth code",
				//Domain: domain,
				Metadata: map[string]string{
					metaDataDescriptionKey: "smth went wrong, resolve via contact centre",
				}})

		assert.Equal(t, grpcErr, c.GRPCStatus())
	})
}

func TestCustomError_GetDescription(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			want: "smth went wrong, resolve via contact centre",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.GetDescription(); got != tt.want {
				t.Errorf("GetDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_GetDetails(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			want: []string{"some detail"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.GetDetails(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_GetErrType(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			want: "smth code",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.GetErrType(); got != tt.want {
				t.Errorf("GetErrType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_Is(t *testing.T) {
	t.Parallel()
	smthError := errors.New("some error")
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "not is",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			args: args{
				err: errors.New("some error"),
			},
			want: false,
		},
		{
			name: "is",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        smthError,
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			args: args{
				err: smthError,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.Is(tt.args.err); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_SetDescription(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	type args struct {
		description string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CustomError
	}{
		{
			name: "",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			args: args{
				description: "new desc",
			},
			want: &CustomError{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "new desc",
				details:     []string{"some detail"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.SetDescription(tt.args.description); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_SetErrType(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	type args struct {
		errType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CustomError
	}{
		{
			name: "",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			args: args{
				errType: "new smth code",
			},
			want: &CustomError{
				errType:     "new smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.SetErrType(tt.args.errType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetErrType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_TypeCodeIs(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	type args struct {
		typeCode string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "is",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			args: args{
				typeCode: "smth code",
			},
			want: true,
		},
		{
			name: "not is",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			args: args{
				typeCode: "other smth code",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if got := c.TypeCodeIs(tt.args.typeCode); got != tt.want {
				t.Errorf("TypeCodeIs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_Unwrap(t *testing.T) {
	t.Parallel()
	type fields struct {
		errType     string
		code        codes.Code
		base        error
		description string
		details     []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				errType:     "smth code",
				code:        codes.NotFound,
				base:        errors.New("some error"),
				description: "smth went wrong, resolve via contact centre",
				details:     []string{"some detail"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CustomError{
				errType:     tt.fields.errType,
				code:        tt.fields.code,
				base:        tt.fields.base,
				description: tt.fields.description,
				details:     tt.fields.details,
			}
			if err := c.Unwrap(); (err != nil) != tt.wantErr {
				t.Errorf("Unwrap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewAbortedError(t *testing.T) {
	t.Parallel()
	_ = NewForbiddenError(errors.New("some error"))
	_ = NewNotFoundError(errors.New("some error"))
	_ = NewUnauthorizedError(errors.New("some error"))
	_ = NewInternalError(errors.New("some error"))
	_ = NewBadRequestError(errors.New("some error"))
}
