// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BgpPeerDiscoverySpec defines the desired state of BgpPeerDiscovery
type BgpPeerDiscoverySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// BgpPeerDiscoveryStatus defines the observed state of BgpPeerDiscovery
type BgpPeerDiscoveryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DiscoveredPeers []string `json:"discovered_peers"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BgpPeerDiscovery is the Schema for the bgppeerdiscoveries API
type BgpPeerDiscovery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BgpPeerDiscoverySpec   `json:"spec,omitempty"`
	Status BgpPeerDiscoveryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BgpPeerDiscoveryList contains a list of BgpPeerDiscovery
type BgpPeerDiscoveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BgpPeerDiscovery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BgpPeerDiscovery{}, &BgpPeerDiscoveryList{})
}
