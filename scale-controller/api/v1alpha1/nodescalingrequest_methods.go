package v1alpha1

// IsFailed returns true if the value of .status.phase is either
// PhaseNodeCreateFailed or PhaseNodeDeleteFailed, otherwise return false.
func (nsr *NodeScalingRequest) IsFailed() bool {
	return nsr.Status.Phase == PhaseNodeCreateFailed ||
		nsr.Status.Phase == PhaseNodeDeleteFailed
}

// IsFinished returns true if the value of .status.phase is either
// PhaseNodeCreated, PhaseNodeDeleted or failure phases, otherwise return false.
func (nsr *NodeScalingRequest) IsFinished() bool {
	return nsr.Status.Phase == PhaseNodeCreated ||
		nsr.Status.Phase == PhaseNodeDeleted ||
		nsr.IsFailed()
}

// IsNotStarted returns true if the value of .status.phase is either
// PhaseNodeScaleInitial or PhaseNodeScalePending, otherwise return false.
func (nsr *NodeScalingRequest) IsNotStarted() bool {
	return nsr.Status.Phase == PhaseNodeScaleInitial ||
		nsr.Status.Phase == PhaseNodeScalePending
}

// IsProcessing returns true if the value of .status.phase is in processing phases,
// otherwise return false.
// Processing phases are all phases except not started and finished phases.
func (nsr *NodeScalingRequest) IsProcessing() bool {
	return !nsr.IsNotStarted() && !nsr.IsFinished()
}
