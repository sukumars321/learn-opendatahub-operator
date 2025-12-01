# Day 9: Practical Examples - Kubebuilder Markers in Action

## 🎯 Real ODH Marker Examples

This document contains actual examples from the OpenDataHub operator codebase, showing how markers are used in practice.

---

## 🔐 RBAC Markers from ODH Controllers

### DataScienceCluster Controller RBAC
From `controllers/datasciencecluster_controller.go`:

```go
//+kubebuilder:rbac:groups=datasciencecluster.opendatahub.io,resources=datascienceclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=datasciencecluster.opendatahub.io,resources=datascienceclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=datasciencecluster.opendatahub.io,resources=datascienceclusters/finalizers,verbs=update

// Core Kubernetes resources
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

// Apps resources
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch

// Route resources (OpenShift)
//+kubebuilder:rbac:groups="route.openshift.io",resources=routes,verbs=get;list;watch;create;update;patch;delete

// Networking resources
//+kubebuilder:rbac:groups="networking.k8s.io",resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update;patch;delete
```

### Component-Specific RBAC
```go
// Dashboard component needs these additional permissions
//+kubebuilder:rbac:groups="authorization.openshift.io",resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="authorization.openshift.io",resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

// Model serving needs these
//+kubebuilder:rbac:groups="serving.kserve.io",resources=inferenceservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="serving.kserve.io",resources=servingruntimes,verbs=get;list;watch;create;update;patch;delete
```

---

## 🛠️ CRD Generation Markers

### DataScienceCluster Type Definition
From `apis/datasciencecluster/v1/datasciencecluster_types.go`:

```go
// DataScienceCluster is the Schema for the datascienceclusters API
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=datascienceclusters,scope=Namespaced
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
type DataScienceCluster struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   DataScienceClusterSpec   `json:"spec,omitempty"`
    Status DataScienceClusterStatus `json:"status,omitempty"`
}
```

### Validation Markers in Spec
```go
type DataScienceClusterSpec struct {
    // Components defines the spec for each ODH component
    //+kubebuilder:validation:Optional
    Components Components `json:"components,omitempty"`
}

type Components struct {
    // Dashboard component configuration
    //+kubebuilder:validation:Optional
    Dashboard Dashboard `json:"dashboard,omitempty"`

    // Workbenches component configuration
    //+kubebuilder:validation:Optional
    Workbenches Workbenches `json:"workbenches,omitempty"`

    // ModelMeshServing component configuration
    //+kubebuilder:validation:Optional
    ModelMeshServing ModelMeshServing `json:"modelmeshserving,omitempty"`

    // DataSciencePipelines component configuration
    //+kubebuilder:validation:Optional
    DataSciencePipelines DataSciencePipelines `json:"datasciencepipelines,omitempty"`

    // Kserve component configuration
    //+kubebuilder:validation:Optional
    Kserve Kserve `json:"kserve,omitempty"`

    // CodeFlare component configuration
    //+kubebuilder:validation:Optional
    CodeFlare CodeFlare `json:"codeflare,omitempty"`

    // Ray component configuration
    //+kubebuilder:validation:Optional
    Ray Ray `json:"ray,omitempty"`

    // TrustyAI component configuration
    //+kubebuilder:validation:Optional
    TrustyAI TrustyAI `json:"trustyai,omitempty"`

    // ModelRegistry component configuration
    //+kubebuilder:validation:Optional
    ModelRegistry ModelRegistry `json:"modelregistry,omitempty"`
}
```

### Component Configuration with Validation
```go
type ManagementState string

const (
    // Managed means the operator is actively managing the component
    Managed ManagementState = "Managed"
    // Removed means the operator is actively removing the component
    Removed ManagementState = "Removed"
)

type Dashboard struct {
    // Management state of the component
    //+kubebuilder:validation:Enum=Managed;Removed
    //+kubebuilder:default=Managed
    ManagementState ManagementState `json:"managementState,omitempty"`
}
```

---

## 🔗 Webhook Markers

### Admission Webhook Configuration
From webhook files:

