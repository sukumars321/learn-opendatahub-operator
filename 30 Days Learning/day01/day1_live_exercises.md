# Day 1 Live Cluster Exercises: Kubernetes API Fundamentals

Perfect! You have access to a live OpenShift cluster with ODH installed. This makes your learning much more concrete and valuable.

## 🔍 What You Just Discovered

From your `kubectl api-resources` and `kubectl api-versions` commands, you can see the ODH operator is actually running! Look at these ODH-specific API groups:

### ODH API Groups Found:
```
components.platform.opendatahub.io/v1alpha1
datasciencecluster.opendatahub.io/v1
datasciencepipelinesapplications.opendatahub.io/v1
dscinitialization.opendatahub.io/v1
dashboard.opendatahub.io/v1
infrastructure.opendatahub.io/v1
features.opendatahub.io/v1
services.platform.opendatahub.io/v1alpha1
```

### Key ODH Resources Available:
```
datascienceclusters                   dsc                  datasciencecluster.opendatahub.io/v1
dscinitializations                    dsci                 dscinitialization.opendatahub.io/v1
codeflares                                                 components.platform.opendatahub.io/v1alpha1
dashboards                                                 components.platform.opendatahub.io/v1alpha1
datasciencepipelines                                       components.platform.opendatahub.io/v1alpha1
workbenches                                                components.platform.opendatahub.io/v1alpha1
modelmeshservings                                          components.platform.opendatahub.io/v1alpha1
```

---

## 🚀 Live Exercises (15 minutes)

### Exercise 1: Explore Live ODH CRDs

```bash
# See all ODH-related CRDs
oc get crd | grep opendatahub

# Look at the main DataScienceCluster CRD
oc get crd datascienceclusters.datasciencecluster.opendatahub.io -o yaml

# See the component CRDs
oc get crd | grep components.platform.opendatahub.io
```

**Run these now and observe:**
- How many ODH CRDs are installed?
- What's the structure of the DataScienceCluster CRD?
- What component types are available?

### Exercise 2: Check Live ODH Resources

```bash
# See if there are any DataScienceClusters running
oc get datascienceclusters
oc get dsc

# Check DSC initialization
oc get dscinitializations
oc get dsci

# Look at component resources
oc get dashboards
oc get workbenches
oc get datasciencepipelines
```

**Questions to answer:**
- Are there any DataScienceClusters already deployed?
- What's the status of ODH initialization?
- Which components are currently managed?

### Exercise 3: Examine Real ODH Resource Structure

If you find any resources, examine them:

```bash
# Get detailed info about a DataScienceCluster (if one exists)
oc get dsc -o yaml

# Look at the status and spec sections
oc describe dsc

# Check component status
oc get dashboard -o yaml
oc get workbenches -o yaml
```

**Learning Points:**
- See the real `spec` and `status` sections
- Understand how conditions are reported
- Notice owner references and relationships

### Exercise 4: Explore API Endpoints

```bash
# Use kubectl to explore the API paths directly
oc api-resources | grep datasciencecluster
oc api-resources | grep components.platform

# Explain the structure of these resources
oc explain datasciencecluster
oc explain datasciencecluster.spec
oc explain datasciencecluster.status

# Look at component resources
oc explain dashboard
oc explain workbenches
```

**Key Observations:**
- How do the API paths match what you learned in theory?
- What fields are available in spec vs status?
- How are the component relationships structured?

### Exercise 5: Understanding API Groups in Practice

```bash
# Compare core Kubernetes resources
oc api-resources | grep "^pods"
oc api-resources | grep "^deployments"

# Compare with ODH custom resources
oc api-resources | grep datasciencecluster
oc api-resources | grep components.platform

# Look at versioning
oc api-versions | grep opendatahub
oc api-versions | grep components.platform
```

**Analysis Questions:**
1. How do ODH APIs follow Kubernetes conventions?
2. Why are some resources `v1` and others `v1alpha1`?
3. What's the difference between cluster-scoped and namespaced resources?

---

## 🧠 Real-World Insights

### What You're Seeing:
1. **Multiple API Groups**: ODH organizes functionality into logical groups
2. **Version Progression**: Some APIs are stable (v1), others experimental (v1alpha1)
3. **Resource Hierarchy**: DataScienceCluster manages component resources
4. **Status Reporting**: Real controllers updating status conditions

### How This Connects to Day 1 Learning:
- **API Server**: Processing your `oc get` requests
- **etcd**: Storing the resource state you're viewing
- **Controllers**: ODH controllers managing these resources
- **REST Patterns**: Every `oc` command maps to HTTP requests

---

## 🔍 Bonus Discovery Exercise

If time permits, explore the controller that makes it all work:

```bash
# Find the ODH operator pod
oc get pods -n opendatahub-operator-system

# Look at the operator logs (if accessible)
oc logs -n opendatahub-operator-system deployment/opendatahub-operator-controller-manager

# Check the operator's RBAC
oc get clusterrole | grep opendatahub
oc describe clusterrole opendatahub-operator-manager-role
```

---

## 📝 Key Takeaways from Live Environment

### What Makes This Real:
1. **Active Controllers**: The ODH operator is running and managing resources
2. **Live State**: You can see actual `spec` and `status` sections
3. **API Extensions**: ODH successfully extends Kubernetes with custom resources
4. **Production Patterns**: Real RBAC, real resource management, real status reporting

### Connection to Operator Development:
- Every CRD you see was generated by Kubebuilder (Day 8-14 topics)
- The controllers watching these resources use controller-runtime (Day 6 topic)
- The status conditions follow Kubernetes conventions
- The resource relationships show owner-reference patterns

---

## ⏰ Time Check and Next Steps

**Time for exercises:** _____ minutes (target: 15)

### Ready for the Study Phase?
Now that you've seen the real ODH APIs in action, the theoretical concepts in your study guide will be much more concrete. The live cluster gives you:
- Real examples of every concept
- Actual resource relationships
- Live status reporting
- Production API patterns

### Continue Your Day 1 Journey:
1. **Complete the study guide** (`day1_complete_guide.md`) - concepts will now make perfect sense
2. **Reference these live examples** as you read about API fundamentals
3. **Take notes** about what you discovered in the cluster
4. **Reflect** on what you learned from both study and hands-on work

Having a live cluster makes your learning journey incredibly valuable - you're seeing production operator patterns in action! 🚀