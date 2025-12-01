# Day 12: Reconciler Implementation - Offline Exercises

## Exercise Time: 15 minutes

These exercises focus on understanding reconciler patterns through code analysis and theoretical implementation without requiring a running Kubernetes cluster.

---

## Exercise 1: Reconciler Pattern Analysis (5 minutes)

### Goal
Analyze different reconciler return patterns and their implications.

### Reconciler Return Pattern Analysis

Study the following reconciler code snippets and identify the pattern and use case for each:

```go
// Pattern A
func (r *MyReconciler) ReconcileA(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ... reconciliation logic ...
    if err := r.ensureDeployment(ctx, instance); err != nil {
        return ctrl.Result{}, err
    }
    return ctrl.Result{}, nil
}

// Pattern B
func (r *MyReconciler) ReconcileB(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ... reconciliation logic ...
    if !r.hasRequiredSecret(ctx, instance) {
        return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
    }
    return ctrl.Result{}, nil
}

// Pattern C
func (r *MyReconciler) ReconcileC(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ... reconciliation logic ...
    if !controllerutil.ContainsFinalizer(instance, myFinalizer) {
        controllerutil.AddFinalizer(instance, myFinalizer)
        if err := r.Update(ctx, instance); err != nil {
            return ctrl.Result{}, err
        }
        return ctrl.Result{Requeue: true}, nil
    }
    return ctrl.Result{}, nil
}

// Pattern D
func (r *MyReconciler) ReconcileD(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ... reconciliation logic ...
    r.updateStatus(ctx, instance, "Ready", metav1.ConditionTrue)
    return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
}
```

### Analysis Questions
For each pattern, identify:

1. **Pattern A**: `return ctrl.Result{}, err`
   - When is this used?
   - What happens on error?
   - What happens on success?

2. **Pattern B**: `return ctrl.Result{RequeueAfter: time.Minute * 5}, nil`
   - Why the delay?
   - When is this pattern appropriate?
   - What's the difference from returning an error?

3. **Pattern C**: `return ctrl.Result{Requeue: true}, nil`
   - Why immediate requeue?
   - When is this better than RequeueAfter?
   - What just happened in the reconciler?

4. **Pattern D**: `return ctrl.Result{RequeueAfter: time.Minute * 10}, nil`
   - Why periodic reconciliation?
   - What's the trade-off with frequency?
   - When is this pattern essential?

### Answers
<details>
<summary>Click to reveal answers</summary>

1. **Pattern A**: Standard error handling
   - Used for transient errors that should be retried
   - On error: Controller-runtime applies exponential backoff
   - On success: No requeue, waits for next event

2. **Pattern B**: Configuration dependency waiting
   - Delay allows time for external dependencies (secrets, configmaps)
   - Appropriate when waiting for user/admin action
   - Avoids spamming logs with repeated failures

3. **Pattern C**: Immediate requeue after resource update
   - Resource was modified (finalizer added), need fresh reconciliation
   - Immediate because the change is internal and ready
   - Just updated the resource metadata

4. **Pattern D**: Periodic reconciliation for drift detection
   - Ensures operator detects configuration drift
   - Trade-off: Higher frequency = more responsive but more load
   - Essential for operators managing external systems
</details>

---

## Exercise 2: Reconciler State Machine Design (5 minutes)

### Goal
Design a state machine for a complex reconciler managing multiple components.

### Scenario
You're building an operator for a "WebApp" resource that manages:
- Deployment (app containers)
- Service (networking)
- Ingress (external access)
- ConfigMap (configuration)
- Secret (credentials)

### Design Challenge
Complete this reconciler state machine:

