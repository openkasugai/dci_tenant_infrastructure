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
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	examplecomv1 "app-conn-info-maker-dci/api/v1"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func dciappconnection_controller_Test22_1(writer bytes.Buffer, reconciler DCIAppConnectionReconciler) { //nolint:gocyclo
	var err error
	ctx := context.Background()
	// ------------------------------
	//  Test22_1 Create
	//  X-A,Y-C X-Y
	//  Y-None
	//  In-X,Out-Y
	// ------------------------------
	fmt.Println(">>>>>")
	fmt.Println(">>>>> dciappconnection_controller_test.go")
	fmt.Println(">>>>> Start Test22_1")
	fmt.Println(">>>>>")

	// Create Deployment
	// NotFound Test Target
	// stDeploymentCR := Deployment01FromX

	// Create Deployment
	stDeploymentCR := Deployment01ToA
	err = k8sClient.Create(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	stDeploymentCR = Deployment01ToC
	err = k8sClient.Create(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Deployment
	stDeploymentCR = Deployment01ToY
	err = k8sClient.Create(ctx, &stDeploymentCR)
	if err != nil {
		fmt.Println("There is a problem in createing Deployment01ToY")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR := ReplicaSet01FromX
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01FromX")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToA
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToA")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToC
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToC")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create ReplicaSet
	stReplicaSetCR = ReplicaSet01ToY
	err = k8sClient.Create(ctx, &stReplicaSetCR)
	if err != nil {
		fmt.Println("There is a problem in createing ReplicaSet01ToY")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create HostDeviceNetwork
	stHostDeviceNetworkCR := HostDeviceNetwork01
	err = k8sClient.Create(ctx, &stHostDeviceNetworkCR)
	if err != nil {
		fmt.Println("There is a problem in createing HostDeviceNetwork01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR := Pod01FromX1
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01FromX1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR FromX1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToA1
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToA2
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToA2")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToA2.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToC1
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToC1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToC1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToC1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToC1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToC1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create Pod
	stPodCR = Pod01ToY1
	err = k8sClient.Create(ctx, &stPodCR)
	if err != nil {
		fmt.Println("There is a problem in createing Pod01ToY1")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToY1.")
	}
	stPodCR.Status.Phase = PodStatusRunning.Phase
	err = k8sClient.Status().Update(ctx, &stPodCR)
	if err != nil {
		fmt.Println("Failed to update PodCR ToY1.")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// Create DCIAppConnectionCR
	stDCIAppConnectionCR := DCIAppConnection01
	err = k8sClient.Create(ctx, &stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("There is a problem in createing DCIAppConnection01")
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	// ------------------------------
	//  Test22_1 Reconcile
	// ------------------------------
	got, err := reconciler.Reconcile(ctx,
		ctrl.Request{NamespacedName: types.NamespacedName{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		}})
	if err != nil {
		By("Reconcile Error")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(ctrl.Result{Requeue: true, RequeueAfter: 1000000000}))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDCIAppConnection01Name,
			Namespace: cDCIAppConnectionNamespace,
		},
		&stDCIAppConnectionCR)
	if err != nil {
		fmt.Println("Failed to get DCIAppConnectionCR.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DCIAppConnectionCR.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDCIAppConnectionCR.Finalizers).To(BeEmpty())
	Expect(stDCIAppConnectionCR.Status.Phase).To(Equal("Pending"))

	var stConnectEndpointCR examplecomv1.ConnectEndpoint
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Not found  to get ConnectEndpointCR FromX1.")
			fmt.Println(err)
		} else {
			fmt.Println("Failed to get ConnectEndpointCR FromX1.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get ConnectEndpointCR FromX1.")
		Expect(err).To(HaveOccurred())
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cConnectEndpointNamespace,
		},
		&stConnectEndpointCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Not found  to get ConnectEndpointCR ToY1.")
			fmt.Println(err)
		} else {
			fmt.Println("Failed to get ConnectEndpointCR ToY1.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get ConnectEndpointCR ToY1.")
		Expect(err).To(HaveOccurred())
	}

	var stServiceCR corev1.Service
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToA,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Not found  to get ServiceCR ToA.")
			fmt.Println(err)
		} else {
			fmt.Println("Failed to get ServiceCR ToA.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get ServiceCR ToA.")
		Expect(err).To(HaveOccurred())
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToC,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Not found  to get ServiceCR ToC.")
			fmt.Println(err)
		} else {
			fmt.Println("Failed to get ServiceCR ToC.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get ServiceCR ToC.")
		Expect(err).To(HaveOccurred())
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cService01NameToY,
			Namespace: cServiceNamespace,
		},
		&stServiceCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Not found  to get ServiceCR ToY.")
			fmt.Println(err)
		} else {
			fmt.Println("Failed to get ServiceCR ToY.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get ServiceCR ToY.")
		Expect(err).To(HaveOccurred())
	}

	var stEndpointSliceCR discoveryv1.EndpointSlice
	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToA,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Not found  to get EndpointSliceCR ToA.")
			fmt.Println(err)
		} else {
			fmt.Println("Failed to get EndpointSliceCR ToA.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get EndpointSliceCR ToA.")
		Expect(err).To(HaveOccurred())
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToC,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Not found  to get EndpointSliceCR ToC.")
			fmt.Println(err)
		} else {
			fmt.Println("Failed to get EndpointSliceCR ToC.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get EndpointSliceCR ToC.")
		Expect(err).To(HaveOccurred())
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cEndpointSlice01NameToY,
			Namespace: cEndpointSliceNamespace,
		},
		&stEndpointSliceCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Not found  to get EndpointSliceCR ToY.")
			fmt.Println(err)
		} else {
			fmt.Println("Failed to get EndpointSliceCR ToY.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get EndpointSliceCR ToY.")
		Expect(err).To(HaveOccurred())
	}

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameFromX,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("NotFound to get DeploymentCR FromX.")
			fmt.Println(err)
			Expect(err).To(HaveOccurred())
		} else {
			fmt.Println("Failed to get DeploymentCR FromX.")
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		fmt.Println("Success to get DeploymentCR FromX.")
		Expect(err).To(HaveOccurred())
	}
	Expect(stDeploymentCR.Finalizers).To(BeEmpty())
	Expect(stDeploymentCR.ObjectMeta.Labels).NotTo(Equal(Deployment01FromXLabels))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToA,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).To(BeEmpty())
	Expect(stDeploymentCR.ObjectMeta.Labels).NotTo(Equal(Deployment01ToALabels))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToC,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).To(BeEmpty())
	Expect(stDeploymentCR.ObjectMeta.Labels).NotTo(Equal(Deployment01ToCLabels))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cDeployment01NameToY,
			Namespace: cDeploymentNamespace,
		},
		&stDeploymentCR)
	if err != nil {
		fmt.Println("Failed to get DeploymentCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get DeploymentCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stDeploymentCR.Finalizers).To(BeEmpty())
	Expect(stDeploymentCR.ObjectMeta.Labels).NotTo(Equal(Deployment01ToYLabels))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameFromX,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR FromX.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR FromX.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToA,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToA.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToA.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToC,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToC.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToC.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cReplicaSet01NameToY,
			Namespace: cReplicaSetNamespace,
		},
		&stReplicaSetCR)
	if err != nil {
		fmt.Println("Failed to get ReplicaSetCR ToY.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get ReplicaSetCR ToY.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stReplicaSetCR.Finalizers).To(BeEmpty())

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameFromX1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR FromX1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR FromX1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01FromX1.Labels))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01ToA1.Labels))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToA2,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToA2.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToA2.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01ToA2.Labels))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToC1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToC1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToC1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01ToC1.Labels))

	err = k8sClient.Get(ctx,
		client.ObjectKey{
			Name:      cPod01NameToY1,
			Namespace: cPodNamespace,
		},
		&stPodCR)
	if err != nil {
		fmt.Println("Failed to get PodCR ToY1.")
		fmt.Println(err)
	} else {
		fmt.Println("Success to get PodCR ToY1.")
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(stPodCR.Finalizers).To(BeEmpty())
	Expect(stPodCR.ObjectMeta.Labels).To(Equal(Pod01ToY1.Labels))

	fmt.Println(writer.String())
	fmt.Println("<<<<<")
	fmt.Println("<<<<< dciappconnection_controller_test.go")
	fmt.Println("<<<<< End Test22_1")
	fmt.Println("<<<<<")
}
