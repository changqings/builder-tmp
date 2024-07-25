/*
Copyright 2024 changqings.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TmpSpec defines the desired state of Tmp
type TmpSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ServieName string `json:"service_name,omitempty"`
	// +kubebuilder:default=ClusterIP
	ServiceType string `json:"service_type,omitempty"`
	// +kubebuilder:default=80
	Port int `json:"port,omitempty"`
	// +kubebuilder:default=80
	ContainerPort int `json:"container_port,omitempty"`
}

// TmpStatus defines the observed state of Tmp
type TmpStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Tmp is the Schema for the tmps API
type Tmp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TmpSpec   `json:"spec,omitempty"`
	Status TmpStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TmpList contains a list of Tmp
type TmpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tmp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tmp{}, &TmpList{})
}
