# Day 10: Practical Examples - ODH Controller Patterns in Action

## 🎯 Real ODH Controller Code Examples

This document contains actual examples from the OpenDataHub operator codebase, showing controller architecture patterns in practice.

---

## 🏗️ DataScienceCluster Controller Structure

### Main Controller Definition
From `controllers/datasciencecluster_controller.go`:

```go
import (
    "context"
    "fmt"
    "time"

    dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/apis/datasciencecluster/v1"
    "github.com/opendatahub-io/opendatahub-operator/v2/pkg/components"
    "github.com/opendatahub-io/opendatahub-operator/v2/pkg/components/codeflare"
    "github.com/opendatahub-io/opendatahub-operator/v2/pkg/components/dashboard"
    // ... other component imports

    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DSCReconciler reconciles a DataScienceCluster object
type DSCReconciler struct {
    client.Client
    Scheme *runtime.Scheme
    Log    logr.Logger
}
```

### Reconcile Method Implementation
```go
//+kubebuilder:rbac:groups=datasciencecluster.opendatahub.io,resources=datascienceclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=datasciencecluster.opendatahub.io,resources=datascienceclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=datasciencecluster.opendatahub.io,resources=datascienceclusters/finalizers,verbs=update

func (r *DSCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := r.Log.WithValues("datasciencecluster", req.NamespacedName)

    // Step 1: Fetch the DataScienceCluster instance
    instance := &dscv1.DataScienceCluster{}
    err := r.Get(ctx, req.NamespacedName, instance)
    if err != nil {
        if errors.IsNotFound(err) {
            log.Info("DataScienceCluster resource not found. Ignoring since object must be deleted")
            return ctrl.Result{}, nil
        }
        log.Error(err, "Failed to get DataScienceCluster")
        return ctrl.Result{}, err
    }

    // Step 2: Handle deletion
    if instance.GetDeletionTimestamp() != nil {
        return r.handleDeletion(ctx, instance, log)
    }

    // Step 3: Add finalizer if not present
    if !controllerutil.ContainsFinalizer(instance, finalizerName) {
        controllerutil.AddFinalizer(instance, finalizerName)
        return ctrl.Result{}, r.Update(ctx, instance)
    }

    // Step 4: Initialize status if needed
    if instance.Status.Phase == "" {
        instance.Status.Phase = dscv1.PhaseProgressing
        instance.Status.Conditions = []metav1.Condition{}
        err := r.Status().Update(ctx, instance)
        if err != nil {
            log.Error(err, "Failed to update DataScienceCluster status")
            return ctrl.Result{}, err
        }
    }

    // Step 5: Reconcile all components
    componentErrors := r.reconcileComponents(ctx, instance, log)

    // Step 6: Update overall status
    err = r.updateStatus(ctx, instance, componentErrors, log)
    if err != nil {
        log.Error(err, "Failed to update DataScienceCluster status")
        return ctrl.Result{}, err
    }

    // Step 7: Determine requeue behavior
    if len(componentErrors) > 0 {
        log.Info("Some components failed, requeuing", "errors", len(componentErrors))
        return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
    }

    log.Info("DataScienceCluster reconciliation completed successfully")
    return ctrl.Result{}, nil
}
```

