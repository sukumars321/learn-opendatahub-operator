# Day 2 Live Cluster Exercises: Custom Resource Definitions Deep Dive

Building on your Day 1 success! You've seen the ODH APIs in action - now let's understand how CRDs make it all possible.

## 🎯 Quick Warm-up: What You Already Know

From Day 1, you discovered ODH has these API groups:
- `datasciencecluster.opendatahub.io/v1`
- `components.platform.opendatahub.io/v1alpha1`
- `dscinitialization.opendatahub.io/v1`

**Today's Focus**: Understand how these APIs were created and how they work.

---

## 🔬 Live Exercises (20 minutes)

### Exercise 1: CRD Deep Dive (5 minutes)

```bash
# Get the complete list of ODH CRDs
oc get crd | grep opendatahub | wc -l
echo "ODH defines this many custom resource types!"

# Examine the main DataScienceCluster CRD structure
oc get crd datascienceclusters.datasciencecluster.opendatahub.io -o yaml | head -50

# Look at the schema section specifically
oc get crd datascienceclusters.datasciencecluster.opendatahub.io -o jsonpath='{.spec.versions[0].schema.openAPIV3Schema}' | jq .

# Compare with a component CRD
oc get crd dashboards.components.platform.opendatahub.io -o yaml | grep -A 20 "openAPIV3Schema:"
```

**Analysis Questions:**
1. How many CRDs does ODH actually define?
2. What's in the `openAPIV3Schema` section?
3. How does the schema define validation rules?

### Exercise 2: Live Resource Inspection (5 minutes)

```bash
# Check what DataScienceClusters exist
oc get datascienceclusters
oc get dsc

# If one exists, examine its structure
oc get dsc -o yaml | head -100

# Look specifically at spec vs status
oc get dsc -o jsonpath='{.items[0].spec}' | jq .
oc get dsc -o jsonpath='{.items[0].status}' | jq .

# Check component resources
oc get dashboards -o yaml
oc get workbenches -o yaml
```

**Key Observations:**
- How is `spec` different from `status`?
- What conditions are reported in status?
- How do owner references connect resources?

### Exercise 3: Schema Validation in Action (5 minutes)

```bash
# Use kubectl explain to see the live schema
oc explain datasciencecluster
oc explain datasciencecluster.spec
oc explain datasciencecluster.spec.components
oc explain datasciencecluster.spec.components.dashboard

# Look at validation details
oc explain datasciencecluster.spec.components.dashboard.managementState

# Compare component schemas
oc explain dashboard.spec
oc explain workbenches.spec
oc explain modelmeshserving.spec
```

**Understanding Questions:**
- What validation rules can you see?
- Which fields are required vs optional?
- How detailed is the auto-generated documentation?

### Exercise 4: Test CRD Validation (5 minutes)

Create a test file to understand validation:

```bash
# Create a valid DataScienceCluster spec
cat << 'EOF' > test-valid-dsc.yaml
apiVersion: datasciencecluster.opendatahub.io/v1
kind: DataScienceCluster
metadata:
  name: test-cluster
spec:
  components:
    dashboard:
      managementState: Managed
    workbenches:
      managementState: Removed
EOF

# Test client-side validation
oc apply --dry-run=client -f test-valid-dsc.yaml

# Test server-side validation
oc apply --dry-run=server -f test-valid-dsc.yaml

# Now test with invalid data
cat << 'EOF' > test-invalid-dsc.yaml
apiVersion: datasciencecluster.opendatahub.io/v1
kind: DataScienceCluster
metadata:
  name: test-cluster-invalid
spec:
  components:
    dashboard:
      managementState: InvalidValue  # This should fail validation
EOF

# See what happens with invalid data
oc apply --dry-run=server -f test-invalid-dsc.yaml
```

**Learning Points:**
- How does schema validation catch errors?
- What's the difference between client and server validation?
- How helpful are the error messages?

---

## 🔍 Advanced Discovery Exercises

### Exercise 5: Explore CRD Relationships

```bash
# Look at owner references in action
oc get dsc -o yaml | grep -A 10 ownerReferences
oc get dashboard -o yaml | grep -A 10 ownerReferences

# Check resource hierarchy
oc get dsc,dsci,dashboard,workbenches --show-labels

# Understand scope differences
oc api-resources | grep datasciencecluster
oc api-resources | grep dashboard
```

### Exercise 6: Compare with Core Kubernetes

```bash
# Compare CRD structure with core resources
oc explain pod
oc explain deployment
oc explain datasciencecluster

# Look at the patterns
oc get pod -o yaml | head -30
oc get deployment -o yaml | head -30
oc get dsc -o yaml | head -30
```

**Pattern Recognition:**
- How do custom resources follow the same patterns as core resources?
- What's universal across all Kubernetes resources?

---

## 🧠 Real-World Insights

### What You're Discovering:

1. **CRDs are Resource Templates**: They define the shape and validation rules for custom resources
2. **Schema-Driven APIs**: OpenAPI v3 schemas provide rich validation and documentation
3. **Consistent Patterns**: Custom resources follow the same metadata/spec/status pattern as core resources
4. **Hierarchical Management**: DataScienceCluster manages component resources through owner references

### Connection to ODH Architecture:

```
CRD Definition (in cluster) → Custom Resource (your YAML) → Controller (makes it real)
       ↓                              ↓                           ↓
  Validation Rules              User's Desired State         Actual Implementation
```

---

## 💡 Key Discoveries

### CRD Power:
- **Extensibility**: Add new resource types without modifying Kubernetes
- **Validation**: Rich schema validation prevents configuration errors
- **Integration**: All Kubernetes tooling works automatically
- **Documentation**: Self-documenting APIs through schemas

### ODH's CRD Strategy:
- **Multiple API Groups**: Logical organization (cluster, components, services)
- **Version Progression**: Start with v1alpha1, mature to v1
- **Scope Decisions**: Cluster-scoped for infrastructure, namespaced for applications
- **Rich Schemas**: Detailed validation and defaults

---

## 📋 Cleanup

```bash
# Remove test files
rm -f test-valid-dsc.yaml test-invalid-dsc.yaml dsc-crd.yaml dashboard-crd.yaml
```

---

## 🎯 Connection to Your Learning Journey

### What You Now Understand:
- **Day 1**: Kubernetes APIs and how they work
- **Day 2**: How CRDs extend those APIs with custom resource types

### The Big Picture:
Every time you see a custom resource like `DataScienceCluster`, there's:
1. **A CRD** that defines its structure and validation
2. **A Controller** that watches for changes and implements behavior
3. **The Kubernetes API** that handles storage and access

### Ready for Day 3:
Tomorrow you'll learn about the **controllers** - the brain that watches your custom resources and makes them actually do something!

---

## ⏰ Time Tracking

- **Hands-on Exercises**: _____ minutes (target: 20)
- **Key Insights Gained**: _____
- **Most Interesting Discovery**: _____

---

## 📝 Notes for Progress Tracker

**What I learned about CRDs:**

**How ODH uses CRDs:**

**Questions for Day 3:**

**Coolest thing I discovered:**