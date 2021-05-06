/*


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
const (
	PhasePending = "PENDING"
	PhaseRunning = "RUNNING"
	PhaseDone    = "DONE"
)

// WebInstallSpec defines the desired state of WebInstall
type WebInstallSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Replicas int32  `json:"replicas,omitempty"`
	Host     string `json:"host,omitempty"`
	Image    string `json:"image,omitempty"`
}

// WebInstallStatus defines the observed state of WebInstall
type WebInstallStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true

// WebInstall is the Schema for the webinstalls API
type WebInstall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebInstallSpec   `json:"spec,omitempty"`
	Status WebInstallStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebInstallList contains a list of WebInstall
type WebInstallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebInstall `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebInstall{}, &WebInstallList{})
}
