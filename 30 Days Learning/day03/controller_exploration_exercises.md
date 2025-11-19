# Day 3 Practical Exercises: Controller Deep Dive

## Exercise Set 1: Controller Observation Lab

### Exercise 1.1: Real-time Controller Monitoring

Set up multiple terminals to observe controller behavior in real-time:

**Terminal 1: Controller Logs**
```bash
# Watch ODH controller logs with filtering
oc logs -f deployment/opendatahub-operator-controller-manager \
  -n opendatahub-operator-system | \
  grep -E "(Reconciling|Error|Created|Updated|Deleted)"
```

**Terminal 2: Resource Watching**
```bash
# Watch all ODH-related resources
watch -n 2 'oc get datasciencecluster,deployments,services,configmaps -A | grep -E "(odh|dashboard|workbench)"'
```

**Terminal 3: Status Monitoring**
```bash
# Monitor status conditions
watch -n 5 'oc get datasciencecluster default-dsc -o jsonpath="{.status}" | jq'
```

**Terminal 4: Event Stream**
```bash
# Watch Kubernetes events related to ODH
oc get events --watch --field-selector involvedObject.apiVersion=datasciencecluster.opendatahub.io/v1
```

### Exercise 1.2: Trigger Controller Actions

Now trigger various events and observe the cascade:

**Test 1: Component State Change**
```bash
# Change dashboard from Managed to Removed
oc patch datasciencecluster default-dsc --type='merge' \
  -p='{"spec":{"components":{"dashboard":{"managementState":"Removed"}}}}'

# Wait 30 seconds, then change back
sleep 30
oc patch datasciencecluster default-dsc --type='merge' \
  -p='{"spec":{"components":{"dashboard":{"managementState":"Managed"}}}}'
```

**Test 2: Simulate Resource Drift**
```bash
# Delete a deployment and watch controller recreate it
oc delete deployment odh-dashboard

# Scale a deployment and watch controller correct it
oc scale deployment odh-dashboard --replicas=3
```

**Test 3: ConfigMap Changes**
```bash
# Create a ConfigMap that might trigger reconciliation
oc create configmap test-config --from-literal=test=value
oc label configmap test-config app.opendatahub.io/datasciencecluster=default-dsc
```

**Learning Goals:**
- Observe reconciliation timing and patterns
- See how controllers respond to different event types
- Understand the relationship between spec changes and status updates

---

## Exercise Set 2: Owner Reference Investigation

### Exercise 2.1: Mapping the Ownership Hierarchy

```bash
# Create a script to visualize ownership relationships
cat << 'EOF' > ownership_mapper.sh
#!/bin/bash

echo "=== ODH Resource Ownership Hierarchy ==="
echo

# Get the main DataScienceCluster
echo "📊 DataScienceCluster:"
oc get datasciencecluster -o name

echo
echo "📦 Resources owned by DataScienceCluster:"

# Find all resources with owner references to DSC
for resource_type in deployment service configmap secret role rolebinding route; do
    echo "  $resource_type:"
    oc get $resource_type -A -o json | \
    jq -r '.items[] | select(.metadata.ownerReferences[]?.kind == "DataScienceCluster") | "    - \(.metadata.namespace)/\(.metadata.name)"' 2>/dev/null || true
done

echo
echo "🔗 Owner Reference Details:"
oc get deployment odh-dashboard -o json | \
jq '.metadata.ownerReferences[]' 2>/dev/null || echo "No dashboard deployment found"
EOF

chmod +x ownership_mapper.sh
./ownership_mapper.sh
```

### Exercise 2.2: Test Garbage Collection

**⚠️ CAUTION: Only do this in a test environment!**

