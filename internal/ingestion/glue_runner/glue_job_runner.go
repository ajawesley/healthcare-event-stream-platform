// glue_job_runner.go
package glue_runner

import "context"

type GlueJobRunner interface {
	StartJob(ctx context.Context, input GlueJobInput) (string, error)
}

type GlueJobInput struct {
	InputPath      string
	OutputBasePath string
	ErrorPath      string
}
