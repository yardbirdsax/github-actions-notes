//go:generate go run github.com/dmarkham/enumer@v1.5.7 -type JobState -json
package distributed

import "sync"

type Action struct {
	Matrix         []*Job `json:"matrix"`
	workflowRunner WorkflowRunner
}

func (a *Action) Run() error {
  // Start goroutine to update jobs and send ready ones back to queue
  // Start goroutines to pull jobs off queue and run them
  // Wait for completion

	return nil
}

func (a *Action) UpdateJobs() {
  for i := range a.Matrix {
		job := a.Matrix[i]

    func() {
      job.mu.Lock()
      defer job.mu.Unlock()

      // We only care about updating jobs that haven't started, since
      // all other states are either final (like in the case of failed,
      // dependency failed, or succeeded) or managed elsewhere (like Ready
      // or Running).
      if job.State != NotStarted {
        return
      }

      for _, depJobName := range job.DependentJobNames {
        depJob := a.getJobByName(depJobName)

        // If we can't find a dependent job, we need to mark
        // the job as having a bad dependency and skip further processing.
        if depJob == nil {
          job.State = BadDependency
          return
        }

        switch depJob.State {
        // If successful, we don't do anything. If we get all the way through the loop
        // then we'll set the job to a ready state.
        case Succeeded:
          continue
        // If either the dependent job failed, or a dependency of the dependent job
        // failed, then we can safely set this one to DependencyFailed and move on
        // to the next available job.
        case Failed:
          job.State = DependencyFailed
          return
        case DependencyFailed:
          job.State = DependencyFailed
          return
        // For the remainder of states, we don't need to do anything, but we know
        // the job can't be marked as ready, so we can skip it.
        default:
          return
        }
      }
      // If we got this far, then we can set the job state to Ready.
      job.State = Ready
    }()
	}
}

func (a *Action) getJobByName(name string) *Job {
  for _, j := range a.Matrix {
    if j.Name == name {
      return j
    }
  }
  return nil
}

func (a *Action) runJob(job *Job) {
	job.mu.Lock()
  defer job.mu.Unlock()
  r := a.workflowRunner.Run(job.WorkflowName, job.Inputs)
	switch {
	case r.Error != nil:
		job.State = Failed
	case r.Error == nil:
		job.State = Succeeded
	}
}

type JobState int

const (
	NotStarted JobState = iota
	Ready
	Started
	Succeeded
	Failed
	DependencyFailed
  BadDependency
)

type Job struct {
	Name              string   `json:"name"`
	DependentJobNames []string `json:"dependentJobNames,omitempty"`
	State             JobState `json:"state"`
	WorkflowName      string   `json:"workflow_name"`
	Inputs            map[string]interface{}
  mu                sync.Mutex
}
