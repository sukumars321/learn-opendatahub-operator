# Day 11: Advanced Component Management Patterns - Practical Examples

**Real ODH Codebase Examples**

This document contains actual code examples from the OpenDataHub operator that demonstrate the advanced component management patterns covered in today's study guide.

---

## Example 1: Advanced Component Interface Implementation

### ODH Component Registry Pattern

The ODH operator implements a sophisticated component registry system that manages multiple components dynamically:

```go
// From controllers/datasciencecluster/datasciencecluster_controller.go
func (r *DataScienceClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Component registry approach
    cr := componentregistry.ComponentRegistry{}

    // Each component is registered with its specific reconciler
    cr.ForEach(func(c component.ComponentInterface) error {
        return c.ReconcileComponent(ctx, instance, r.Client, r.Log.WithValues("component", c.GetComponentName()))
    })

    return ctrl.Result{}, nil
}
```

### Real Component Interface Implementation

From the Dashboard component controller:

```go
// From controllers/datasciencecluster/dashboard/dashboard_controller.go
type DashboardReconciler struct {
    client.Client
    Scheme *runtime.Scheme
    Log    logr.Logger
}

func (r *DashboardReconciler) ReconcileComponent(ctx context.Context,
    instance *dsciv1.DataScienceCluster) error {

    enabled := instance.Spec.Components.Dashboard.ManagementState == operatorv1.Managed
    if !enabled {
        return r.cleanupDashboardResources(ctx, instance)
    }

    // Multi-step reconciliation with action pipeline
    return r.reconcileDashboard(ctx, instance)
}
```

---

## Example 2: Sophisticated Resource Management

### Dynamic Resource Ownership

ODH implements dynamic resource watching and ownership:

```go
// From dashboard reconciler setup
func (r *DashboardReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&dsciv1.DataScienceCluster{}).
        // Static resource ownership
        Owns(&corev1.ConfigMap{}).
        Owns(&corev1.Secret{}).
        Owns(&rbacv1.ClusterRoleBinding{}).
        Owns(&rbacv1.ClusterRole{}).
        // Dynamic GVK watching for extensibility
        OwnsGVK(gvk.DashboardAcceleratorProfile, reconciler.Dynamic()).
        OwnsGVK(gvk.OdhApplication, reconciler.Dynamic()).
        Complete(r)
}
```

### Advanced Action Pipeline

```go
// From dashboard reconciliation logic
func (r *DashboardReconciler) reconcileDashboard(ctx context.Context,
    instance *dsciv1.DataScienceCluster) error {

    // Comprehensive action pipeline
    pipeline := reconciler.ReconcilerFor(mgr, &componentApi.Dashboard{}).
        WithAction(initialize).
        WithAction(setKustomizedParams).
        WithAction(configureDependencies).
        WithAction(kustomize.NewAction()).
        WithAction(customizeResources).
        WithAction(deploy.NewAction()).
        WithAction(deployments.NewAction()).
        WithAction(reconcileHardwareProfiles).
        WithAction(updateStatus).
        WithAction(gc.NewAction()) // Garbage collection

    return pipeline.Reconcile(ctx)
}
```

---

## Example 3: Configuration Management in Practice

### Hierarchical Configuration Resolution

```go
// From component configuration management
type DashboardConfig struct {
    // Base configuration
    ManagementState operatorv1.ManagementState `json:"managementState,omitempty"`

    // Component-specific configuration
    DashboardConfig componentv1.DashboardConfig `json:"dashboardConfig,omitempty"`
}

func (r *DashboardReconciler) resolveConfiguration(ctx context.Context,
    instance *dsciv1.DataScienceCluster) (*DashboardConfig, error) {

    config := &DashboardConfig{}

    // Apply default configuration
    applyDefaults(config)

    // Apply platform-specific overrides
    if err := r.applyPlatformConfig(ctx, config); err != nil {
        return nil, err
    }

    // Apply user-specified configuration
    if instance.Spec.Components.Dashboard.DashboardConfig != nil {
        config.DashboardConfig = *instance.Spec.Components.Dashboard.DashboardConfig
    }

    return config, nil
}
```

### Configuration Validation Pattern

```go
func (r *DashboardReconciler) validateConfiguration(config *DashboardConfig) error {
    if config.ManagementState == operatorv1.Managed {
        // Validate required configuration for managed state
        if config.DashboardConfig.EnableOAuthProxy == nil {
            return fmt.Errorf("OAuth proxy configuration required when dashboard is managed")
        }
    }

    // Component-specific validation
    return r.validateDashboardSpecificConfig(&config.DashboardConfig)
}
```

---

## Example 4: Advanced Status Management

### Comprehensive Status Structure

```go
// From api/datasciencecluster/v1/datasciencecluster_types.go
type DataScienceClusterStatus struct {
    Phase      string                 `json:"phase,omitempty"`
    Conditions []metav1.Condition     `json:"conditions,omitempty"`

    // Component-specific status aggregation
    InstalledComponents map[string]bool `json:"installedComponents,omitempty"`
    ErrorMessage        string          `json:"errorMessage,omitempty"`
}
```

### Status Update Implementation