```bash
# 1. Create a test DataScienceCluster
cat << 'EOF' | oc apply -f -
apiVersion: datasciencecluster.opendatahub.io/v1
kind: DataScienceCluster
metadata:
  name: test-dsc
spec:
  components:
    dashboard:
      managementState: Managed
EOF

# 2. Wait for resources to be created
sleep 60

# 3. List resources owned by test-dsc
echo "Resources before deletion:"
oc get all -A --show-labels | grep test-dsc

# 4. Delete the DataScienceCluster
oc delete datasciencecluster test-dsc

# 5. Watch resources disappear
echo "Watching for resource cleanup..."
for i in {1..30}; do
    count=$(oc get all -A --show-labels 2>/dev/null | grep -c test-dsc || echo "0")
    echo "Remaining resources: $count"
    if [ "$count" -eq "0" ]; then
        echo "All resources cleaned up!"
        break
    fi
    sleep 2
done
```

**Learning Goals:**
- Understand how owner references enable automatic cleanup
- See Kubernetes garbage collection in action
- Appreciate the importance of proper resource ownership

---

## Exercise Set 3: Controller Code Analysis

### Exercise 3.1: Trace a Reconciliation Path

Create a code reading guide:

```bash
# Create a study guide for following reconciliation flow
cat << 'EOF' > reconciliation_trace.md
# Reconciliation Flow Trace

## Starting Point: User Action
```yaml
# User runs: oc patch datasciencecluster default-dsc --type='merge' -p='{"spec":{"components":{"dashboard":{"managementState":"Removed"}}}}'
```

## Code Path to Follow:

### 1. Controller Registration
**File**: `controllers/datasciencecluster/datasciencecluster_controller.go`
**Function**: `SetupWithManager()`
- Look for: Watch configuration for DataScienceCluster resources
- Find: How controller registers to receive events

### 2. Event Reception
**File**: `controllers/datasciencecluster/datasciencecluster_controller.go`
**Function**: `Reconcile()`
- Look for: How the controller receives the reconcile request
- Find: Request object containing namespace/name

### 3. Resource Fetching
**Function**: `Reconcile()` continued
- Look for: `r.Get()` call to fetch the DataScienceCluster
- Find: Error handling for not found resources

### 4. Component Processing
**File**: `controllers/components/dashboard.go`
**Function**: `ReconcileComponent()`
- Look for: ManagementState decision logic
- Find: What happens when state is "Removed"

### 5. Resource Cleanup
**Function**: `Cleanup()` or similar
- Look for: How dashboard resources are identified and deleted
- Find: Owner reference usage

### 6. Status Update
**Function**: Status update logic
- Look for: How status.conditions are updated
- Find: Status phase calculation

## Questions to Answer While Reading:
1. How does the controller know which resources to delete?
2. What error handling exists if deletion fails?
3. How is the status updated to reflect the change?
4. What triggers a requeue if something goes wrong?
EOF
```

### Exercise 3.2: Controller Metrics Analysis

```bash
# Set up port forwarding to controller metrics
oc port-forward svc/opendatahub-operator-controller-manager-metrics-service 8080 -n opendatahub-operator-system &

# Create metrics analysis script
cat << 'EOF' > analyze_controller_metrics.sh
#!/bin/bash

echo "=== Controller Runtime Metrics ==="
echo

echo "📈 Reconciliation Metrics:"
curl -s http://localhost:8080/metrics | grep controller_runtime_reconcile | head -10

echo
echo "⏱️  Reconciliation Duration:"
curl -s http://localhost:8080/metrics | grep reconcile_time | head -5

echo
echo "🔄 Work Queue Metrics:"
curl -s http://localhost:8080/metrics | grep workqueue | head -10

echo
echo "⚠️  Error Metrics:"
curl -s http://localhost:8080/metrics | grep error | head -5

echo
echo "💾 Resource Metrics:"
curl -s http://localhost:8080/metrics | grep -E "(go_memstats|process_)" | head -5
EOF

chmod +x analyze_controller_metrics.sh
./analyze_controller_metrics.sh

# Kill the port-forward when done
pkill -f "port-forward.*8080"
```

**Learning Goals:**
- Understand how to monitor controller performance
- Identify bottlenecks and error patterns
- Connect metrics to controller behavior

