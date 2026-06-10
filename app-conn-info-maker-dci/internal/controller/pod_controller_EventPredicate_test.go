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

package controller

import (
	"bytes"
	"fmt"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func pod_controller_EventPredicate_Update_Test01(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test01: ObjectOld is nil
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test01: ObjectOld is nil")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	newObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: nil,
		ObjectNew: newObj,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "If ObjectOld is nil, it should return true.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test01")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Update_Test02(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test02: ObjectNew is nil
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test02: ObjectNew is nil")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	oldObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: oldObj,
		ObjectNew: nil,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "If ObjectNew is nil, it should return true.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test02")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Update_Test03(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test03: ObjectOld is not Pod type
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test03: ObjectOld is not Pod type")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	// Configure non-Pod-type objects
	oldObj := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
	}

	newObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: oldObj,
		ObjectNew: newObj,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "If ObjectOld is not of type Pod, it should return true.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test03")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Update_Test04(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test04: ObjectNew is not Pod type
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test04: ObjectNew is not Pod type")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	oldObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	// Configure non-Pod-type objects
	newObj := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: oldObj,
		ObjectNew: newObj,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "If ObjectNew is not of type Pod, it should return true.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test04")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Update_Test05(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test05: DeletionTimestamp is not zero
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test05: DeletionTimestamp is not zero")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	oldObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	deletionTime := metav1.Now()
	newObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pod",
			Namespace:         "default",
			DeletionTimestamp: &deletionTime,
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: oldObj,
		ObjectNew: newObj,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "If DeletionTimestamp is set, it should return true.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test05")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Update_Test06(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test06: Annotations are different
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test06: Annotations are different")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	oldObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				"key1": "value1",
			},
		},
	}

	newObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				"key1": "value2",
			},
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: oldObj,
		ObjectNew: newObj,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "If the annotations differ, it should return true.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test06")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Update_Test07(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test07: Pod Phase is different
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test07: Pod Phase is different")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	oldObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}

	newObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: oldObj,
		ObjectNew: newObj,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "If the Pod Phase is different, it should return true.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test07")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Update_Test08(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test08: Normal update (all conditions pass)
	//  Expected behavior: Returns false
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test08: Normal update (all conditions pass)")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	oldObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				"key1": "value1",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	newObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				"key1": "value1",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: oldObj,
		ObjectNew: newObj,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeFalse(), "In the case of a normal update (no changes), it should return false.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test08")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Create_Test01(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Create_Test01: Create event always returns true
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Create_Test01: Create event always returns true")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	obj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	createEvent := event.CreateEvent{
		Object: obj,
	}

	result := predicate.Create(createEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "The Create event should always return true.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Create_Test01")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Create_Test02(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Create_Test02: Create event with nil object
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Create_Test02: Create event with nil object")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	createEvent := event.CreateEvent{
		Object: nil,
	}

	result := predicate.Create(createEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "The Create event should always return true (even for nil objects).")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Create_Test02")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Create_Test03(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Create_Test03: Create event with DeletionTimestamp set
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Create_Test03: Create event with DeletionTimestamp set")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	deletionTime := metav1.Now()
	obj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pod",
			Namespace:         "default",
			DeletionTimestamp: &deletionTime,
		},
	}

	createEvent := event.CreateEvent{
		Object: obj,
	}

	result := predicate.Create(createEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "The Create event should always return true (even if DeletionTimestamp is set).")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Create_Test03")
	fmt.Println("<<<<<")
}

func pod_controller_EventPredicate_Create_Test04(writer bytes.Buffer, reconciler PodReconciler) {
	// ------------------------------
	//  EventPredicate_Create_Test04: Create event with non-Pod type
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> pod_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Create_Test04: Create event with non-Pod type")
	fmt.Println(">>>>>")

	predicate := PodEventPredicate{}

	// Objects that are not Pod-type
	obj := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
	}

	createEvent := event.CreateEvent{
		Object: obj,
	}

	result := predicate.Create(createEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "The Create event should always return true (regardless of type).")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Create_Test04")
	fmt.Println("<<<<<")
}
