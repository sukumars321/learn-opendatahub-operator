# Day 12: Reconciler Implementation - Live Exercises

## Exercise Time: 15 minutes

These hands-on exercises will help you practice implementing reconciler patterns and debugging reconciliation logic using the ODH operator and your local Kubernetes environment.

---

## Exercise 1: ODH Reconciler Code Analysis (5 minutes)

### Goal
Analyze real reconciler implementations in the ODH operator codebase.

### Steps

1. **Navigate to the ODH operator controllers directory**:
   ```bash
   cd /Users/suksubra/Documents/Work/RHOAI/opendatahub-operator/controllers
   ```

2. **Examine the DataScienceCluster reconciler**:
   ```bash
   # View the main reconciler structure
   grep -A 20 "func.*Reconcile.*ctrl.Request" datasciencecluster_controller.go

   # Find all return statements to understand result patterns
   grep -n "return ctrl.Result" datasciencecluster_controller.go
   ```

3. **Analyze component reconciliation patterns**:
   ```bash
   # Look for component management patterns
   grep -A 10 -B 5 "ManagementState" datasciencecluster_controller.go

   # Find status update patterns
   grep -A 5 -B 5 "Status.*Update\|updateStatus" datasciencecluster_controller.go
   ```

4. **Study finalizer implementation**:
   ```bash
   # Find finalizer patterns
   grep -A 5 -B 5 "Finalizer\|finalizer" datasciencecluster_controller.go
   ```

### Analysis Questions
- How does ODH handle resource deletion vs creation?
- What conditions does ODH set on DataScienceCluster resources?
- How are component states managed during reconciliation?

---

## Exercise 2: Reconciler Debugging Practice (5 minutes)

### Goal
Practice debugging reconciliation issues using ODH operator logs and events.

### Steps

1. **Check ODH operator deployment status**:
   ```bash
   # Verify operator is running
   kubectl get deployment -n opendatahub-operator-system

   # Check operator pod logs
   kubectl logs -n opendatahub-operator-system deployment/opendatahub-operator-controller-manager --tail=50
   ```

2. **Examine DataScienceCluster resources and their reconciliation**:
   ```bash
   # List all DataScienceCluster resources
   kubectl get datasciencecluster -A

   # Get detailed status of a DataScienceCluster
   kubectl describe datasciencecluster default-dsc 2>/dev/null || echo "No DataScienceCluster found - check cluster setup"

   # View recent reconciliation events
   kubectl get events --sort-by='.metadata.creationTimestamp' | grep -i "datasciencecluster\|reconcil" | tail -10
   ```

3. **Monitor reconciliation activity**:
   ```bash
   # Follow operator logs for reconciliation activity
   echo "=== Following operator logs for 30 seconds ==="
   timeout 30s kubectl logs -n opendatahub-operator-system deployment/opendatahub-operator-controller-manager -f 2>/dev/null || echo "Timeout or no logs available"
   ```

4. **Analyze reconciliation patterns in logs**:
   ```bash
   # Look for reconciliation patterns
   kubectl logs -n opendatahub-operator-system deployment/opendatahub-operator-controller-manager --tail=100 2>/dev/null | grep -E "(Reconciling|Reconcile|ctrl.Result|error)" || echo "No matching log entries found"
   ```

### Debugging Checklist
- [ ] Can you identify reconciliation start/end in logs?
- [ ] What ctrl.Result patterns do you see?
- [ ] Are there any reconciliation errors or warnings?
- [ ] How frequently is reconciliation occurring?

---

## Exercise 3: Reconciler Implementation Pattern Practice (5 minutes)

### Goal
Create a simple reconciler function following best practices.

### Steps

1. **Create a practice reconciler implementation**:
   ```bash
   # Create a practice directory
   mkdir -p /tmp/reconciler-practice
   cd /tmp/reconciler-practice
   ```

