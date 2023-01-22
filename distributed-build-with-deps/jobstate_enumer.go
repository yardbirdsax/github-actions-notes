// Code generated by "enumer -type JobState -json"; DO NOT EDIT.

package distributed

import (
	"encoding/json"
	"fmt"
	"strings"
)

const _JobStateName = "NotStartedReadyQueuedStartedSucceededFailedDependencyFailedBadDependency"

var _JobStateIndex = [...]uint8{0, 10, 15, 21, 28, 37, 43, 59, 72}

const _JobStateLowerName = "notstartedreadyqueuedstartedsucceededfaileddependencyfailedbaddependency"

func (i JobState) String() string {
	if i < 0 || i >= JobState(len(_JobStateIndex)-1) {
		return fmt.Sprintf("JobState(%d)", i)
	}
	return _JobStateName[_JobStateIndex[i]:_JobStateIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _JobStateNoOp() {
	var x [1]struct{}
	_ = x[NotStarted-(0)]
	_ = x[Ready-(1)]
	_ = x[Queued-(2)]
	_ = x[Started-(3)]
	_ = x[Succeeded-(4)]
	_ = x[Failed-(5)]
	_ = x[DependencyFailed-(6)]
	_ = x[BadDependency-(7)]
}

var _JobStateValues = []JobState{NotStarted, Ready, Queued, Started, Succeeded, Failed, DependencyFailed, BadDependency}

var _JobStateNameToValueMap = map[string]JobState{
	_JobStateName[0:10]:       NotStarted,
	_JobStateLowerName[0:10]:  NotStarted,
	_JobStateName[10:15]:      Ready,
	_JobStateLowerName[10:15]: Ready,
	_JobStateName[15:21]:      Queued,
	_JobStateLowerName[15:21]: Queued,
	_JobStateName[21:28]:      Started,
	_JobStateLowerName[21:28]: Started,
	_JobStateName[28:37]:      Succeeded,
	_JobStateLowerName[28:37]: Succeeded,
	_JobStateName[37:43]:      Failed,
	_JobStateLowerName[37:43]: Failed,
	_JobStateName[43:59]:      DependencyFailed,
	_JobStateLowerName[43:59]: DependencyFailed,
	_JobStateName[59:72]:      BadDependency,
	_JobStateLowerName[59:72]: BadDependency,
}

var _JobStateNames = []string{
	_JobStateName[0:10],
	_JobStateName[10:15],
	_JobStateName[15:21],
	_JobStateName[21:28],
	_JobStateName[28:37],
	_JobStateName[37:43],
	_JobStateName[43:59],
	_JobStateName[59:72],
}

// JobStateString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func JobStateString(s string) (JobState, error) {
	if val, ok := _JobStateNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _JobStateNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to JobState values", s)
}

// JobStateValues returns all values of the enum
func JobStateValues() []JobState {
	return _JobStateValues
}

// JobStateStrings returns a slice of all String values of the enum
func JobStateStrings() []string {
	strs := make([]string, len(_JobStateNames))
	copy(strs, _JobStateNames)
	return strs
}

// IsAJobState returns "true" if the value is listed in the enum definition. "false" otherwise
func (i JobState) IsAJobState() bool {
	for _, v := range _JobStateValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for JobState
func (i JobState) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for JobState
func (i *JobState) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("JobState should be a string, got %s", data)
	}

	var err error
	*i, err = JobStateString(s)
	return err
}
