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

package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ProviderType is a type for provider type.
type ProviderType string

// Definition for environment variable.
const (
	ProviderTypeUnknown ProviderType = "Unknown"
	ProviderTypeDCI     ProviderType = "DCI"
	ProviderTypeDocker  ProviderType = "Docker"
)

// Available MachineTemplate string.
// This string is used to check kind of infrastructureRef.
var MachineTemplateNameMap = map[ProviderType]string{
	ProviderTypeDCI:    "DCIMachineTemplate",
	ProviderTypeDocker: "DockerMachineTemplate",
}

const (
	// Environment variable name for provider type.
	ProviderTypeEnvVarName = "PROVIDER_NAME"
)

// GetProviderTypeFromEnv gets provider type from environment variable "PROVIDER_NAME".
func GetProviderTypeFromEnv() (ProviderType, error) {
	pt := ProviderTypeDCI
	provider, isSet := os.LookupEnv(ProviderTypeEnvVarName)
	if isSet {
		if provider == "" {
			return ProviderTypeUnknown, fmt.Errorf("provider type is empty: available(%s,%s)", ProviderTypeDCI, ProviderTypeDocker)
		} else if strings.EqualFold(provider, string(ProviderTypeDCI)) {
			pt = ProviderTypeDCI
		} else if strings.EqualFold(provider, string(ProviderTypeDocker)) {
			pt = ProviderTypeDocker
		} else {
			return ProviderTypeUnknown, fmt.Errorf("provider type(%s) is invalid: available(%s,%s)",
				provider, ProviderTypeDCI, ProviderTypeDocker)
		}
	}
	return pt, nil
}

// GetTTLConfigFromEnv gets TTL configuration from environment variable.
//
//nolint:gocritic // ignore named returns
func GetTTLConfigFromEnv(envKey string) (bool, int, error) {
	ttlSecondsString, isSet := os.LookupEnv(envKey)
	if !isSet {
		return false, 0, nil
	} // else is set

	// Validate TTL seconds
	if ttlSecondsString == "" {
		return false, 0, fmt.Errorf("argument for setting TTL seconds is empty")
	}
	ttlSeconds, err := strconv.ParseInt(ttlSecondsString, 10, 32)
	if err != nil {
		return false, 0, fmt.Errorf("argument for setting TTL seconds cannot convert into int32: %w", err)
	}
	if ttlSeconds < 0 {
		return false, 0, fmt.Errorf("argument for setting TTL seconds is invalid(%d)", ttlSeconds)
	}
	return true, int(ttlSeconds), nil
}
