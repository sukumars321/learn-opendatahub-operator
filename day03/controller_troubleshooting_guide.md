# Day 3 Troubleshooting Guide: Controller Issues

## Common Controller Problems & Solutions

### Problem 1: Controller Not Responding to Changes

**Symptoms:**
- Made changes to DataScienceCluster spec
- No corresponding changes in deployed resources
- Controller logs show no reconciliation events

**Debugging Steps:**

```bash
# 1. Check if controller is running
oc get pods -n opendatahub-operator-system
oc logs deployment/opendatahub-operator-controller-manager -n opendatahub-operator-system --tail=50

# 2. Verify watch configuration
oc describe deployment opendatahub-operator-controller-manager -n opendatahub-operator-system

# 3. Check for webhook issues
oc get validatingwebhookconfiguration,mutatingwebhookconfiguration | grep opendatahub

# 4. Force reconciliation by adding annotation
oc annotate datasciencecluster default-dsc reconcile.opendatahub.io/timestamp="$(date +%s)"
```

**Common Causes:**
- Controller pod crashed or restarted
- RBAC permissions missing
- Webhook configuration blocking updates
- Watch cache not synced

**Solutions:**
- Restart controller deployment
- Check and fix RBAC permissions
- Verify webhook endpoints are accessible
- Wait for cache sync (usually automatic)

---

### Problem 2: Resources Created But Status Never Updates

**Symptoms:**
- Deployments and services are created successfully
- DataScienceCluster status remains in "NotReady" or shows old conditions
- Component conditions never transition to "Ready"

**Debugging Steps:**

```bash
# 1. Check resource readiness
oc get deployments -o wide | grep odh
oc describe deployment odh-dashboard

# 2. Examine status update logic in logs
oc logs deployment/opendatahub-operator-controller-manager -n opendatahub-operator-system | grep -i status

# 3. Check for status update conflicts
oc get datasciencecluster default-dsc -o yaml | grep resourceVersion

# 4. Manually trigger status check
oc patch datasciencecluster default-dsc --type='merge' -p='{"metadata":{"annotations":{"debug.opendatahub.io/force-status-update":"true"}}}'
```

**Common Causes:**
- Deployment ready but pods not ready
- Status update failing due to resource conflict
- Controller logic waiting for additional conditions
- Missing permissions to update status subresource

**Solutions:**
- Check pod readiness and fix underlying issues
- Implement retry logic for status updates
- Review component readiness conditions
- Verify status subresource permissions

---

### Problem 3: Infinite Reconciliation Loops

**Symptoms:**
- Controller logs show continuous reconciliation for same resource
- High CPU usage on controller
- Frequent "reconciling" messages in logs

**Debugging Steps:**

```bash
# 1. Monitor reconciliation frequency
oc logs deployment/opendatahub-operator-controller-manager -n opendatahub-operator-system | grep -E "Reconciling|reconcile" | tail -20

# 2. Check for resource modifications
oc get events --sort-by=.metadata.creationTimestamp | grep default-dsc | tail -10

# 3. Look for spec vs status conflicts
oc get datasciencecluster default-dsc -o yaml | yq '.spec' > /tmp/spec.yaml
oc get datasciencecluster default-dsc -o yaml | yq '.status' > /tmp/status.yaml
diff /tmp/spec.yaml /tmp/status.yaml

# 4. Monitor resource generation numbers
oc get datasciencecluster default-dsc -o jsonpath='{.metadata.generation} {.status.observedGeneration}'
```

**Common Causes:**
- Controller modifying spec during reconciliation
- Status update triggering unnecessary reconciliation
- External process modifying resources
- Admission webhooks modifying resources

**Solutions:**
- Only update status, never spec in reconciliation
- Use proper owner references to prevent external modifications
- Implement proper comparison logic to detect real changes
- Add generation tracking to avoid unnecessary reconciliation

---

### Problem 4: Resource Cleanup Not Working