### Component Reconciliation Logic
```go
func (r *DSCReconciler) reconcileComponents(ctx context.Context, instance *dscv1.DataScienceCluster, log logr.Logger) map[string]error {
    componentErrors := make(map[string]error)

    // Define component map with their configurations
    components := map[string]struct {
        component       components.ComponentInterface
        managementState dscv1.ManagementState
    }{
        "dashboard": {
            component:       r.newDashboardComponent(instance),
            managementState: instance.Spec.Components.Dashboard.ManagementState,
        },
        "workbenches": {
            component:       r.newWorkbenchesComponent(instance),
            managementState: instance.Spec.Components.Workbenches.ManagementState,
        },
        "modelmeshserving": {
            component:       r.newModelMeshComponent(instance),
            managementState: instance.Spec.Components.ModelMeshServing.ManagementState,
        },
        "datasciencepipelines": {
            component:       r.newDSPComponent(instance),
            managementState: instance.Spec.Components.DataSciencePipelines.ManagementState,
        },
    }

    // Reconcile each component
    for name, comp := range components {
        componentLog := log.WithValues("component", name)

        if comp.managementState == dscv1.Managed {
            componentLog.Info("Reconciling managed component")
            if err := comp.component.ReconcileComponent(ctx, instance); err != nil {
                componentErrors[name] = err
                componentLog.Error(err, "Failed to reconcile component")
            } else {
                componentLog.Info("Component reconciled successfully")
            }
        } else if comp.managementState == dscv1.Removed {
            componentLog.Info("Removing component")
            if err := comp.component.Cleanup(ctx); err != nil {
                componentErrors[name] = err
                componentLog.Error(err, "Failed to remove component")
            } else {
                componentLog.Info("Component removed successfully")
            }
        }
    }

    return componentErrors
}
```

---

## 🧩 Component Interface Implementation

### Component Interface Definition
From `pkg/components/component.go`:

```go
// ComponentInterface defines the interface that all ODH components must implement
type ComponentInterface interface {
    // ReconcileComponent handles the main reconciliation logic for the component
    ReconcileComponent(ctx context.Context, owner metav1.Object) error

    // Cleanup handles the removal of component resources
    Cleanup(ctx context.Context) error

    // GetComponentName returns the name of the component
    GetComponentName() string

    // GetManagementState returns the current management state
    GetManagementState() dscv1.ManagementState

    // IsReady checks if the component is ready and healthy
    IsReady(ctx context.Context) (bool, error)

    // GetConditions returns the current conditions for the component
    GetConditions() []metav1.Condition

    // UpdateComponentCondition updates a specific condition
    UpdateComponentCondition(conditionType string, status metav1.ConditionStatus, reason, message string)
}
```

### Dashboard Component Implementation
From `pkg/components/dashboard/dashboard.go`:

```go
// Dashboard represents the dashboard component
type Dashboard struct {
    client.Client
    Scheme *runtime.Scheme
    Log    logr.Logger

    // Component configuration
    ManagementState dscv1.ManagementState
    DevFlags        []dscv1.DevFlag

    // Status tracking
    conditions []metav1.Condition
}

// GetComponentName returns the component name
func (d *Dashboard) GetComponentName() string {
    return "dashboard"
}

// GetManagementState returns the management state
func (d *Dashboard) GetManagementState() dscv1.ManagementState {
    return d.ManagementState
}

// ReconcileComponent implements the main reconciliation logic
func (d *Dashboard) ReconcileComponent(ctx context.Context, owner metav1.Object) error {
    log := d.Log.WithValues("component", d.GetComponentName())

    // Step 1: Check if component should be managed
    if d.ManagementState != dscv1.Managed {
        log.Info("Component not managed, skipping reconciliation")
        return nil
    }

    // Step 2: Generate manifests based on configuration
    manifests, err := d.generateManifests()
    if err != nil {
        d.UpdateComponentCondition(
            "Ready",
            metav1.ConditionFalse,
            "ManifestGenerationFailed",
            fmt.Sprintf("Failed to generate manifests: %v", err),
        )
        return err
    }

    // Step 3: Apply manifests
    for _, manifest := range manifests {
        if err := d.applyManifest(ctx, manifest, owner); err != nil {
            d.UpdateComponentCondition(
                "Ready",
                metav1.ConditionFalse,
                "ManifestApplyFailed",
                fmt.Sprintf("Failed to apply manifest %s: %v", manifest.GetName(), err),
            )
            return err
        }
    }

    // Step 4: Verify deployment status
    ready, err := d.IsReady(ctx)
    if err != nil {
        d.UpdateComponentCondition(
            "Ready",
            metav1.ConditionUnknown,
            "ReadinessCheckFailed",
            fmt.Sprintf("Failed to check readiness: %v", err),
        )
        return err
    }

    // Step 5: Update conditions based on readiness
    if ready {
        d.UpdateComponentCondition(
            "Ready",
            metav1.ConditionTrue,
            "ComponentReady",
            "Dashboard component is ready and available",
        )
    } else {
        d.UpdateComponentCondition(
            "Ready",
            metav1.ConditionFalse,
            "ComponentNotReady",
            "Dashboard component is not yet ready",
        )
    }

    log.Info("Component reconciliation completed", "ready", ready)
    return nil
}

// IsReady checks if the dashboard deployment is ready
func (d *Dashboard) IsReady(ctx context.Context) (bool, error) {
    // Check if dashboard deployment exists and is ready
    deployment := &appsv1.Deployment{}
    err := d.Get(ctx, types.NamespacedName{
        Name:      "odh-dashboard",
        Namespace: "opendatahub",
    }, deployment)

    if err != nil {
        if errors.IsNotFound(err) {
            return false, nil
        }
        return false, err
    }

    // Check deployment conditions
    for _, condition := range deployment.Status.Conditions {
        if condition.Type == appsv1.DeploymentAvailable &&
           condition.Status == corev1.ConditionTrue {
            return deployment.Status.ReadyReplicas == deployment.Status.Replicas, nil
        }
    }

    return false, nil
}

// Cleanup removes dashboard resources
func (d *Dashboard) Cleanup(ctx context.Context) error {
    log := d.Log.WithValues("component", d.GetComponentName())
    log.Info("Starting component cleanup")

    // List of resources to clean up
    resourcesToCleanup := []client.Object{
        &appsv1.Deployment{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "odh-dashboard",
                Namespace: "opendatahub",
            },
        },
        &corev1.Service{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "odh-dashboard",
                Namespace: "opendatahub",
            },
        },
        &corev1.ConfigMap{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "odh-dashboard-config",
                Namespace: "opendatahub",
            },
        },
    }

    // Delete each resource
    for _, resource := range resourcesToCleanup {
        err := d.Delete(ctx, resource)
        if err != nil && !errors.IsNotFound(err) {
            log.Error(err, "Failed to delete resource", "resource", resource.GetName())
            return err
        }
        log.Info("Resource deleted", "resource", resource.GetName())
    }

    d.UpdateComponentCondition(
        "Ready",
        metav1.ConditionFalse,
        "ComponentRemoved",
        "Dashboard component has been removed",
    )

    log.Info("Component cleanup completed")
    return nil
}
```

---

## ⚡ Action-Based Architecture

### Base Action Interface
From `pkg/controller/actions/action.go`:

```go
// Action represents a discrete operation that can be performed on a component
type Action interface {
    // Execute performs the action
    Execute(ctx context.Context) error

    // GetName returns the action name for logging
    GetName() string

    // GetComponent returns the component this action operates on
    GetComponent() string
}

// BaseAction provides common functionality for all actions
type BaseAction struct {
    Client        client.Client
    ComponentName string
    Owner         metav1.Object
    Log           logr.Logger
}

func (a *BaseAction) GetComponent() string {
    return a.ComponentName
}
```

### Install Action Implementation
```go
// InstallAction installs component resources
type InstallAction struct {
    BaseAction
    Manifests []unstructured.Unstructured
}

func (a *InstallAction) GetName() string {
    return fmt.Sprintf("install-%s", a.ComponentName)
}

func (a *InstallAction) Execute(ctx context.Context) error {
    log := a.Log.WithValues("action", a.GetName())
    log.Info("Executing install action", "manifests", len(a.Manifests))

    for i, manifest := range a.Manifests {
        // Set owner reference for garbage collection
        if err := controllerutil.SetControllerReference(a.Owner, &manifest, a.Client.Scheme()); err != nil {
            return fmt.Errorf("failed to set owner reference for manifest %d: %w", i, err)
        }

        // Apply the manifest
        if err := a.applyManifest(ctx, manifest); err != nil {
            return fmt.Errorf("failed to apply manifest %s: %w", manifest.GetName(), err)
        }

        log.Info("Manifest applied successfully", "name", manifest.GetName(), "kind", manifest.GetKind())
    }

    log.Info("Install action completed successfully")
    return nil
}

func (a *InstallAction) applyManifest(ctx context.Context, manifest unstructured.Unstructured) error {
    // Check if resource already exists
    existing := &unstructured.Unstructured{}
    existing.SetGroupVersionKind(manifest.GroupVersionKind())

    err := a.Client.Get(ctx, types.NamespacedName{
        Name:      manifest.GetName(),
        Namespace: manifest.GetNamespace(),
    }, existing)

    if err != nil {
        if errors.IsNotFound(err) {
            // Create new resource
            return a.Client.Create(ctx, &manifest)
        }
        return err
    }

    // Update existing resource
    manifest.SetResourceVersion(existing.GetResourceVersion())
    return a.Client.Update(ctx, &manifest)
}
```

