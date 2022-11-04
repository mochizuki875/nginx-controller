/*
Copyright 2022.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NginxSpec defines the desired state of Nginx
type NginxSpec struct {
	// +kubebuilder:default = 1
	// +kubebuilder:validation:Minimum = 0

	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:default = ClusterIP

	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
}

// NginxStatus defines the observed state of Nginx
type NginxStatus struct {
	DeploymentName    string `json:"deploymentName"`
	AvailableReplicas int32  `json:"availableReplicas"`
	ServiceName       string `json:"serviceName"`

	ClusterIP string `json:"clusterIP,omitempty"`

	ExternalIP string `json:"externalIP,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:shortName="ng"
// +kubebuilder:printcolumn:JSONPath=".status.availableReplicas",name=Replicas,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.serviceName",name=Service_Name,type=string
// +kubebuilder:printcolumn:JSONPath=".status.clusterIP",name=Cluster-IP,type=string
// +kubebuilder:printcolumn:JSONPath=".status.externalIP",name=External-IP,type=string

// Nginx is the Schema for the nginxes API
type Nginx struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NginxSpec   `json:"spec,omitempty"`
	Status NginxStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NginxList contains a list of Nginx
type NginxList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nginx `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Nginx{}, &NginxList{})
}
