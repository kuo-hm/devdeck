package ui

type LogMsg struct {
	TaskIndex int
	Content   string
}

type ProcessFinishedMsg struct {
	TaskIndex int
	Err       error
}
