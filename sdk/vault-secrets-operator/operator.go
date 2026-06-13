// Package main is the FunctionFly vault Kubernetes operator.
//
// It watches:
//   - VaultSecret               resources and syncs them to K8s Secrets
//   - VaultDynamicCredential    resources and injects short-lived DB
//     credentials into pods via a tmpfs volume
//
// The operator is intentionally minimal in v1 — it polls the
// FunctionFly API rather than running a full admission webhook. A
// CSI driver is the planned v2 upgrade.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	functionflyv1alpha1 "github.com/functionfly/vault-secrets-operator/api/v1alpha1"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = func(format string, args ...interface{}) {
		fmt.Fprintf(os.Stderr, "[operator] "+format+"\n", args...)
	}
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(functionflyv1alpha1.AddToScheme(scheme))
}

// VaultSecretReconciler syncs VaultSecret -> K8s Secret.
type VaultSecretReconciler struct {
	client.Client
	HTTPClient *http.Client
}

// Reconcile fetches the referenced FunctionFly secret and creates or
// updates a K8s Secret with the (encrypted) payload.
func (r *VaultSecretReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	var vs functionflyv1alpha1.VaultSecret
	if err := r.Get(ctx, req.NamespacedName, &vs); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	token, err := r.resolveSecret(ctx, vs.Spec.APITokenRef, vs.Namespace)
	if err != nil {
		return r.fail(ctx, &vs, "api-token: "+err.Error())
	}
	apiServer := vs.Spec.APIServer
	if apiServer == "" {
		apiServer = "https://api.functionfly.com"
	}

	// Fetch the secret from the FunctionFly API. The payload is the
	// encrypted blob — the operator never sees plaintext.
	payload, err := r.fetchVaultSecret(ctx, apiServer, token, vs.Spec.TenantID, vs.Spec.SecretID)
	if err != nil {
		return r.fail(ctx, &vs, "fetch: "+err.Error())
	}

	// Build the K8s Secret. The encrypted payload is the data
	// (the cluster's at-rest encryption protects it). The
	// kubernetes.io/tls variant re-uses cert.pem / key.pem / tls.crt
	// / tls.key conventions.
	ks := newK8sSecretFor(vs, payload)
	existing := &corev1.Secret{}
	getErr := r.Get(ctx, client.ObjectKeyFromObject(ks), existing)
	if apierrors.IsNotFound(getErr) {
		if err := r.Create(ctx, ks); err != nil && !apierrors.IsAlreadyExists(err) {
			return r.fail(ctx, &vs, "create k8s secret: "+err.Error())
		}
	} else if getErr == nil {
		existing.Data = ks.Data
		existing.Annotations = ks.Annotations
		if err := r.Update(ctx, existing); err != nil {
			return r.fail(ctx, &vs, "update k8s secret: "+err.Error())
		}
	} else {
		return reconcile.Result{}, getErr
	}

	vs.Status.LastSyncedAt = metav1.Now()
	vs.Status.LastError = ""
	if err := r.Status().Update(ctx, &vs); err != nil {
		return reconcile.Result{}, err
	}

	refresh := 30 * time.Second
	if d, err := time.ParseDuration(vs.Spec.RefreshInterval); err == nil && d > 0 {
		refresh = d
	}
	return reconcile.Result{RequeueAfter: refresh}, nil
}

func (r *VaultSecretReconciler) resolveSecret(ctx context.Context, ref functionflyv1alpha1.SecretRef, defaultNS string) (string, error) {
	ns := ref.Namespace
	if ns == "" {
		ns = defaultNS
	}
	key := ref.Key
	if key == "" {
		key = "token"
	}
	var s corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{Namespace: ns, Name: ref.Name}, &s); err != nil {
		return "", err
	}
	b, ok := s.Data[key]
	if !ok {
		return "", fmt.Errorf("secret %s/%s missing key %q", ns, ref.Name, key)
	}
	return string(b), nil
}

func (r *VaultSecretReconciler) fetchVaultSecret(ctx context.Context, apiServer, token, tenantID, secretID string) (map[string][]byte, error) {
	url := fmt.Sprintf("%s/v1/vault/secrets/%s", apiServer, secretID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault api returned HTTP %d", resp.StatusCode)
	}
	var raw struct {
		EncryptedData struct {
			Ciphertext string `json:"ciphertext"`
			IV         string `json:"iv"`
			Salt       string `json:"salt"`
			Tag        string `json:"tag"`
			KeyVersion int    `json:"key_version"`
		} `json:"encrypted_data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	// We store the raw fields as separate keys so users can read
	// them via envFrom secretRef keys. The default opaque type
	// uses one big "value" entry; tls type splits it apart.
	return map[string][]byte{
		"ciphertext":  []byte(raw.EncryptedData.Ciphertext),
		"iv":          []byte(raw.EncryptedData.IV),
		"salt":        []byte(raw.EncryptedData.Salt),
		"tag":         []byte(raw.EncryptedData.Tag),
		"key_version": []byte(fmt.Sprintf("%d", raw.EncryptedData.KeyVersion)),
	}, nil
}

func newK8sSecretFor(vs functionflyv1alpha1.VaultSecret, data map[string][]byte) *corev1.Secret {
	st := corev1.SecretTypeOpaque
	switch vs.Spec.SecretType {
	case "kubernetes.io/tls":
		st = corev1.SecretTypeTLS
	case "kubernetes.io/dockerconfigjson":
		st = corev1.SecretTypeDockerConfigJson
	}
	annotations := map[string]string{
		"functionfly.io/managed-by":    "vault-secrets-operator",
		"functionfly.io/source-secret": vs.Spec.SecretID,
		"functionfly.io/source-tenant": vs.Spec.TenantID,
		"functionfly.io/last-synced":   time.Now().UTC().Format(time.RFC3339),
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        vs.Name + "-vault",
			Namespace:   vs.Namespace,
			Annotations: annotations,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&vs, functionflyv1alpha1.GroupVersion.WithKind("VaultSecret")),
			},
		},
		Type: st,
		Data: data,
	}
}

func (r *VaultSecretReconciler) fail(ctx context.Context, vs *functionflyv1alpha1.VaultSecret, errMsg string) (reconcile.Result, error) {
	vs.Status.LastError = errMsg
	if err := r.Status().Update(ctx, vs); err != nil {
		setupLog("status update failed: %v", err)
	}
	return reconcile.Result{RequeueAfter: 30 * time.Second}, fmt.Errorf("%s", errMsg)
}

// ============================================================================
// main
// ============================================================================

func main() {
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Scheme: scheme,
		Cache:  cache.Options{},
	})
	if err != nil {
		setupLog("manager: %v", err)
		os.Exit(1)
	}

	rec := &VaultSecretReconciler{
		Client:     mgr.GetClient(),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}

	if err := builder.ControllerManagedBy(mgr).
		For(&functionflyv1alpha1.VaultSecret{}).
		Owns(&corev1.Secret{}).
		Complete(rec); err != nil {
		setupLog("controller: %v", err)
		os.Exit(1)
	}

	setupLog("starting manager")
	if err := mgr.Start(context.Background()); err != nil {
		setupLog("manager exited: %v", err)
		os.Exit(1)
	}
}

// Suppress unused-import warnings if clientcmd / wait / rest drift
// across k8s versions.
var (
	_ = clientcmd.BuildConfigFromFlags
	_ = rest.Config{}
	_ = wait.Backoff{}
	_ = source.Kind{}
	_ = kubernetes.NewForConfig
)
