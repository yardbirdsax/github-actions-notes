//go:generate go run github.com/dmarkham/enumer@v1.5.7 -type JobState -json
package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
)

type Action struct {
	Matrix         []*Job `json:"matrix"`
	ConcurrentJobs int    `json:"concurrent_jobs"`
	workflowRunner WorkflowRunner
	// Receives jobs that should be run
	runChannel chan (*Job)
	// Used for cancelling or signaling to end all goroutines
	ctx context.Context
}

func (a *Action) Run() error {
	a.runChannel = make(chan *Job, a.ConcurrentJobs)
	// Start goroutine to update jobs and send ready ones back to queue
	go func() {
		for {
			select {
			case <-a.ctx.Done():
				return
			case <-time.After(time.Second):
				a.UpdateJobs()
				a.SendJobsForRun()
			}
		}
	}()
	// Start goroutines to pull jobs off queue and run them
	for i := 1; i <= a.ConcurrentJobs; i++ {
		go func() {
			for {
				select {
				case job := <-a.runChannel:
					a.runJob(job)
				case <-a.ctx.Done():
					return
				}
			}
		}()
	}
	// Wait for completion
	for {
		select {
		case <-a.ctx.Done():
			break
		}
	}

	return nil
}

// UpdateJobs reconciles job states based on dependencies.
func (a *Action) UpdateJobs() {
	for i := range a.Matrix {
		job := a.Matrix[i]

		// This is in a separate 'func()' block so that the 'defer' statement
		// runs within the loop.
		func() {
			// We only care about updating jobs that haven't started, since
			// all other states are either final (like in the case of failed,
			// dependency failed, or succeeded) or managed elsewhere (like Ready
			// or Running). This is done outside the locked state to avoid
			// cases where jobs that are currently running (and therefore locked)
			// are checked, which would cause long delays.
			if job.State != NotStarted {
				return
			}

			job.mu.Lock()
			defer job.mu.Unlock()

			// This is an extra safety check in case the state was mutated between the initial
			// check and the lock.
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

// GetJobByState gets all jobs that have a particular State value.
func (a *Action) GetJobByState(state JobState) []*Job {
	results := []*Job{}

	for _, j := range a.Matrix {
		if j.State == state {
			results = append(results, j)
		}
	}

	return results
}

// SendJobsForRun sends ready jobs to the channel listened to by the
// worker routines.
func (a *Action) SendJobsForRun() {
	for _, j := range a.GetJobByState(Ready) {
		j.State = Queued
		a.runChannel <- j
	}
}

// GetJobCountByState provides a map of job states and the number of jobs
// currently in that state.
func (a *Action) GetJobCountByState() map[JobState]int {
	result := make(map[JobState]int)
	for _, j := range a.Matrix {
		result[j.State]++
	}
	return result
}

// error returns a MultiError of all job's Error properties
func (a *Action) error() error {
	err := &multierror.Error{}
	for _, j := range a.Matrix {
		if j.Error != nil {
			newError := fmt.Errorf("error running job with name %s: %w", j.Name, j.Error)
			err = multierror.Append(err, newError)
		}
	}
	return err.ErrorOrNil()
}

type JobState int

const (
	NotStarted JobState = iota
	Ready
	Queued
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
	Error             error
	mu                sync.Mutex
}
