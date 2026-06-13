// Package v1alpha1 defines the FunctionFly vault CRDs.
//
// VaultSecret   — sync a FunctionFly vault secret into a K8s Secret
// VaultDynamicCredential — generate short-lived dynamic credentials
//
//	and inject them into pods as env vars.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultSecretSpec is the desired state of a VaultSecret.
type VaultSecretSpec struct {
	// SecretID is the FunctionFly vault secret UUID.
	SecretID string `json:"secretId"`

	// TenantID is the FunctionFly tenant ID.
	TenantID string `json:"tenantId"`

	// APITokenRef references a Kubernetes secret holding a
	// FunctionFly API token (the secret must contain a key called
	// "token"). The operator uses this token to fetch the secret
	// from the FunctionFly API.
	APITokenRef SecretRef `json:"apiTokenRef"`

	// APIServer is the FunctionFly API base URL. Defaults to
	// https://api.functionfly.com.
	APIServer string `json:"apiServer,omitempty"`

	// RefreshInterval is how often the operator polls the vault for
	// updates. Default: 30s. Set to 0 to disable polling (you then
	// need an out-of-band mechanism to trigger reconciliation).
	RefreshInterval string `json:"refreshInterval,omitempty"`

	// SecretType controls the type of the produced Kubernetes secret.
	// One of: Opaque, kubernetes.io/tls, kubernetes.io/dockerconfigjson.
	// Default: Opaque.
	SecretType string `json:"secretType,omitempty"`

	// DecryptionSecretRef points to a Kubernetes secret that holds
	// the operator's envelope-decryption key. Optional for v1
	// (server returns ciphertext only); the operator always stores
	// the encrypted payload unless this is set and contains a
	// "key" entry.
	DecryptionSecretRef *SecretRef `json:"decryptionSecretRef,omitempty"`
}

// SecretRef references a Kubernetes secret by name and namespace.
type SecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key,omitempty"`
}

// VaultSecretStatus reports the last sync state.
type VaultSecretStatus struct {
	LastSyncedAt    metav1.Time `json:"lastSyncedAt,omitempty"`
	LastError       string      `json:"lastError,omitempty"`
	ObservedVersion int         `json:"observedVersion,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type VaultSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultSecretSpec   `json:"spec,omitempty"`
	Status VaultSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VaultSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultSecret `json:"items"`
}

// ============================================================================
// VaultDynamicCredential
// ============================================================================

// VaultDynamicCredentialSpec configures dynamic credential injection.
type VaultDynamicCredentialSpec struct {
	// CredentialID is the FunctionFly dynamic credential UUID.
	CredentialID string `json:"credentialId"`

	// TenantID is the FunctionFly tenant ID.
	TenantID string `json:"tenantId"`

	// APITokenRef references a Kubernetes secret with a "token" key.
	APITokenRef SecretRef `json:"apiTokenRef"`

	// APIServer is the FunctionFly API base URL. Defaults to
	// https://api.functionfly.com.
	APIServer string `json:"apiServer,omitempty"`

	// TTLSeconds is how long each lease should last. Must not exceed
	// the credential's MaxTTLSeconds. Default: 3600.
	TTLSeconds int `json:"ttlSeconds,omitempty"`

	// RenewBeforeSeconds is how long before expiry the operator
	// should renew the lease. Default: 300.
	RenewBeforeSeconds int `json:"renewBeforeSeconds,omitempty"`

	// Env is the list of env-var names the credential will be exposed
	// as inside the pod. The operator mounts a tmpfs volume at
	// /vault/dynamic and writes:
	//   <env_prefix>_USERNAME
	//   <env_prefix>_PASSWORD
	//   <env_prefix>_HOST
	//   <env_prefix>_PORT
	//   <env_prefix>_DATABASE
	//   <env_prefix>_EXPIRES_AT
	Env EnvMapping `json:"env"`
}

// EnvMapping maps the produced credential fields into env-var names.
type EnvMapping struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	Database  string `json:"database"`
	ExpiresAt string `json:"expiresAt"`
}

// VaultDynamicCredentialStatus reports the current lease.
type VaultDynamicCredentialStatus struct {
	LeaseID     string      `json:"leaseId,omitempty"`
	ExpiresAt   metav1.Time `json:"expiresAt,omitempty"`
	LastRenewed metav1.Time `json:"lastRenewed,omitempty"`
	LastError   string      `json:"lastError,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type VaultDynamicCredential struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultDynamicCredentialSpec   `json:"spec,omitempty"`
	Status VaultDynamicCredentialStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VaultDynamicCredentialList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultDynamicCredential `json:"items"`
}
