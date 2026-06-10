// Copyright 2026 NTT, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
