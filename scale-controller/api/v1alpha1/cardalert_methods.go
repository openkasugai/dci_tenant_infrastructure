package v1alpha1

// IsFailed returns true if the value of .status.state is Failed,
// otherwise return false.
func (csr *CardAlert) IsFailed() bool {
	return csr.Status.State == CardAlertFailedState
}

// IsFinished returns true if the value of .status.state is either
// Completed or failure states, otherwise return false.
func (csr *CardAlert) IsFinished() bool {
	return csr.Status.State == CardAlertCompletedState ||
		csr.IsFailed()
}

// IsFinalized returns true if the value of .status.state is Finalized,
// otherwise return false.
func (csr *CardAlert) IsFinalized() bool {
	return csr.Status.State == CardAlertFinalizedState
}

// IsNotStarted returns true if the value of .status.state is either
// Initial or Pending, otherwise return false.
func (csr *CardAlert) IsNotStarted() bool {
	return csr.Status.State == CardAlertInitialState ||
		csr.Status.State == CardAlertPendingState
}

// IsProcessing returns true if the value of .status.state is in processing states,
// otherwise return false.
// Processing states are all states except not started, finished, and finalized states.
func (csr *CardAlert) IsProcessing() bool {
	return !csr.IsNotStarted() && !csr.IsFinished() && !csr.IsFinalized()
}
