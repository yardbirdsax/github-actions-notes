package distributed

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/yardbirdsax/github-actions-notes/distributed-build-with-deps/mock"
	"github.com/yardbirdsax/go-ghworkflow"
)

func TestUpdateJobs(t *testing.T) {
	testCases := []struct {
		Name            string
		Jobs            []*Job
		ExpectedResults []*Job
	}{
		{
			Name: "AllDependentsSucceeded",
			Jobs: []*Job{
				{
					Name:  "FirstJob",
					State: Succeeded,
				},
				{
					Name:  "DependentJob",
					State: NotStarted,
					DependentJobNames: []string{
						"FirstJob",
					},
				},
				{
					Name:  "UnrelatedJob",
					State: NotStarted,
				},
			},
			ExpectedResults: []*Job{
				{
					Name:  "FirstJob",
					State: Succeeded,
				},
				{
					Name:  "DependentJob",
					State: Ready,
					DependentJobNames: []string{
						"FirstJob",
					},
				},
				{
					Name:  "UnrelatedJob",
					State: Ready,
				},
			},
		},
		{
			Name: "FailedDependency",
			Jobs: []*Job{
				{
					Name:  "FirstJob",
					State: Failed,
				},
				{
					Name:              "DependentJob",
					State:             NotStarted,
					DependentJobNames: []string{"FirstJob"},
				},
				{
					Name:              "SecondDependentJob",
					State:             NotStarted,
					DependentJobNames: []string{"DependentJob"},
				},
			},
			ExpectedResults: []*Job{
				{
					Name:  "FirstJob",
					State: Failed,
				},
				{
					Name:  "DependentJob",
					State: DependencyFailed,
					DependentJobNames: []string{
						"FirstJob",
					},
				},
				{
					Name:              "SecondDependentJob",
					State:             DependencyFailed,
					DependentJobNames: []string{"DependentJob"},
				},
			},
		},
		{
			Name: "BadDependency",
			Jobs: []*Job{
				{
					Name:              "BadDependency",
					State:             NotStarted,
					DependentJobNames: []string{"NonExistentJob"},
				},
			},
			ExpectedResults: []*Job{
				{
					Name:              "BadDependency",
					State:             BadDependency,
					DependentJobNames: []string{"NonExistentJob"},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a := Action{
				Matrix: tc.Jobs,
			}

			a.UpdateJobs()

			assert.EqualValues(t, tc.ExpectedResults, a.Matrix)
		})
	}
}

func TestRunJob(t *testing.T) {
	testCases := []struct {
		Name              string
		InputJob          *Job
		ReturnWorkflowRun *ghworkflow.WorkflowRun
		ExpectedResult    *Job
	}{
		{
			Name: "Successful Execution",
			InputJob: &Job{
				Name:         "Job",
				State:        NotStarted,
				WorkflowName: "workflow",
			},
			ReturnWorkflowRun: &ghworkflow.WorkflowRun{
				Error: nil,
			},
			ExpectedResult: &Job{
				Name:         "Job",
				State:        Succeeded,
				WorkflowName: "workflow",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockWorkflow := mock.NewMockWorkflowRunner(ctrl)
			mockWorkflow.EXPECT().Run(tc.InputJob.WorkflowName, tc.InputJob.Inputs).Times(1).Return(tc.ReturnWorkflowRun)
			a := &Action{
				workflowRunner: mockWorkflow,
			}

			a.runJob(tc.InputJob)

		})
	}
}

func TestGetByState(t *testing.T) {
	a := &Action{
		Matrix: []*Job{
			{
				Name:  "Ready",
				State: Ready,
			},
			{
				Name:  "Failed",
				State: Failed,
			},
		},
	}
	expected := []*Job{
		{
			Name:  "Ready",
			State: Ready,
		},
	}

	actual := a.GetJobByState(Ready)

	assert.EqualValues(t, expected, actual)
}

func TestSendJobsForRun(t *testing.T) {
	readyJob := &Job{
		Name:  "Ready",
		State: Ready,
	}
	queuedJob := &Job{
		Name:  "Queued",
		State: Queued,
	}
	a := &Action{
		Matrix: []*Job{
			readyJob,
			queuedJob,
		},
		runChannel: make(chan *Job, 2),
	}
	expected := &Job{
		Name:  "Ready",
		State: Queued,
	}

	a.SendJobsForRun()

	var actual *Job

	select {
	case actual = <-a.runChannel:
		assert.EqualValues(t, expected, actual, "expected value was not sent on channel")
	case <-time.After(5 * time.Second):
		t.Log("expected value not received in a timely manner")
		t.Fail()
	}
}

func TestGetJobCountByState(t *testing.T) {
	a := &Action{
		Matrix: []*Job{
			{
				Name:  "Ready",
				State: Ready,
			},
			{
				Name:  "Queued",
				State: Queued,
			},
			{
				Name:  "NotStarted",
				State: NotStarted,
			},
		},
	}
	expectedResults := map[JobState]int{
		NotStarted: 1,
		Ready:      1,
		Queued:     1,
	}

	actualResults := a.GetJobCountByState()

	assert.Equal(t, expectedResults, actualResults)
}

func TestErrors(t *testing.T) {
	testCases := []struct {
		Name          string
		Matrix        []*Job
		ExpectedError *multierror.Error
	}{
		{
			Name: "WithErrors",
			ExpectedError: &multierror.Error{
				Errors: []error{
					fmt.Errorf("error running job with name JobWithError: %w", errors.New("an error")),
					fmt.Errorf("error running job with name AnotherJobWithError: %w", errors.New("another error")),
				},
			},
			Matrix: []*Job{
				{
					Name:  "JobWithError",
					Error: errors.New("an error"),
				},
				{
					Name:  "AnotherJobWithError",
					Error: errors.New("another error"),
				},
			},
		},
		{
			Name:          "NoErrors",
			ExpectedError: nil,
			Matrix: []*Job{
				{
					Name:  "JobWithError",
					Error: nil,
				},
				{
					Name:  "AnotherJobWithError",
					Error: nil,
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := &Action{
				Matrix: tc.Matrix,
			}

			actualError := a.error()

			if tc.ExpectedError != nil {
				assert.Equal(t, tc.ExpectedError, actualError, "actual and expected error are different")
			} else {
				assert.Nil(t, actualError, "actual error was not expected nil")
			}
		})
	}

}
