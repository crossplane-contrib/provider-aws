package jobrun

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/aws/aws-sdk-go/service/emrcontainers/emrcontainersiface"
)

type MockClient struct {
	emrcontainersiface.EMRContainersAPI
	output output
}

type output struct {
	err error
	*emrcontainers.ListTagsForResourceOutput
	*emrcontainers.UntagResourceOutput
	*emrcontainers.TagResourceOutput
	*emrcontainers.DescribeJobRunOutput
}

func (m *MockClient) ListTagsForResourceWithContext(ctx aws.Context, input *emrcontainers.ListTagsForResourceInput, opts ...request.Option) (*emrcontainers.ListTagsForResourceOutput, error) {
	return m.output.ListTagsForResourceOutput, m.output.err
}

func (m *MockClient) UntagResourceWithContext(ctx aws.Context, input *emrcontainers.UntagResourceInput, opts ...request.Option) (*emrcontainers.UntagResourceOutput, error) {
	return m.output.UntagResourceOutput, m.output.err
}

func (m *MockClient) TagResourceWithContext(ctx aws.Context, input *emrcontainers.TagResourceInput, opts ...request.Option) (*emrcontainers.TagResourceOutput, error) {
	return m.output.TagResourceOutput, m.output.err
}

func (m *MockClient) DescribeJobRunWithContext(ctx aws.Context, input *emrcontainers.DescribeJobRunInput, opts ...request.Option) (*emrcontainers.DescribeJobRunOutput, error) {
	return m.output.DescribeJobRunOutput, m.output.err
}
