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
