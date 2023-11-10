/*
Copyright 2023 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package errors

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/smithy-go"
	"github.com/pkg/errors"
)

// Wrap will remove the request-specific information from the error and only then
// wrap it.
func Wrap(err error, msg string) error {
	// NOTE(muvaf): nil check is done for performance, otherwise errors.As makes
	// a few reflection calls before returning false, letting awsErr be nil.
	if err == nil {
		return nil
	}
	var awsErr smithy.APIError
	if errors.As(err, &awsErr) {
		return errors.Wrap(awsErr, msg)
	}
	// AWS SDK v1 uses different interfaces than v2 and it doesn't unwrap to
	// the underlying error. So, we need to strip off the unique request ID
	// manually.
	if v1RequestError, ok := err.(awserr.RequestFailure); ok { //nolint:errorlint
		// TODO(negz): This loses context about the underlying error
		// type, preventing us from using errors.As to figure out what
		// kind of error it is. Could we do this without losing
		// context?
		return errors.Wrap(errors.New(strings.ReplaceAll(err.Error(), v1RequestError.RequestID(), "")), msg)
	}
	return errors.Wrap(err, msg)
}

// Combine returns a new error where the message is a comma separated list of
// all given error messages.
func Combine(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	errStrings := make([]string, 0, len(errs))
	for _, e := range errs {
		errStrings = append(errStrings, e.Error())
	}
	return errors.New(strings.Join(errStrings, ", "))
}