---

## Exercise Set 4: Status and Conditions Deep Dive

### Exercise 4.1: Status Condition Analysis

```bash
# Create a status monitoring script
cat << 'EOF' > status_monitor.sh
#!/bin/bash

echo "=== DataScienceCluster Status Analysis ==="
echo

# Get current status
dsc_status=$(oc get datasciencecluster default-dsc -o json | jq '.status')

echo "📊 Overall Status:"
echo "$dsc_status" | jq -r '.phase // "Unknown"'

echo
echo "🎯 Status Conditions:"
echo "$dsc_status" | jq -r '.conditions[]? | "Type: \(.type) | Status: \(.status) | Reason: \(.reason) | Message: \(.message)"'

echo
echo "⏰ Last Transition Times:"
echo "$dsc_status" | jq -r '.conditions[]? | "Type: \(.type) | Last Change: \(.lastTransitionTime)"'

echo
echo "🔍 Condition Summary:"
ready_count=$(echo "$dsc_status" | jq -r '.conditions[]? | select(.status == "True") | .type' | wc -l)
total_count=$(echo "$dsc_status" | jq -r '.conditions[]? | .type' | wc -l)
echo "Ready Conditions: $ready_count / $total_count"

if [ "$ready_count" -eq "$total_count" ] && [ "$total_count" -gt "0" ]; then
    echo "✅ All components are ready"
else
    echo "⚠️  Some components are not ready"
    echo "$dsc_status" | jq -r '.conditions[]? | select(.status != "True") | "❌ \(.type): \(.message)"'
fi
EOF

chmod +x status_monitor.sh
./status_monitor.sh
```

### Exercise 4.2: Simulate Status Changes

```bash
# Create status change simulation
cat << 'EOF' > simulate_status_changes.sh
#!/bin/bash

echo "=== Simulating Status Changes ==="
echo

# Function to show current status
show_status() {
    echo "📊 Current Status ($(date)):"
    oc get datasciencecluster default-dsc -o json | \
    jq -r '.status.conditions[]? | "  \(.type): \(.status) (\(.reason))"'
    echo
}

# Initial status
show_status

# 1. Remove a component
echo "🔄 Step 1: Removing dashboard component"
oc patch datasciencecluster default-dsc --type='merge' \
  -p='{"spec":{"components":{"dashboard":{"managementState":"Removed"}}}}'

echo "Waiting for reconciliation..."
sleep 30
show_status

# 2. Add it back
echo "🔄 Step 2: Re-enabling dashboard component"
oc patch datasciencecluster default-dsc --type='merge' \
  -p='{"spec":{"components":{"dashboard":{"managementState":"Managed"}}}}'

echo "Waiting for reconciliation..."
sleep 60
show_status

echo "✅ Status change simulation complete"
EOF

chmod +x simulate_status_changes.sh
./simulate_status_changes.sh
```

**Learning Goals:**
- Understand how status conditions reflect system state
- See how conditions change over time
- Learn to interpret condition messages and reasons

---

## Exercise Set 5: Advanced Controller Patterns

### Exercise 5.1: Finalizer Investigation

```bash
# Examine finalizer usage in ODH
echo "=== Finalizer Analysis ==="

echo "📋 DataScienceCluster Finalizers:"
oc get datasciencecluster default-dsc -o json | jq '.metadata.finalizers'

echo
echo "🔍 Resources with Finalizers:"
oc get all -A -o json | \
jq -r '.items[] | select(.metadata.finalizers != null) | "\(.kind)/\(.metadata.name) in \(.metadata.namespace // "cluster-scope"): \(.metadata.finalizers | join(", "))"'

echo
echo "⚙️  Understanding Finalizer Flow:"
cat << 'EOF'
Finalizer Flow:
1. User deletes resource (kubectl delete)
2. Kubernetes sets deletionTimestamp but keeps resource
3. Controller sees deletionTimestamp and runs cleanup
4. Controller removes its finalizer when cleanup complete
5. Kubernetes deletes resource when all finalizers removed

ODH Finalizer Pattern:
- Ensures orderly cleanup of complex resource hierarchies
- Prevents orphaned resources in child clusters
- Allows graceful shutdown of running workloads
EOF
```

