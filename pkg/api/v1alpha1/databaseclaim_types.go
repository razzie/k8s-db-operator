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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type DatabaseType string

const (
	PostgreSQL DatabaseType = "PostgreSQL"
)

// DatabaseClaimSpec defines the desired state of DatabaseClaim
type DatabaseClaimSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Type of the database to generate credentials for, known values are (`PostgreSQL`).
	DatabaseType DatabaseType `json:"databaseType"`

	// SecretName is the name of the secret resource that will be automatically
	// created and managed by this Certificate resource.
	// It will be populated with the login credentials.
	SecretName string `json:"secretName"`
}

// DatabaseClaimStatus defines the observed state of DatabaseClaim
type DatabaseClaimStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=dbclaim
//+kubebuilder:printcolumn:name="Database",type="string",JSONPath=".spec.databaseType"
//+kubebuilder:printcolumn:name="Secret",type="string",JSONPath=".spec.secretName"
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"

// DatabaseClaim is the Schema for the databaseclaims API
type DatabaseClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseClaimSpec   `json:"spec,omitempty"`
	Status DatabaseClaimStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DatabaseClaimList contains a list of DatabaseClaim
type DatabaseClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseClaim{}, &DatabaseClaimList{})
}
