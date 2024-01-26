/*
Copyright 2024.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	TopologyValue = "topology.cninanny.sap.cc/value"
)

// LabelDiscoverySpec defines the desired state of LabelDiscovery
type LabelDiscoverySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TopologyLabel is node label used for peer discovery job placement
	TopologyLabel string `json:"topology_label"`
}

// LabelDiscoveryStatus defines the observed state of LabelDiscovery
type LabelDiscoveryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// DiscoveredTopologyValues collects discovered values
	DiscoveredTopologyValues map[string]DiscoveredTopologyValue `json:"discovered_topology_values"`
}

type DiscoveredTopologyValue struct {
	Finalized bool `json:"finalized"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LabelDiscovery is the Schema for the labeldiscoveries API
type LabelDiscovery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabelDiscoverySpec   `json:"spec,omitempty"`
	Status LabelDiscoveryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LabelDiscoveryList contains a list of LabelDiscovery
type LabelDiscoveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabelDiscovery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabelDiscovery{}, &LabelDiscoveryList{})
}