```go
type WebAppPhase string

const (
    WebAppPhasePending     WebAppPhase = "Pending"
    WebAppPhaseCreating    WebAppPhase = "Creating"
    WebAppPhaseReady       WebAppPhase = "Ready"
    WebAppPhaseUpdating    WebAppPhase = "Updating"
    WebAppPhaseDeleting    WebAppPhase = "Deleting"
    WebAppPhaseFailed      WebAppPhase = "Failed"
)

func (r *WebAppReconciler) determineDesiredPhase(ctx context.Context, webapp *WebApp) (WebAppPhase, error) {
    // TODO: Complete this state machine logic

    // Consider:
    // - Is the resource being deleted?
    // - Are dependencies ready (secret, configmap)?
    // - Is this a new resource?
    // - Are there spec changes requiring updates?
    // - Are components healthy?

    return WebAppPhasePending, nil
}
```

### State Transition Rules
Fill in the state transition matrix:

| Current Phase | Condition | Next Phase | Action Required |
|---------------|-----------|------------|----------------|
| Pending | DeletionTimestamp != nil | ? | ? |
| Pending | Dependencies missing | ? | ? |
| Pending | Dependencies ready | ? | ? |
| Creating | All components created | ? | ? |
| Creating | Component creation failed | ? | ? |
| Ready | Spec changed | ? | ? |
| Ready | Component unhealthy | ? | ? |
| Updating | Update completed | ? | ? |
| Failed | User fixed config | ? | ? |

### Implementation Template
```go
func (r *WebAppReconciler) determineDesiredPhase(ctx context.Context, webapp *WebApp) (WebAppPhase, error) {
    // Step 1: Check for deletion
    if webapp.DeletionTimestamp != nil {
        return WebAppPhaseDeleting, nil
    }

    // Step 2: Check dependencies
    if !r.dependenciesReady(ctx, webapp) {
        return _________, nil  // Fill in the blank
    }

    // Step 3: Check if new resource
    if webapp.Status.Phase == "" {
        return _________, nil  // Fill in the blank
    }

    // Step 4: Check for spec changes
    if r.specChanged(webapp) {
        return _________, nil  // Fill in the blank
    }

    // Step 5: Check component health
    if !r.allComponentsHealthy(ctx, webapp) {
        return _________, nil  // Fill in the blank
    }

    // Step 6: Default to current state if stable
    return webapp.Status.Phase, nil
}
```

### Solution
<details>
<summary>Click to reveal solution</summary>

State Transition Matrix:
| Current Phase | Condition | Next Phase | Action Required |
|---------------|-----------|------------|----------------|
| Pending | DeletionTimestamp != nil | Deleting | Run cleanup |
| Pending | Dependencies missing | Pending | Wait/requeue |
| Pending | Dependencies ready | Creating | Create components |
| Creating | All components created | Ready | Update status |
| Creating | Component creation failed | Failed | Log error, requeue |
| Ready | Spec changed | Updating | Update components |
| Ready | Component unhealthy | Failed | Investigate, requeue |
| Updating | Update completed | Ready | Update status |
| Failed | User fixed config | Pending | Retry reconciliation |

Implementation:
```go
// Step 2: return WebAppPhasePending, nil
// Step 3: return WebAppPhaseCreating, nil
// Step 4: return WebAppPhaseUpdating, nil
// Step 5: return WebAppPhaseFailed, nil
```
</details>

---

## Exercise 3: Error Handling Strategy Design (5 minutes)

### Goal
Design comprehensive error handling strategies for different types of reconciler failures.

### Error Scenarios
Classify each error scenario and design the appropriate response:

```go
// Scenario 1: Network timeout creating deployment
err1 := errors.NewTimeoutError("timeout creating deployment", 30)

// Scenario 2: Invalid image name in spec
err2 := errors.NewInvalid(schema.GroupKind{}, "webapp", field.ErrorList{
    field.Invalid(field.NewPath("spec.image"), "invalid:image:name", "invalid format"),
})

// Scenario 3: Insufficient RBAC permissions
err3 := errors.NewForbidden(schema.GroupResource{}, "deployment", fmt.Errorf("cannot create deployment"))

// Scenario 4: Secret dependency not found
err4 := errors.NewNotFound(schema.GroupResource{}, "required-secret")

// Scenario 5: API server temporarily unavailable
err5 := errors.NewServiceUnavailable("API server overloaded")
```

### Error Handling Design
For each scenario, complete the error handling:

