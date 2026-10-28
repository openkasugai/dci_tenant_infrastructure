package controller

import (
	"bytes"
	"fmt"

	. "github.com/onsi/gomega"

	examplecomv1 "app-conn-info-maker-dci/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func dciappconnection_controller_EventPredicate_Update_Test01(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test01: ObjectOld is nil
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test01: ObjectOld is nil")
	fmt.Println(">>>>>")

	predicate := DCIAppConnectionEventPredicate{}

	newObj := &examplecomv1.DCIAppConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dciappconnection",
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

func dciappconnection_controller_EventPredicate_Update_Test02(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test02: ObjectNew is nil
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test02: ObjectNew is nil")
	fmt.Println(">>>>>")

	predicate := DCIAppConnectionEventPredicate{}

	oldObj := &examplecomv1.DCIAppConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dciappconnection",
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

func dciappconnection_controller_EventPredicate_Update_Test03(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test03: ObjectNew is not DCIAppConnection type
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test03: ObjectNew is not DCIAppConnection type")
	fmt.Println(">>>>>")

	predicate := DCIAppConnectionEventPredicate{}

	oldObj := &examplecomv1.DCIAppConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dciappconnection",
			Namespace: "default",
		},
	}

	// Configuring an object that is not of type DCIAppConnection
	newObj := &examplecomv1.ConnectEndpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-connectendpoint",
			Namespace: "default",
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: oldObj,
		ObjectNew: newObj,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeTrue(), "If ObjectNew is not of type DCIAppConnection, it should return true.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test03")
	fmt.Println("<<<<<")
}

func dciappconnection_controller_EventPredicate_Update_Test04(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test04: DeletionTimestamp is not zero
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test04: DeletionTimestamp is not zero")
	fmt.Println(">>>>>")

	predicate := DCIAppConnectionEventPredicate{}

	oldObj := &examplecomv1.DCIAppConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dciappconnection",
			Namespace: "default",
		},
	}

	deletionTime := metav1.Now()
	newObj := &examplecomv1.DCIAppConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-dciappconnection",
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
	fmt.Println("<<<<< End EventPredicate_Update_Test04")
	fmt.Println("<<<<<")
}

func dciappconnection_controller_EventPredicate_Update_Test05(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) {
	// ------------------------------
	//  EventPredicate_Update_Test05: Normal update (all conditions pass)
	//  Expected behavior: Returns false
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Update_Test05: Normal update (all conditions pass)")
	fmt.Println(">>>>>")

	predicate := DCIAppConnectionEventPredicate{}

	oldObj := &examplecomv1.DCIAppConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dciappconnection",
			Namespace: "default",
		},
	}

	newObj := &examplecomv1.DCIAppConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dciappconnection",
			Namespace: "default",
		},
	}

	updateEvent := event.UpdateEvent{
		ObjectOld: oldObj,
		ObjectNew: newObj,
	}

	result := predicate.Update(updateEvent)

	fmt.Println(">>>>> Result:", result)
	Expect(result).To(BeFalse(), "In the case of a normal update, it should return false.")

	fmt.Println("<<<<<")
	fmt.Println("<<<<< End EventPredicate_Update_Test05")
	fmt.Println("<<<<<")
}

func dciappconnection_controller_EventPredicate_Create_Test01(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) {
	// ------------------------------
	//  EventPredicate_Create_Test01: Create event always returns true
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Create_Test01: Create event always returns true")
	fmt.Println(">>>>>")

	predicate := DCIAppConnectionEventPredicate{}

	obj := &examplecomv1.DCIAppConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dciappconnection",
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

func dciappconnection_controller_EventPredicate_Create_Test02(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) {
	// ------------------------------
	//  EventPredicate_Create_Test02: Create event with nil object
	//  Expected behavior: Returns true
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Create_Test02: Create event with nil object")
	fmt.Println(">>>>>")

	predicate := DCIAppConnectionEventPredicate{}

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

func dciappconnection_controller_EventPredicate_Create_Test03(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) {
	// ------------------------------
	//  EventPredicate_Create_Test03: Create event with DeletionTimestamp set
	//  Expected behavior: Returns true (Create events always return true)
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Create_Test03: Create event with DeletionTimestamp set")
	fmt.Println(">>>>>")

	predicate := DCIAppConnectionEventPredicate{}

	deletionTime := metav1.Now()
	obj := &examplecomv1.DCIAppConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-dciappconnection",
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

func dciappconnection_controller_EventPredicate_Create_Test04(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) {
	// ------------------------------
	//  EventPredicate_Create_Test04: Create event with non-DCIAppConnection type
	//  Expected behavior: Returns true (The Create event is always true regardless of type)
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_EventPredicate_test.go")
	fmt.Println(">>>>> Start EventPredicate_Create_Test04: Create event with non-DCIAppConnection type")
	fmt.Println(">>>>>")

	predicate := DCIAppConnectionEventPredicate{}

	// Object not of type DCIAppConnection
	obj := &examplecomv1.ConnectEndpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-connectendpoint",
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