### Update Action Implementation
```go
// UpdateAction updates existing component resources
type UpdateAction struct {
    BaseAction
    ExistingManifest unstructured.Unstructured
    DesiredManifest  unstructured.Unstructured
}

func (a *UpdateAction) GetName() string {
    return fmt.Sprintf("update-%s", a.ComponentName)
}

func (a *UpdateAction) Execute(ctx context.Context) error {
    log := a.Log.WithValues("action", a.GetName())

    // Check if update is needed
    if !a.needsUpdate() {
        log.Info("No update needed")
        return nil
    }

    log.Info("Updating manifest", "name", a.ExistingManifest.GetName())

    // Preserve important fields
    a.DesiredManifest.SetResourceVersion(a.ExistingManifest.GetResourceVersion())
    a.DesiredManifest.SetUID(a.ExistingManifest.GetUID())

    // Perform update
    err := a.Client.Update(ctx, &a.DesiredManifest)
    if err != nil {
        return fmt.Errorf("failed to update manifest: %w", err)
    }

    log.Info("Update action completed successfully")
    return nil
}

func (a *UpdateAction) needsUpdate() bool {
    // Compare spec sections to determine if update is needed
    existingSpec, _, _ := unstructured.NestedMap(a.ExistingManifest.Object, "spec")
    desiredSpec, _, _ := unstructured.NestedMap(a.DesiredManifest.Object, "spec")

    return !reflect.DeepEqual(existingSpec, desiredSpec)
}
```

---

## 🔄 Status Management Patterns

### Status Update Implementation
From `controllers/datasciencecluster_controller.go`:

```go
func (r *DSCReconciler) updateStatus(ctx context.Context, instance *dscv1.DataScienceCluster, componentErrors map[string]error, log logr.Logger) error {
    // Collect component conditions
    allConditions := []metav1.Condition{}
    allReady := true
    hasErrors := len(componentErrors) > 0

    // Get conditions from each component
    for name, comp := range r.getComponents(instance) {
        if comp.GetManagementState() == dscv1.Managed {
            componentConditions := comp.GetConditions()

            // Add component prefix to condition types
            for _, condition := range componentConditions {
                prefixedCondition := condition.DeepCopy()
                prefixedCondition.Type = fmt.Sprintf("%s%s", name, condition.Type)
                allConditions = append(allConditions, *prefixedCondition)

                // Check if component is ready
                if condition.Type == "Ready" && condition.Status != metav1.ConditionTrue {
                    allReady = false
                }
            }
        }
    }

    // Determine overall phase
    var phase dscv1.DSCPhase
    var readyCondition metav1.Condition

    if hasErrors {
        phase = dscv1.PhaseError
        readyCondition = metav1.Condition{
            Type:               "Ready",
            Status:             metav1.ConditionFalse,
            Reason:             "ComponentErrors",
            Message:            fmt.Sprintf("Errors in %d components", len(componentErrors)),
            LastTransitionTime: metav1.Now(),
        }
    } else if allReady {
        phase = dscv1.PhaseReady
        readyCondition = metav1.Condition{
            Type:               "Ready",
            Status:             metav1.ConditionTrue,
            Reason:             "AllComponentsReady",
            Message:            "All managed components are ready",
            LastTransitionTime: metav1.Now(),
        }
    } else {
        phase = dscv1.PhaseProgressing
        readyCondition = metav1.Condition{
            Type:               "Ready",
            Status:             metav1.ConditionFalse,
            Reason:             "ComponentsProgressing",
            Message:            "Some components are still progressing",
            LastTransitionTime: metav1.Now(),
        }
    }

    // Update instance status
    instance.Status.Phase = phase
    instance.Status.Conditions = append(allConditions, readyCondition)
    instance.Status.ObservedGeneration = instance.Generation

    // Perform status update
    err := r.Status().Update(ctx, instance)
    if err != nil {
        return fmt.Errorf("failed to update status: %w", err)
    }

    log.Info("Status updated", "phase", phase, "conditions", len(allConditions))
    return nil
}
```