**Symptoms:**
- Deleted DataScienceCluster but child resources remain
- Setting managementState to "Removed" doesn't clean up resources
- Resources stuck in "Terminating" state

**Debugging Steps:**

```bash
# 1. Check finalizers
oc get datasciencecluster -o yaml | grep finalizers -A 5

# 2. Check owner references on stuck resources
oc get deployment odh-dashboard -o yaml | grep ownerReferences -A 10

# 3. Look for terminating resources
oc get all | grep Terminating

# 4. Check for stuck finalizers on child resources
oc get pods,deployments,services -o yaml | grep finalizers -B 2 -A 2
```

**Common Causes:**
- Missing or incorrect owner references
- Finalizers not being removed properly
- External dependencies preventing deletion
- Admission webhooks blocking deletion

**Solutions:**
- Fix owner reference configuration
- Implement proper finalizer cleanup logic
- Remove external dependencies before deletion
- Check admission webhook logic for deletion events

---

### Problem 5: Controller Permission Errors

**Symptoms:**
- "Forbidden" errors in controller logs
- Controller cannot create/update/delete resources
- Operations work manually but not through controller

**Debugging Steps:**

```bash
# 1. Check controller service account
oc get sa -n opendatahub-operator-system
oc describe sa opendatahub-operator-controller-manager -n opendatahub-operator-system

# 2. Check role bindings
oc get clusterrolebinding | grep opendatahub
oc describe clusterrolebinding opendatahub-operator-manager-rolebinding

# 3. Test permissions manually
oc auth can-i create deployments --as=system:serviceaccount:opendatahub-operator-system:opendatahub-operator-controller-manager
oc auth can-i update datascienceclusters/status --as=system:serviceaccount:opendatahub-operator-system:opendatahub-operator-controller-manager

# 4. Check for missing permissions in logs
oc logs deployment/opendatahub-operator-controller-manager -n opendatahub-operator-system | grep -i forbidden
```

**Common Causes:**
- Missing RBAC rules for new resource types
- Incorrect service account configuration
- Cluster role not bound to service account
- Missing permissions for status subresource

**Solutions:**
- Add missing RBAC rules to ClusterRole
- Verify service account is correctly referenced
- Check ClusterRoleBinding configuration
- Add status subresource permissions

---

## Diagnostic Commands Reference

### Controller Health Check
```bash
#!/bin/bash
echo "=== ODH Controller Health Check ==="

# Controller pod status
echo "📊 Controller Pod Status:"
oc get pods -n opendatahub-operator-system -o wide

# Recent logs
echo "📝 Recent Controller Logs:"
oc logs deployment/opendatahub-operator-controller-manager -n opendatahub-operator-system --tail=10

# Resource status
echo "🎯 Resource Status:"
oc get datasciencecluster,deployments,services | grep -E "(NAME|default-dsc|odh-)"

# Reconciliation metrics
echo "📈 Controller Metrics:"
oc port-forward svc/opendatahub-operator-controller-manager-metrics-service 8080 -n opendatahub-operator-system &
sleep 2
curl -s http://localhost:8080/metrics | grep controller_runtime_reconcile_total | head -5
pkill -f "port-forward.*8080"
```

### Reconciliation Trace
```bash
#!/bin/bash
echo "=== Reconciliation Trace ==="

# Watch reconciliation in real-time
echo "Starting reconciliation trace (Ctrl+C to stop):"
oc logs -f deployment/opendatahub-operator-controller-manager -n opendatahub-operator-system | \
grep -E "(Reconciling|Error|reconcile)" | \
while read line; do
    echo "$(date '+%H:%M:%S') | $line"
done
```

