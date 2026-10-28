package v1alpha1

// IsFailed returns true if the value of .status.phase is either NodeScaleUpFailed or NodeScaleDownFailed,
// otherwise return false.
func (csr *CardScalingRequest) IsFailed() bool {
	return csr.Status.Phase == PhaseNodeScaleUpFailed ||
		csr.Status.Phase == PhaseNodeScaleDownFailed
}

// IsFinished returns true if the value of .status.phase is either
// PhaseCardScaleCompleted or failure phases, otherwise return false.
func (csr *CardScalingRequest) IsFinished() bool {
	return csr.Status.Phase == PhaseCardScaleCompleted ||
		csr.IsFailed()
}

// IsNotStarted returns true if the value of .status.phase is either
// PhaseCardScaleInitial or PhaseCardScalePending, otherwise return false.
func (csr *CardScalingRequest) IsNotStarted() bool {
	return csr.Status.Phase == PhaseCardScaleInitial ||
		csr.Status.Phase == PhaseCardScalePending
}

// IsProcessing returns true if the value of .status.phase is in processing phases,
// otherwise return false.
// Processing phases are all phases except not started and finished phases.
func (csr *CardScalingRequest) IsProcessing() bool {
	return !csr.IsNotStarted() && !csr.IsFinished()
}
