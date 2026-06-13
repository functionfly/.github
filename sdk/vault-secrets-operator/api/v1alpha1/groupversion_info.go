// Package v1alpha1 is the registration target for the CRDs. The
// generated zz_generated_deepcopy.go and groupversion_info.go files
// are intentionally omitted — v1 of the operator is hand-rolled
// (we generate them once we promote the API out of v1alpha1).
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersion is the group/version pair for the CRDs.
var GroupVersion = schema.GroupVersion{Group: "functionfly.io", Version: "v1alpha1"}

// SchemeBuilder is the scheme builder for the package.
var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// AddToScheme registers the v1alpha1 types with the runtime scheme.
var AddToScheme = SchemeBuilder.AddToScheme

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&VaultSecret{},
		&VaultSecretList{},
		&VaultDynamicCredential{},
		&VaultDynamicCredentialList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