### Status Deep Dive
```bash
#!/bin/bash
echo "=== Status Deep Dive ==="

dsc_name=${1:-default-dsc}

echo "📊 Overall Status:"
oc get datasciencecluster $dsc_name -o jsonpath='{.status.phase}'
echo

echo "🎯 Detailed Conditions:"
oc get datasciencecluster $dsc_name -o json | jq -r '
.status.conditions[]? |
"Type: \(.type)
Status: \(.status)
Reason: \(.reason)
Message: \(.message)
Last Transition: \(.lastTransitionTime)
---"'

echo "🔍 Component Status Summary:"
oc get datasciencecluster $dsc_name -o json | jq -r '
.status.conditions[]? |
select(.type | endswith("Ready")) |
"Component: \(.type | sub("Ready$"; "")) | Status: \(.status)"'
```

### Owner Reference Audit
```bash
#!/bin/bash
echo "=== Owner Reference Audit ==="

dsc_name=${1:-default-dsc}
dsc_uid=$(oc get datasciencecluster $dsc_name -o jsonpath='{.metadata.uid}')

echo "🔗 Resources owned by DataScienceCluster $dsc_name:"
echo "DataScienceCluster UID: $dsc_uid"
echo

for resource_type in deployment service configmap secret role rolebinding route; do
    echo "📦 $resource_type resources:"
    oc get $resource_type -A -o json | jq -r --arg uid "$dsc_uid" '
    .items[] |
    select(.metadata.ownerReferences[]?.uid == $uid) |
    "  - \(.metadata.namespace // "cluster-scope")/\(.metadata.name)"'
done

echo "⚠️  Resources without proper owner references:"
oc get deployment,service -o json | jq -r '
.items[] |
select(.metadata.labels["app.opendatahub.io"] != null) |
select(.metadata.ownerReferences == null or (.metadata.ownerReferences | length) == 0) |
"  - \(.kind)/\(.metadata.name) in \(.metadata.namespace)"'
```

## Prevention Best Practices

### 1. Proper Error Handling
```go
// Always handle errors gracefully
if err := r.createDeployment(ctx, deployment); err != nil {
    log.Error(err, "Failed to create deployment", "deployment", deployment.Name)
    // Update status to reflect error
    r.updateCondition(ctx, resource, "ComponentReady", metav1.ConditionFalse, "DeploymentFailed", err.Error())
    // Return error to trigger requeue with backoff
    return ctrl.Result{}, fmt.Errorf("failed to create deployment %s: %w", deployment.Name, err)
}
```

### 2. Idempotent Operations
```go
// Always check if resource exists before creating
existing := &appsv1.Deployment{}
err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
if err != nil && errors.IsNotFound(err) {
    // Resource doesn't exist, create it
    return r.Create(ctx, desired)
} else if err != nil {
    // Other error
    return fmt.Errorf("failed to get deployment: %w", err)
}

// Resource exists, check if update needed
if !reflect.DeepEqual(existing.Spec, desired.Spec) {
    existing.Spec = desired.Spec
    return r.Update(ctx, existing)
}

// No changes needed
return nil
```

### 3. Proper Status Management
```go
// Always separate spec and status updates
func (r *Reconciler) updateStatus(ctx context.Context, resource *MyResource, condition metav1.Condition) error {
    // Get fresh copy for status update
    fresh := &MyResource{}
    if err := r.Get(ctx, client.ObjectKeyFromObject(resource), fresh); err != nil {
        return err
    }

    // Update conditions
    meta.SetStatusCondition(&fresh.Status.Conditions, condition)

    // Calculate overall phase
    fresh.Status.Phase = r.calculatePhase(fresh.Status.Conditions)

    // Update only status subresource
    return r.Status().Update(ctx, fresh)
}
```

## When to Escalate

Contact the ODH team or file an issue if you encounter:

1. **Controller crashes repeatedly** with stack traces
2. **Memory or CPU usage continuously growing** (memory leak)
3. **Reconciliation takes extremely long** (> 5 minutes for simple operations)
4. **Data corruption or resource conflicts** that prevent recovery
5. **Security-related errors** or permission escalation issues

For community support:
- **GitHub Issues**: https://github.com/opendatahub-io/opendatahub-operator/issues
- **Slack**: #opendatahub-operator channel
- **Documentation**: https://opendatahub.io/docs/