---

## 🎯 Error Handling and Resilience

### Finalizer Pattern
```go
const finalizerName = "datasciencecluster.opendatahub.io/finalizer"

func (r *DSCReconciler) handleDeletion(ctx context.Context, instance *dscv1.DataScienceCluster, log logr.Logger) (ctrl.Result, error) {
    log.Info("Handling DataScienceCluster deletion")

    // Get all components for cleanup
    components := r.getComponents(instance)

    // Clean up components in reverse order
    componentNames := make([]string, 0, len(components))
    for name := range components {
        componentNames = append(componentNames, name)
    }

    // Reverse the slice for cleanup
    for i := len(componentNames) - 1; i >= 0; i-- {
        name := componentNames[i]
        component := components[name]

        log.Info("Cleaning up component", "component", name)
        if err := component.Cleanup(ctx); err != nil {
            log.Error(err, "Failed to cleanup component", "component", name)
            return ctrl.Result{RequeueAfter: time.Minute}, err
        }
    }

    // Remove finalizer
    controllerutil.RemoveFinalizer(instance, finalizerName)
    err := r.Update(ctx, instance)
    if err != nil {
        log.Error(err, "Failed to remove finalizer")
        return ctrl.Result{}, err
    }

    log.Info("DataScienceCluster deletion completed")
    return ctrl.Result{}, nil
}
```

### Retry Logic Pattern
```go
func (r *DSCReconciler) reconcileComponentWithRetry(ctx context.Context, component components.ComponentInterface, owner metav1.Object) error {
    const maxRetries = 3
    const baseDelay = time.Second

    var lastErr error
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := component.ReconcileComponent(ctx, owner)
        if err == nil {
            return nil
        }

        lastErr = err

        // Check if error is retryable
        if !isRetryableError(err) {
            return err
        }

        // Exponential backoff
        delay := baseDelay * time.Duration(1<<attempt)
        time.Sleep(delay)

        r.Log.Info("Retrying component reconciliation",
            "component", component.GetComponentName(),
            "attempt", attempt+1,
            "delay", delay,
        )
    }

    return fmt.Errorf("component %s failed after %d attempts: %w",
        component.GetComponentName(), maxRetries, lastErr)
}

func isRetryableError(err error) bool {
    // Define which errors are worth retrying
    if errors.IsServiceUnavailable(err) ||
       errors.IsTimeout(err) ||
       errors.IsServerTimeout(err) {
        return true
    }

    // Check for temporary network issues
    if strings.Contains(err.Error(), "connection refused") ||
       strings.Contains(err.Error(), "timeout") {
        return true
    }

    return false
}
```

---

## 🎯 Summary

These examples demonstrate ODH's sophisticated controller architecture:

1. **Hierarchical Design**: Main controller orchestrates component controllers
2. **Interface Consistency**: All components implement the same interface
3. **Action Modularity**: Discrete actions for install, update, delete operations
4. **Robust Status Management**: Comprehensive condition tracking and aggregation
5. **Error Resilience**: Proper finalizer handling and retry logic
6. **Resource Management**: Owner references for garbage collection

This architecture enables ODH to manage a complex multi-component platform while maintaining consistency, reliability, and extensibility.