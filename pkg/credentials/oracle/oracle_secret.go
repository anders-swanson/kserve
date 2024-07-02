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
	v1 "k8s.io/api/core/v1"
)

type AuthType string

const (
	AuthenticationType = "OCI_AUTHENTICATION_TYPE"
	Region             = "OCI_REGION"
	User               = "OCI_USER"
	Tenant             = "OCI_TENANT"
	Privatekey         = "OCI_PRIVATEKEY"
	Passphrase         = "OCI_PASSPHRASE"
	Fingerprint        = "OCI_FINGERPRINT"

	AuthenticationTypeSimple            AuthType = "simple"
	AuthenticationTypeInstancePrincipal AuthType = "instance-principal"
	// nolint: unused
	AuthenticationTypeWorkloadIdentity AuthType = "workload-identity"

	resourcePrincipalVersion   = "OCI_RESOURCE_PRINCIPAL_VERSION"
	resourcePrincipalVersionV2 = "2.2"
	resourcePrincipalRegion    = "OCI_RESOURCE_PRINCIPAL_REGION"
)

func BuildSecretEnvs(secret *v1.Secret) []v1.EnvVar {
	var envs []v1.EnvVar

	switch AuthType(secret.Data[AuthenticationType]) {
	case AuthenticationTypeInstancePrincipal:
		break // Nothing to set for instance principal.
	case AuthenticationTypeSimple:
		envs = addKeysFromSecret(envs, secret, Region, Fingerprint, User, Tenant, Privatekey, Passphrase)
	default: // Default to configuration file auth
		envs = addKeysFromSecret(envs, secret, Region)
		// Workload Identity requires specific environment variables to be set (OCI SDK).
		envs = append(envs, v1.EnvVar{
			Name:  resourcePrincipalVersion,
			Value: resourcePrincipalVersionV2,
		}, v1.EnvVar{
			Name: resourcePrincipalRegion,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: Region,
				},
			},
		})
	}
	envs = addKeyFromSecret(envs, secret, AuthenticationType)

	return envs
}

func addKeysFromSecret(envs []v1.EnvVar, secret *v1.Secret, keys ...string) []v1.EnvVar {
	for _, key := range keys {
		envs = addKeyFromSecret(envs, secret, key)
	}
	return envs
}

func addKeyFromSecret(envs []v1.EnvVar, secret *v1.Secret, key string) []v1.EnvVar {
	if _, ok := secret.Data[key]; ok {
		envs = append(envs, newEnvVar(secret.Name, key))
	}
	return envs
}

func newEnvVar(secretName, key string) v1.EnvVar {
	return v1.EnvVar{
		Name: key,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key: key,
			},
		},
	}
}
