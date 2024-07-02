/*
Copyright 2024 The KServe Authors.

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

package oracle

import (
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"testing"
)

const secretName = "oci-secret"

func TestOracleSecret(t *testing.T) {
	scenarios := map[string]struct {
		secret   *v1.Secret
		expected []v1.EnvVar
	}{
		"simple-auth": {
			secret:   makeSecret(AuthenticationTypeSimple, Region, Fingerprint, User, Tenant, Privatekey),
			expected: makeValueFromEnvVars(Region, Fingerprint, User, Tenant, Privatekey, AuthenticationType),
		},
		"simple-auth-passphrase": {
			secret:   makeSecret(AuthenticationTypeSimple, Region, Fingerprint, User, Tenant, Privatekey, Passphrase),
			expected: makeValueFromEnvVars(Region, Fingerprint, User, Tenant, Privatekey, Passphrase, AuthenticationType),
		},
		"workload-identity": {
			secret: makeSecret(AuthenticationTypeWorkloadIdentity, Region),
			expected: append([]v1.EnvVar{newEnvVar(secretName, Region), newEnvVar(secretName, AuthenticationType)}, v1.EnvVar{
				Name:  resourcePrincipalVersion,
				Value: resourcePrincipalVersionV2,
			}, v1.EnvVar{
				Name: resourcePrincipalRegion,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: secretName,
						},
						Key: Region,
					},
				},
			}),
		},
		"instance-principal": {
			secret:   makeSecret(AuthenticationTypeInstancePrincipal),
			expected: []v1.EnvVar{newEnvVar(secretName, AuthenticationType)},
		},
	}

	for name, scenario := range scenarios {
		t.Run(name, func(t *testing.T) {
			envs := BuildSecretEnvs(scenario.secret)
			sort.Slice(envs, func(i, j int) bool {
				return envs[i].Name < envs[j].Name
			})
			sort.Slice(scenario.expected, func(i, j int) bool {
				return scenario.expected[i].Name < scenario.expected[j].Name
			})
			if diff := cmp.Diff(scenario.expected, envs); diff != "" {
				t.Errorf("Test %q unexpected result (-want +got): %v", name, diff)
			}
		})
	}
}

func makeSecret(authType AuthType, keys ...string) *v1.Secret {
	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
	}
	s.Data = map[string][]byte{}
	s.Data[AuthenticationType] = []byte(authType)
	for _, k := range keys {
		s.Data[k] = []byte(k)
	}
	return s
}

func makeValueFromEnvVars(keys ...string) []v1.EnvVar {
	var envs []v1.EnvVar
	for _, k := range keys {
		envs = append(envs, newEnvVar(secretName, k))
	}
	return envs
}
