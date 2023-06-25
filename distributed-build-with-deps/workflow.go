//go:generate go run github.com/golang/mock/mockgen -source=workflow.go -destination mock/workflow.go -package mock
package distributed

import (
	"github.com/yardbirdsax/go-ghworkflow"
)

type WorkflowRunner interface {
	Run(path string, inputs map[string]interface{}) *ghworkflow.WorkflowRun
}

// This is just a wrapper around github.com/yardbirdsax/go-ghworkflow's WorkflowRun
type WorkflowRun struct{}

func (w *WorkflowRun) Run(path string, inputs map[string]interface{}) *ghworkflow.WorkflowRun {
	return ghworkflow.Run(path, ghworkflow.WithInputs(inputs)).Wait()
}