### Exercise 5.2: Controller Error Simulation

**⚠️ Do this carefully in test environment only!**

```bash
# Simulate various error conditions
echo "=== Error Condition Simulation ==="

# 1. Create resource with invalid spec (if possible)
echo "🚫 Test 1: Invalid ManagementState"
cat << 'EOF' | oc apply -f - --dry-run=server
apiVersion: datasciencecluster.opendatahub.io/v1
kind: DataScienceCluster
metadata:
  name: invalid-dsc
spec:
  components:
    dashboard:
      managementState: "InvalidState"  # This should be rejected by validation
EOF

# 2. Resource quota exhaustion simulation
echo "🚫 Test 2: Simulating resource pressure"
echo "Creating ResourceQuota to limit resources..."
cat << 'EOF' | oc apply -f -
apiVersion: v1
kind: ResourceQuota
metadata:
  name: test-quota
  namespace: opendatahub
spec:
  hard:
    requests.cpu: "1m"
    requests.memory: "1Mi"
EOF

echo "This should cause deployment failures..."
oc describe resourcequota test-quota -n opendatahub

# Clean up
oc delete resourcequota test-quota -n opendatahub
```

**Learning Goals:**
- Understand how controllers handle error conditions
- See validation in action
- Learn about resource constraints and their effects

---

## Synthesis Exercise: Controller Behavior Model

Create a comprehensive mental model:

```bash
cat << 'EOF' > controller_behavior_model.md
# OpenDataHub Controller Behavior Model

## Architecture Layers

```
User Intent (kubectl apply)
         ↓
DataScienceCluster Resource (stored in etcd)
         ↓
DataScienceCluster Controller (watches and reconciles)
         ↓
Component Controllers (dashboard, workbenches, etc.)
         ↓
Kubernetes Resources (deployments, services, etc.)
         ↓
Container Runtime (actual running pods)
```

## Event Flow Patterns

### Happy Path:
1. **User Action** → Spec change
2. **Event Generation** → Watch triggers
3. **Reconciliation** → Controller compares desired vs actual
4. **Action Execution** → Create/update/delete resources
5. **Status Update** → Report back to user
6. **Stability** → No more changes needed

### Error Path:
1. **User Action** → Spec change
2. **Event Generation** → Watch triggers
3. **Reconciliation** → Controller compares desired vs actual
4. **Action Failure** → Resource creation fails
5. **Error Reporting** → Status shows error condition
6. **Retry Logic** → Requeue with backoff
7. **Recovery** → Eventually succeeds or user intervenes

## Key Insights

### What Makes ODH Controllers Sophisticated:
- **Multi-level hierarchy** reduces complexity
- **Component abstraction** enables modularity
- **Platform awareness** handles OpenShift vs Kubernetes
- **Rich status reporting** provides detailed feedback
- **Owner references** enable automatic cleanup
- **Finalizers** ensure orderly shutdown

### Controller Design Principles:
- **Idempotent operations** - safe to run multiple times
- **Level-triggered** - react to current state, not edges
- **Error recovery** - always try to reach desired state
- **Graceful degradation** - partial failures don't stop everything
- **Observability** - rich logging and metrics
EOF
```

## Next Steps for Day 4

After completing these exercises, you should be ready for Day 4 topics:

1. **Go Programming Patterns** - How to structure controller code
2. **Client-go Library** - Kubernetes API interaction patterns
3. **Controller-runtime** - High-level controller building blocks
4. **Testing Patterns** - How to test controller logic
5. **Error Handling** - Robust error management strategies

The exercises above give you hands-on experience with the concepts you'll be implementing in code on Day 4!