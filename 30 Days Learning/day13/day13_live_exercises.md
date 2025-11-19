# Day 13: Watching and Event Filtering - Live Exercises

## Exercise Time: 20 minutes

These hands-on exercises will help you explore real watch configurations and event filtering mechanisms in the ODH operator codebase, giving you practical experience with the concepts covered in the study guide.

---

## Exercise 1: ODH Watch Configuration Analysis (8 minutes)

### Goal
Analyze the watch setup in ODH controllers to understand real-world watch patterns and configurations.

### Steps

1. **Navigate to the ODH operator codebase**:
   ```bash
   cd /Users/suksubra/Documents/Work/RHOAI/opendatahub-operator
   ```

2. **Examine DataScienceCluster controller watch setup**:
   ```bash
   # Find the SetupWithManager method
   grep -A 30 "func.*SetupWithManager" internal/controller/datasciencecluster/datasciencecluster_controller.go

   # Look for watch configurations
   grep -A 5 -B 2 "\.Owns\|\.Watches\|\.For" internal/controller/datasciencecluster/datasciencecluster_controller.go
   ```

3. **Analyze component controller watch patterns**:
   ```bash
   # Check workbenches controller
   grep -A 20 "SetupWithManager" internal/controller/components/workbenches/workbenches_controller.go

   # Find predicate usage
   grep -A 3 -B 3 "Predicate\|predicate" internal/controller/components/workbenches/workbenches_controller.go
   ```

4. **Explore cross-resource watching**:
   ```bash
   # Look for event mappers
   grep -A 10 -B 5 "EventMapper\|WithEventMapper" internal/controller/datasciencecluster/datasciencecluster_controller.go

   # Find watch functions
   grep -n "watchDataScienceClusters\|watch.*Clusters" internal/controller/datasciencecluster/datasciencecluster_controller.go
   ```

### Analysis Questions
- How many different resource types does the DataScienceCluster controller watch?
- What predicates are used and why?
- How does the event mapper work for DSCInitialization watching?

---

## Exercise 2: Predicate Implementation Deep Dive (7 minutes)

### Goal
Understand how ODH implements custom predicates for efficient event filtering.

### Steps

1. **Locate predicate implementations**:
   ```bash
   # Find predicate directories
   find . -path "*/predicates/*" -name "*.go" | head -10

   # Look for dependent predicate
   ls -la pkg/controller/predicates/dependent/
   ```

2. **Analyze the dependent predicate**:
   ```bash
   # View the dependent predicate implementation
   cat pkg/controller/predicates/dependent/predicate.go

   # Find how it's used
   grep -r "dependent\.New\|dependent\.WithWatchStatus" internal/controller/
   ```

3. **Examine resource-specific predicates**:
   ```bash
   # Check deployment predicate
   cat pkg/controller/predicates/resources/deployment_predicate.go

   # Find component predicates
   find . -name "*predicate*.go" -exec grep -l "component\|Component" {} \;
   ```

4. **Understand predicate usage patterns**:
   ```bash
   # Find all predicate usages
   grep -r "WithPredicates\|Predicate" internal/controller/ | grep -v ".git" | head -10

   # Look for label-based predicates
   grep -A 5 -B 5 "ForLabel\|labels\." internal/controller/components/workbenches/workbenches_controller.go
   ```

### Analysis Questions
- What conditions trigger the dependent predicate to return true?
- How does the deployment predicate optimize reconciliation?
- What label-based filtering is used in component controllers?

---

## Exercise 3: Event Flow Tracing (5 minutes)

### Goal
Trace how events flow from resource changes to reconciler execution in ODH.

### Steps

1. **Set up a local ODH environment** (if available):
   ```bash
   # Check if ODH is running locally
   kubectl get datasciencecluster -o yaml 2>/dev/null || echo "No DSC found"

   # Check for ODH controllers
   kubectl get pods -n opendatahub-operator-system 2>/dev/null || echo "ODH not running"
   ```

2. **Examine controller logs for watch events** (if running):
   ```bash
   # Get controller logs to see watch activity
   kubectl logs -n opendatahub-operator-system -l control-plane=controller-manager --tail=50

   # Look for reconciliation triggers
   kubectl logs -n opendatahub-operator-system -l control-plane=controller-manager | grep -i "reconcil\|watch\|event"
   ```

3. **Analyze event generation patterns**:
   ```bash
   # Find event recording in ODH code
   grep -r "Event\|recorder\." internal/controller/ | grep -v ".git" | head -5

   # Look for status update patterns that trigger events
   grep -A 5 -B 5 "Status.*Update\|updateStatus" internal/controller/datasciencecluster/datasciencecluster_controller.go
   ```

4. **Test watch behavior** (if environment available):
   ```bash
   # Create a simple resource change and watch for events
   kubectl get events --sort-by='.metadata.creationTimestamp' | tail -10

   # Monitor for new events
   kubectl get events --watch --field-selector reason=ReconciliationStarted &
   sleep 5
   kill %1 2>/dev/null
   ```

### Analysis Questions
- What types of events trigger DataScienceCluster reconciliation?
- How quickly do status changes propagate through the watch system?
- What events are generated during reconciliation?

---

## Bonus Challenge: Design Your Own Watch Configuration (Optional)

If you complete the exercises early, try this design challenge:

### Scenario
Design a watch configuration for a hypothetical "MLWorkflow" controller that needs to:
- Watch MLWorkflow custom resources (primary)
- React to changes in Jobs and Pods it creates
- Monitor ConfigMaps that contain workflow templates
- Trigger reconciliation when ServiceAccounts change permissions

### Your Task
Write pseudo-code for the `SetupWithManager` method with appropriate:
- Watch configurations (`.For()`, `.Owns()`, `.Watches()`)
- Predicates for efficient filtering
- Event mappers for cross-resource relationships

### Example Structure
```go
func (r *MLWorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&mlv1.MLWorkflow{}).
        // Add your watch configurations here
        Complete(r)
}
```

---

## Key Observations to Document

As you work through these exercises, note:

1. **Watch Patterns**: Which resources are watched and why
2. **Predicate Logic**: How predicates filter events efficiently
3. **Cross-Resource Relationships**: How changes in one resource trigger reconciliation of another
4. **Performance Considerations**: How ODH optimizes watch performance
5. **Debugging Insights**: What information is available for troubleshooting watch issues

These real-world patterns from ODH will help you design effective watch configurations in your own operators.

---

## Troubleshooting Tips

**If commands don't find expected files**:
- Check you're in the right directory: `/Users/suksubra/Documents/Work/RHOAI/opendatahub-operator`
- ODH code structure may have changed - adapt paths as needed

**If no ODH instance is running**:
- Focus on static code analysis
- Use `grep` and `find` commands to explore patterns
- The code analysis alone provides valuable insights

**If you find different code patterns**:
- ODH evolves rapidly - document what you find
- Compare patterns across different controllers
- Note any new predicate or watch patterns you discover