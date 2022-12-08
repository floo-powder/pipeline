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
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/floo-powder/pkg/consts"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SourceSpec defines the desired state of Source
type SourceSpec struct {
	// Host connect host
	Host string `json:"host"`
	// Port connect port
	Port int `json:"port"`
	// Plugin for db
	Plugin string `json:"plugin"`
	// Config for connect a json string
	Config string `json:"config"`
}

func (in *SourceSpec) checkValue(value interface{}, kind consts.DatabaseConfigFieldKind) bool {
	ok := false
	switch kind {
	case consts.StringConfig, consts.PasswordConfig:
		_, ok = value.(string)
	case consts.IntConfig:
		_, ok = value.(int)
	case consts.BoolConfig:
		_, ok = value.(bool)
	}
	return ok
}

// ValidateConfig check config is validate
func (in *SourceSpec) ValidateConfig(pluginObj Plugin) (reason string, message string) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(in.Config), &config); err != nil {
		reason = string(consts.BadRequest)
		message = "Unknown Config Format"
		return
	}
	for _, fieldConfig := range pluginObj.Spec.Config {
		isExist := false
		for key, value := range config {
			if key != fieldConfig.Name {
				continue
			}
			isExist = true
			ok := in.checkValue(value, fieldConfig.Kind)
			if !ok {
				reason = string(consts.BadRequest)
				message = fmt.Sprintf(
					"Wrong parameter value type: %s should be %s",
					fieldConfig.Name, fieldConfig.Kind,
				)
				return
			}
		}
		if !isExist {
			reason = string(consts.BadRequest)
			message = fmt.Sprintf("Missing parameter: %s", fieldConfig.Name)
			return
		}
	}
	reason = string(consts.Succeed)
	return
}

// SourceStatus defines the observed state of Source
type SourceStatus struct {
	*duckv1.Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Source is the Schema for the sources API
type Source struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SourceSpec   `json:"spec,omitempty"`
	Status SourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SourceList contains a list of Source
type SourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Source `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Source{}, &SourceList{})
}