```go
//+kubebuilder:webhook:path=/mutate-datasciencecluster-opendatahub-io-v1-datasciencecluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=datasciencecluster.opendatahub.io,resources=datascienceclusters,verbs=create;update,versions=v1,name=mdatasciencecluster.kb.io,admissionReviewVersions=v1

//+kubebuilder:webhook:path=/validate-datasciencecluster-opendatahub-io-v1-datasciencecluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=datasciencecluster.opendatahub.io,resources=datascienceclusters,verbs=create;update,versions=v1,name=vdatasciencecluster.kb.io,admissionReviewVersions=v1

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *DataScienceCluster) Default() {
    datascienceclusterlog.Info("default", "name", r.Name)

    // Set default values for components if not specified
    if r.Spec.Components.Dashboard.ManagementState == "" {
        r.Spec.Components.Dashboard.ManagementState = Managed
    }
}
```

---

## 🏗️ Generated Files Walkthrough

### 1. Generated CRD (config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml)

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.9.2
  creationTimestamp: null
  name: datascienceclusters.datasciencecluster.opendatahub.io
spec:
  group: datasciencecluster.opendatahub.io
  names:
    kind: DataScienceCluster
    listKind: DataScienceClusterList
    plural: datascienceclusters
    singular: datasciencecluster
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    - jsonPath: .status.phase
      name: Phase
      type: string
    - jsonPath: .status.conditions[?(@.type=='Ready')].status
      name: Ready
      type: string
    name: v1
    schema:
      openAPIV3Schema:
        description: DataScienceCluster is the Schema for the datascienceclusters API
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation...'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource...'
            type: string
          metadata:
            type: object
          spec:
            description: DataScienceClusterSpec defines the desired state of DataScienceCluster
            properties:
              components:
                properties:
                  dashboard:
                    properties:
                      managementState:
                        default: Managed
                        enum:
                        - Managed
                        - Removed
                        type: string
                    type: object
```

### 2. Generated RBAC (config/rbac/role.yaml)

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: manager-role
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - datasciencecluster.opendatahub.io
  resources:
  - datascienceclusters
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - datasciencecluster.opendatahub.io
  resources:
  - datascienceclusters/finalizers
  verbs:
  - update
- apiGroups:
  - datasciencecluster.opendatahub.io
  resources:
  - datascienceclusters/status
  verbs:
  - get
  - patch
  - update
```

### 3. Generated DeepCopy Methods (apis/datasciencecluster/v1/zz_generated.deepcopy.go)

```go
//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
    runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataScienceCluster) DeepCopyInto(out *DataScienceCluster) {
    *out = *in
    out.TypeMeta = in.TypeMeta
    in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
    in.Spec.DeepCopyInto(&out.Spec)
    in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataScienceCluster.
func (in *DataScienceCluster) DeepCopy() *DataScienceCluster {
    if in == nil {
        return nil
    }
    out := new(DataScienceCluster)
    in.DeepCopyInto(out)
    return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataScienceCluster) DeepCopyObject() runtime.Object {
    if c := in.DeepCopy(); c != nil {
        return c
    }
    return nil
}
```

---

## ⚙️ Make Generate Workflow

### Makefile Target
```makefile
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
    $(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2)

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
    $(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
    $(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
```

### What Each Command Does

#### `make generate`
```bash
controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
```
- Scans all Go files in the project
- Generates DeepCopy methods for types marked with `//+kubebuilder:object:generate=true`
- Creates `zz_generated.deepcopy.go` files

#### `make manifests`
```bash
controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
```
- Scans for RBAC markers → generates `config/rbac/role.yaml`
- Scans for CRD markers → generates CRD YAML files in `config/crd/bases/`
- Scans for webhook markers → generates webhook configurations

---

## 🎯 Marker Pattern Summary

### Common ODH Patterns

#### Every Controller Has:
```go
//+kubebuilder:rbac:groups=GROUPNAME,resources=RESOURCENAME,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=GROUPNAME,resources=RESOURCENAME/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=GROUPNAME,resources=RESOURCENAME/finalizers,verbs=update
```

#### Every Root Type Has:
```go
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
```

#### Common Validation Patterns:
```go
//+kubebuilder:validation:Optional
//+kubebuilder:validation:Enum=Value1;Value2;Value3
//+kubebuilder:default=DefaultValue
```

---

## 📚 Learning Summary

These examples show how ODH leverages Kubebuilder markers to:

1. **Automate RBAC**: Generate precise permissions for each controller
2. **Define APIs**: Create comprehensive CRDs with validation
3. **Enable Webhooks**: Configure admission control
4. **Maintain Consistency**: Ensure all generated files follow the same patterns
5. **Reduce Errors**: Eliminate manual manifest writing and maintenance

The markers serve as the single source of truth, keeping the Go code and Kubernetes manifests perfectly synchronized!