```go
func (r *DashboardReconciler) updateComponentStatus(ctx context.Context,
    instance *dsciv1.DataScienceCluster, err error) error {

    // Component-level status conditions
    conditions := []metav1.Condition{}

    if err != nil {
        conditions = append(conditions, metav1.Condition{
            Type:    "DashboardReady",
            Status:  metav1.ConditionFalse,
            Reason:  "ReconciliationFailed",
            Message: err.Error(),
            LastTransitionTime: metav1.Now(),
        })
    } else {
        conditions = append(conditions, metav1.Condition{
            Type:    "DashboardReady",
            Status:  metav1.ConditionTrue,
            Reason:  "ReconciliationSucceeded",
            Message: "Dashboard component successfully reconciled",
            LastTransitionTime: metav1.Now(),
        })
    }

    return r.updateStatusConditions(ctx, instance, conditions)
}
```

---

## Example 5: Dependency Management Patterns

### Component Dependency Checking

```go
func (r *DashboardReconciler) checkDependencies(ctx context.Context,
    instance *dsciv1.DataScienceCluster) error {

    // Check for required components
    if instance.Spec.Components.Workbenches.ManagementState != operatorv1.Managed {
        return fmt.Errorf("dashboard requires workbenches component to be enabled")
    }

    // Check optional dependencies
    if r.isServiceMeshEnabled(ctx, instance) {
        if err := r.validateServiceMeshIntegration(ctx); err != nil {
            r.Log.Info("Service mesh integration not available, continuing without it",
                "reason", err.Error())
        }
    }

    return nil
}
```

### Conditional Component Installation

```go
func (r *DashboardReconciler) reconcileOptionalFeatures(ctx context.Context,
    instance *dsciv1.DataScienceCluster) error {

    // Conditionally enable features based on other components
    if r.isModelServingEnabled(instance) {
        if err := r.enableModelServingIntegration(ctx); err != nil {
            return fmt.Errorf("failed to enable model serving integration: %w", err)
        }
    }

    if r.isNotebooksEnabled(instance) {
        if err := r.enableNotebookIntegration(ctx); err != nil {
            // Soft failure - log but don't fail reconciliation
            r.Log.Error(err, "Failed to enable notebook integration, continuing without it")
        }
    }

    return nil
}
```

---

## Example 6: Error Handling and Recovery

### Sophisticated Error Classification

```go
func (r *DashboardReconciler) handleReconciliationError(ctx context.Context,
    err error, instance *dsciv1.DataScienceCluster) (ctrl.Result, error) {

    switch {
    case apierrors.IsNotFound(err):
        // Resource not found - likely needs creation
        r.Log.Info("Resource not found, will be created on next reconciliation")
        return ctrl.Result{RequeueAfter: time.Second * 10}, nil

    case apierrors.IsConflict(err):
        // Resource conflict - retry quickly
        r.Log.Info("Resource conflict detected, retrying")
        return ctrl.Result{RequeueAfter: time.Second * 2}, nil

    case isTransientError(err):
        // Temporary failure - exponential backoff
        return ctrl.Result{RequeueAfter: time.Second * 30}, nil

    default:
        // Permanent failure - log and update status
        r.Log.Error(err, "Permanent error in dashboard reconciliation")
        if statusErr := r.updateComponentStatus(ctx, instance, err); statusErr != nil {
            r.Log.Error(statusErr, "Failed to update status")
        }
        return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
    }
}
```

### Circuit Breaker Implementation

```go
type ComponentCircuitBreaker struct {
    failures    int
    lastFailure time.Time
    threshold   int
    timeout     time.Duration
    state       CircuitState
}

func (cb *ComponentCircuitBreaker) Execute(operation func() error) error {
    if cb.state == OpenState {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = HalfOpenState
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }

    err := operation()
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()

        if cb.failures >= cb.threshold {
            cb.state = OpenState
        }
        return err
    }

    // Success - reset circuit breaker
    cb.failures = 0
    cb.state = ClosedState
    return nil
}
```

---

## Example 7: Resource Cleanup and Garbage Collection

### Intelligent Resource Cleanup

```go
func (r *DashboardReconciler) cleanupDashboardResources(ctx context.Context,
    instance *dsciv1.DataScienceCluster) error {

    // Use garbage collection action for intelligent cleanup
    gc := gc.NewAction()

    // Define resources to clean up
    gc.WithResourcesToDelete(
        &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "odh-dashboard-config"}},
        &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "dashboard-oauth-config"}},
        &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "dashboard-cluster-role-binding"}},
    )

    return gc.Execute(ctx)
}
```

---

## Key Patterns Demonstrated

### 1. Component Registry Pattern
- Central registry manages all components
- Dynamic component discovery and registration
- Consistent reconciliation interface across components

### 2. Action Pipeline Pattern
- Modular reconciliation steps
- Reusable actions across components
- Clear separation of concerns

### 3. Configuration Hierarchy
- Default → Platform → User configuration layers
- Validation at multiple levels
- Runtime configuration updates

### 4. Status Aggregation
- Component-level conditions roll up to cluster level
- Consistent condition types and messages
- Intelligent status transitions

### 5. Dependency Management
- Hard and soft dependency checking
- Graceful degradation for optional dependencies
- Conditional feature enablement

### 6. Error Recovery
- Error classification for appropriate handling
- Circuit breaker pattern for repeated failures
- Exponential backoff for transient errors

These real-world examples from the ODH codebase demonstrate how advanced component management patterns are implemented in production Kubernetes operators. They provide a solid foundation for building robust, scalable operators that can manage complex, interdependent systems effectively.