/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package phase

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	kmapi "kmodules.xyz/client-go/api/v1"
)

func PhaseFromCondition(conditions []kmapi.Condition) api.DatabasePhase {
	// Generally, the conditions should maintain the following chronological order
	// For normal restore process:
	//   ProvisioningStarted --> ReplicaReady --> AcceptingConnection --> DataRestoreStarted --> DataRestored --> Ready --> Provisioned
	// For restoring the volumes (PerconaXtraDB):
	//	 ProvisioningStarted --> DataRestoreStarted --> DataRestored --> ReplicaReady --> AcceptingConnection --> Ready --> Provisioned

	// These are transitional conditions. They can update any time. So, their order may vary:
	// 1. ReplicaReady
	// 2. AcceptingConnection
	// 3. DataRestoreStarted
	// 4. DataRestored
	// 5. Ready
	// 6. Paused

	var phase api.DatabasePhase

	// ==================================  Handling "ProvisioningStarted" condition  ========================
	// If the condition is present and its "true", then the phase should be "Provisioning".
	if kmapi.IsConditionTrue(conditions, api.DatabaseProvisioningStarted) {
		phase = api.DatabasePhaseProvisioning
	}

	// ================================== Handling "Halted" condition =======================================
	// The "Halted" condition has higher priority, that's why it is placed at the top.
	// If the condition is present and its "true", then the phase should be "Halted".
	if kmapi.IsConditionTrue(conditions, api.DatabaseHalted) {
		return api.DatabasePhaseHalted
	}

	// =================================== Handling "DataRestoreStarted" and "DataRestored" conditions  ==================================================
	// For data restoring, there could be the following scenarios:
	// 1. Data cond["DataRestoreStarted"] = nil and cond["DataRestored"] = nil. In this case, phase will depend on the other conditions.
	// 2. Data cond["DataRestoreStarted"] = true but cond["DataRestored"] = nil. In this case, phase should be "DataRestoring".
	// 3. Data cond["DataRestoreStarted"] = true but cond["DataRestored"] = false. In this case, there could be following scenarios:
	// 		a. Data cond["DataRestoreStarted"].LastTransitionTime > cond["DataRestored"].LastTransitionTime. In this case phase should be "DataRestoring".
	// 		b. Data cond["DataRestoreStarted"].LastTransitionTime <= cond["DataRestored"].LastTransitionTime. In this case phase should be "NotReady".
	// 4. Data cond["DataRestoreStarted"] = true but cond["DataRestored"] = true. In this case, there could be the following scenarios:
	// 		a. Data cond["DataRestoreStarted"].LastTransitionTime > cond["DataRestored"].LastTransitionTime. In this case phase should be "DataRestoring".
	// 		b. Data cond["DataRestoreStarted"].LastTransitionTime <= cond["DataRestored"].LastTransitionTime. In this case, phase will depend on the other conditions.
	if kmapi.HasCondition(conditions, api.DatabaseDataRestoreStarted) {
		// scenario 2
		if kmapi.IsConditionTrue(conditions, api.DatabaseDataRestoreStarted) && !kmapi.HasCondition(conditions, api.DatabaseDataRestored) {
			// When the database is restoring, no other conditions does matter. So, just return the phase.
			return api.DatabasePhaseDataRestoring
		}
		// scenario 3
		if kmapi.IsConditionTrue(conditions, api.DatabaseDataRestoreStarted) && !kmapi.IsConditionTrue(conditions, api.DatabaseDataRestored) {
			res := compareLastTransactionTime(conditions, api.DatabaseDataRestoreStarted, api.DatabaseDataRestored)
			if res <= 0 {
				// Database has failed to restore. User should not connect to the database. So, the database phase should be "NotReady"
				return api.DatabasePhaseNotReady
			}
			return api.DatabasePhaseDataRestoring
		}
		// scenario 4
		if kmapi.IsConditionTrue(conditions, api.DatabaseDataRestoreStarted) && kmapi.IsConditionTrue(conditions, api.DatabaseDataRestored) {
			res := compareLastTransactionTime(conditions, api.DatabaseDataRestoreStarted, api.DatabaseDataRestored)
			if res == 1 {
				return api.DatabasePhaseDataRestoring
			}
		}
	}

	// ================================= Handling "AcceptingConnection" condition ==========================
	// If the condition is present and its "false", then the phase should be "NotReady".
	// Skip if the database isn't provisioned yet.
	if kmapi.IsConditionFalse(conditions, api.DatabaseAcceptingConnection) && kmapi.IsConditionTrue(conditions, api.DatabaseProvisioned) {
		return api.DatabasePhaseNotReady
	}

	// ================================= Handling "ReplicaReady" condition ==========================
	// If the condition is present and its "false", then the phase should be "Critical".
	// Skip if the database isn't provisioned yet.
	if kmapi.IsConditionFalse(conditions, api.DatabaseReplicaReady) && kmapi.IsConditionTrue(conditions, api.DatabaseProvisioned) {
		return api.DatabasePhaseCritical
	}

	// ================================= Handling "Ready" condition ==========================
	// Skip if the database isn't provisioned yet.
	if kmapi.IsConditionFalse(conditions, api.DatabaseReady) && kmapi.IsConditionTrue(conditions, api.DatabaseProvisioned) {
		return api.DatabasePhaseCritical
	}
	// Ready, if the database is provisioned and readinessProbe passed.
	if kmapi.IsConditionTrue(conditions, api.DatabaseReady) && kmapi.IsConditionTrue(conditions, api.DatabaseProvisioned) {
		return api.DatabasePhaseReady
	}

	// ================================= Handling "Provisioned" and "Paused" conditions ==========================
	// These conditions does not have any effect on the database phase. They are only for internal usage.
	// So, we don't have to do anything for them.
	return phase
}

// compareLastTransactionTime compare two condition's "LastTransactionTime" and return an integer based on the followings:
// 1. If both conditions does not exist, then return 0
// 2. If cond1 exist but cond2 does not, then return 1
// 3. If cond1 does not exist but cond2 exist, then return -1
// 3. If cond1.LastTransactionTime > cond2.LastTransactionTime, then return 1
// 4. If cond1.LastTransactionTime = cond2.LastTransactionTime, then return 0
// 5. If cond1.LastTransactionTime < cond2.LastTransactionTime, then return -1
func compareLastTransactionTime(conditions []kmapi.Condition, type1, type2 string) int32 {
	idx1, cond1 := kmapi.GetCondition(conditions, type1)
	idx2, cond2 := kmapi.GetCondition(conditions, type2)
	// both condition does not exist
	if idx1 == -1 && idx2 == -1 {
		return 0
	}
	// cond1 exist but cond2 does not
	if idx1 != -1 && idx2 == -1 {
		return 1
	}
	// cond2 does not exist but cond2 exist
	if idx1 == -1 && idx2 != -1 {
		return -1
	}

	if cond1.LastTransitionTime.After(cond2.LastTransitionTime.Time) {
		// cond1 is newer than cond2
		return 1
	} else if cond2.LastTransitionTime.After(cond1.LastTransitionTime.Time) {
		// cond1 is older than cond2
		return -1
	}
	return 0
}
