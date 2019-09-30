/*
Copyright 2019 the original author or authors.

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

package requests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gomodules.xyz/jsonpatch/v2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type noResourceRequests struct {
	client  client.Client
	decoder *admission.Decoder
}

func NewNoResourceRequests() *noResourceRequests {
	return &noResourceRequests{}
}

var _ ExtendedHandler = &noResourceRequests{}

var (
	podResource = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
)

func (i *noResourceRequests) Handle(ctx context.Context, req admission.Request) admission.Response {
	// Ignore non-pod resources.
	if req.Resource != podResource {
		return admission.Patched(fmt.Sprintf("unexpected resource (not a pod): %v", req.Resource))
	}

	rawPod := req.Object.Raw
	pod := corev1.Pod{}

	if len(rawPod) > 0 {
		err := i.decoder.Decode(req, &pod)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("decoding pod: %s", err))
		}
	} else {
		err := i.client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: req.Name}, &pod)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("getting pod: %s", err))
		}
	}

	modified := false

	for j, _ := range pod.Spec.InitContainers {
		container := &pod.Spec.InitContainers[j]
		if _, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
			container.Resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(0, "m")
			modified = true
		}
		if _, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
			container.Resources.Requests[corev1.ResourceMemory] = *resource.NewQuantity(0, "Mi")
			modified = true
		}
	}

	for j, _ := range pod.Spec.Containers {
		container := &pod.Spec.Containers[j]
		if _, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
			container.Resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(0, "m")
			modified = true
		}
		if _, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
			container.Resources.Requests[corev1.ResourceMemory] = *resource.NewQuantity(0, "Mi")
			modified = true
		}
	}

	if !modified {
		return admission.Patched("unmodified")
	}

	updatedPodRaw, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("marshaling pod: %s", err))
	}

	patch, err := jsonpatch.CreatePatch(rawPod, updatedPodRaw)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("creating patch: %s", err))
	}

	return admission.Patched("pod requests stripped", patch...)
}

func (i *noResourceRequests) InjectClient(client client.Client) error {
	i.client = client
	return nil
}

func (i *noResourceRequests) InjectDecoder(decoder *admission.Decoder) error {
	i.decoder = decoder
	return nil
}