```go
func (r *WebAppReconciler) handleReconcileError(err error, webapp *WebApp) (ctrl.Result, error) {
    switch {
    case errors.IsTimeout(err):
        // Scenario 1: Network timeout
        r.setCondition(webapp, "Ready", metav1.ConditionFalse, "NetworkTimeout", err.Error())
        return _________, _________  // Fill in appropriate response

    case errors.IsInvalid(err):
        // Scenario 2: Invalid configuration
        r.setCondition(webapp, "Ready", metav1.ConditionFalse, "InvalidConfiguration", err.Error())
        return _________, _________  // Fill in appropriate response

    case errors.IsForbidden(err):
        // Scenario 3: RBAC issue
        r.setCondition(webapp, "Ready", metav1.ConditionFalse, "InsufficientPermissions", err.Error())
        return _________, _________  // Fill in appropriate response

    case errors.IsNotFound(err):
        // Scenario 4: Missing dependency
        r.setCondition(webapp, "Ready", metav1.ConditionFalse, "MissingDependency", err.Error())
        return _________, _________  // Fill in appropriate response

    case errors.IsServiceUnavailable(err):
        // Scenario 5: API server issue
        return _________, _________  // Fill in appropriate response

    default:
        // Unknown error
        return _________, _________  // Fill in appropriate response
    }
}
```

### Error Handling Questions
1. Which errors should be retried immediately?
2. Which errors should have delayed retry?
3. Which errors should not be retried automatically?
4. How should each error type be communicated to users?

### Solution
<details>
<summary>Click to reveal solution</summary>

```go
case errors.IsTimeout(err):
    // Network timeout - retry with exponential backoff
    return ctrl.Result{}, err

case errors.IsInvalid(err):
    // Invalid config - don't retry until user fixes
    return ctrl.Result{RequeueAfter: time.Minute * 10}, nil

case errors.IsForbidden(err):
    // RBAC issue - don't spam, retry occasionally
    return ctrl.Result{RequeueAfter: time.Minute * 15}, nil

case errors.IsNotFound(err):
    // Missing dependency - retry with delay
    return ctrl.Result{RequeueAfter: time.Minute * 2}, nil

case errors.IsServiceUnavailable(err):
    // API server issue - let controller-runtime handle backoff
    return ctrl.Result{}, err

default:
    // Unknown error - default exponential backoff
    return ctrl.Result{}, err
```

Error Type Classification:
- **Immediate retry**: Transient network issues, API server temporary problems
- **Delayed retry**: Missing dependencies, RBAC issues
- **No automatic retry**: Invalid configuration (user must fix)
- **User communication**: All errors should update status conditions with clear messages
</details>

---

## Exercise Summary

### Key Concepts Practiced
1. **Reconciler Return Patterns**: Understanding when to use different ctrl.Result patterns
2. **State Machine Design**: Managing complex reconciler state transitions
3. **Error Handling Strategy**: Categorizing and responding to different error types appropriately

### Reconciler Design Principles Learned
- **Idempotent Operations**: Same input always produces same output
- **Error Classification**: Different errors require different retry strategies
- **State Management**: Clear state machines prevent reconciliation confusion
- **User Communication**: Status conditions and events provide operational visibility

### Best Practices Identified
- Use exponential backoff for transient errors
- Use delayed requeue for configuration issues
- Use immediate requeue after resource updates
- Use periodic requeue for drift detection
- Always update status conditions for user visibility

### Pattern Library Built
```go
// Success patterns
return ctrl.Result{}, nil                           // Complete, wait for events
return ctrl.Result{RequeueAfter: duration}, nil    // Periodic reconciliation

// Error patterns
return ctrl.Result{}, err                          // Transient error, backoff
return ctrl.Result{RequeueAfter: duration}, nil   // Config error, delayed retry

// Update patterns
return ctrl.Result{Requeue: true}, nil            // Resource updated, immediate retry
```

## Next Steps
- Day 13: Event Watching and Filtering
- Continue analyzing ODH operator reconciler implementations
- Practice implementing these patterns in real reconciler code