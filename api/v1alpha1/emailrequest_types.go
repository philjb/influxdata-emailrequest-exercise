/*
Copyright 2023.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EmailRequestSpec defines the desired state of EmailRequest
type EmailRequestSpec struct {
	// name of the person to send the email to
	Name string `json:"name,omitempty"`
	// email address of the person to send the email to
	Address string `json:"address,omitempty"`
	// email's retry policy on blocked emails, true means to retry, false means to not retry
	RetryBlockedPolicy bool `json:"retryBlockedPolicy,omitempty"`
}

// EmailRequestStatus defines the observed state of EmailRequest
type EmailRequestStatus struct {
	// Represents the observations of the emailrequest's current state.
	// Known .status.conditions.type are: "Available", "Progressing", and "Degraded"
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EmailRequest is the Schema for the emailrequests API
type EmailRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EmailRequestSpec   `json:"spec,omitempty"`
	Status EmailRequestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EmailRequestList contains a list of EmailRequest
type EmailRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EmailRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EmailRequest{}, &EmailRequestList{})
}