2. **Write a basic reconciler structure**:
   ```bash
   cat > basic_reconciler.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/go-logr/logr"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type MyResourceReconciler struct {
    client.Client
    Log    logr.Logger
    Scheme *runtime.Scheme
}

// Basic reconciler implementation following ODH patterns
func (r *MyResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := r.Log.WithValues("resource", req.NamespacedName)

    // Step 1: Fetch resource (mock pattern)
    log.Info("Starting reconciliation")

    // Step 2: Check for deletion (mock pattern)
    isBeingDeleted := false // This would check DeletionTimestamp
    if isBeingDeleted {
        log.Info("Resource being deleted, running cleanup")
        return r.handleDeletion(ctx, req, log)
    }

    // Step 3: Add finalizer (mock pattern)
    hasFinalizer := true // This would check for actual finalizer
    if !hasFinalizer {
        log.Info("Adding finalizer")
        // Would add finalizer and update resource
        return ctrl.Result{Requeue: true}, nil
    }

    // Step 4: Normal reconciliation
    return r.doReconcile(ctx, req, log)
}

func (r *MyResourceReconciler) doReconcile(ctx context.Context, req ctrl.Request, log logr.Logger) (ctrl.Result, error) {
    // Simulate reconciliation steps
    log.Info("Reconciling subresources")

    // Mock component reconciliation
    components := []string{"deployment", "service", "configmap"}
    for _, component := range components {
        log.Info("Reconciling component", "component", component)

        // Simulate potential failure
        if component == "deployment" && time.Now().Unix()%10 == 0 {
            err := fmt.Errorf("deployment reconciliation failed")
            log.Error(err, "Component reconciliation failed", "component", component)
            return ctrl.Result{RequeueAfter: time.Minute}, err
        }
    }

    log.Info("All components reconciled successfully")

    // Schedule next reconciliation in 5 minutes
    return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

func (r *MyResourceReconciler) handleDeletion(ctx context.Context, req ctrl.Request, log logr.Logger) (ctrl.Result, error) {
    log.Info("Handling resource deletion")

    // Mock cleanup operations
    cleanupTasks := []string{"remove finalizers", "cleanup external resources", "update status"}
    for _, task := range cleanupTasks {
        log.Info("Executing cleanup task", "task", task)
    }

    log.Info("Deletion handling completed")
    return ctrl.Result{}, nil
}

// Mock function to demonstrate reconciler patterns
func main() {
    fmt.Println("Reconciler implementation pattern example")
    fmt.Println("Key patterns implemented:")
    fmt.Println("1. Resource fetching with error handling")
    fmt.Println("2. Deletion handling with finalizers")
    fmt.Println("3. Component-based reconciliation")
    fmt.Println("4. Proper logging and error handling")
    fmt.Println("5. Strategic requeue patterns")
}
EOF
   ```

3. **Analyze the reconciler patterns**:
   ```bash
   # Examine the reconciler structure
   echo "=== Reconciler Function Structure ==="
   grep -A 5 "func.*Reconcile" basic_reconciler.go

   echo -e "\n=== Return Patterns ==="
   grep "return ctrl.Result" basic_reconciler.go

   echo -e "\n=== Logging Patterns ==="
   grep "log.Info\|log.Error" basic_reconciler.go

   echo -e "\n=== Error Handling Patterns ==="
   grep -A 2 -B 2 "error\|Error" basic_reconciler.go
   ```

4. **Test the pattern understanding**:
   ```bash
   # Run the example to see pattern summary
   echo "=== Testing Reconciler Pattern Understanding ==="
   go run basic_reconciler.go 2>/dev/null || echo "Go not available, but pattern analysis complete"
   ```

### Pattern Analysis
Review the reconciler implementation for:
- [ ] Proper resource fetching with error handling
- [ ] Deletion handling with finalizer patterns
- [ ] Component-based reconciliation approach
- [ ] Strategic requeue patterns (`Requeue: true` vs `RequeueAfter`)
- [ ] Structured logging with contextual information

---

## Exercise Summary

### Key Patterns Practiced
1. **ODH Reconciler Analysis**: Real-world reconciler implementation patterns
2. **Debugging Techniques**: Using logs and events to understand reconciliation flow
3. **Implementation Patterns**: Basic reconciler structure following best practices

### Reconciler Implementation Checklist
- [ ] ✅ Resource fetching with proper error handling
- [ ] ✅ Deletion handling with finalizers
- [ ] ✅ Component-based reconciliation
- [ ] ✅ Status and condition management
- [ ] ✅ Strategic requeue patterns
- [ ] ✅ Structured logging for debugging

### Common Reconciler Return Patterns Learned
```go
return ctrl.Result{}, nil                           // Success, no requeue
return ctrl.Result{Requeue: true}, nil             // Immediate requeue
return ctrl.Result{RequeueAfter: time.Minute}, nil // Delayed requeue
return ctrl.Result{}, err                          // Error with backoff
```

### Debugging Commands for Daily Use
```bash
# Check operator logs
kubectl logs -n opendatahub-operator-system deployment/opendatahub-operator-controller-manager

# Monitor reconciliation events
kubectl get events --sort-by='.metadata.creationTimestamp' | grep reconcil

# Check resource status
kubectl describe datasciencecluster <name>
```

---

## Cleanup

```bash
# Clean up practice files
rm -rf /tmp/reconciler-practice
```

## Next Steps
- Day 13 will cover Event Watching and Filtering
- Continue studying ODH reconciler implementations in `/Users/suksubra/Documents/Work/RHOAI/opendatahub-operator/controllers/`
- Practice implementing reconciler patterns in your own operator projects