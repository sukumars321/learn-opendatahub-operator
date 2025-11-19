# Day 9: Hands-on Exercises - Kubebuilder Markers and Code Generation

## 🎯 Exercise Overview
**Time**: 20 minutes
**Goal**: Explore ODH codebase to find and understand Kubebuilder markers in action

---

## 🔧 Setup (2 minutes)

### Prerequisites
- ODH operator source code cloned locally
- Basic understanding of Go syntax
- Terminal/command line access

### Quick Environment Check
```bash
# Navigate to ODH operator directory
cd /path/to/opendatahub-operator

# Verify you're in the right place
ls -la | grep -E "(Makefile|main.go|controllers|apis)"

# Check if controller-gen is available
make help | grep generate
```

---

## 🕵️ Exercise 1: Marker Hunting (8 minutes)

### 1.1 Find RBAC Markers (3 minutes)
Let's hunt for RBAC markers in the controller files:

```bash
# Search for RBAC markers in controller files
grep -r "//+kubebuilder:rbac" controllers/

# Expected output should show lines like:
# //+kubebuilder:rbac:groups=datasciencecluster.opendatahub.io,resources=datascienceclusters,verbs=get;list;watch;create;update;patch;delete
```

**Questions to Explore:**
1. How many different API groups do the controllers need access to?
2. What's the most common verb pattern you see?
3. Which controllers have the most RBAC markers?

### 1.2 Explore CRD Markers (3 minutes)
Look for CRD generation markers in type definitions:

```bash
# Find object and validation markers
grep -r "//+kubebuilder:object\|//+kubebuilder:validation" apis/

# Look for printcolumn markers (these control kubectl output)
grep -r "//+kubebuilder:printcolumn" apis/

# Find subresource markers
grep -r "//+kubebuilder:subresource" apis/
```

**Questions to Explore:**
1. Which types have status subresources?
2. What custom columns are defined for kubectl output?
3. What validation rules are most commonly used?

### 1.3 Discover Webhook Markers (2 minutes)
Search for webhook configuration markers:

```bash
# Find webhook markers
grep -r "//+kubebuilder:webhook" controllers/

# Look for webhook-related files
find . -name "*webhook*" -type f
```

**Questions to Explore:**
1. How many webhooks does ODH define?
2. Are they mutating or validating webhooks?
3. What resources do they target?

---

## 🏗️ Exercise 2: Code Generation in Action (6 minutes)

### 2.1 Examine Generated Files (3 minutes)
Look at what the markers have generated:

```bash
# Check the generated CRD files
ls -la config/crd/bases/
cat config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml | head -30

# Look at generated RBAC
cat config/rbac/role.yaml | head -20

# Find generated DeepCopy methods
find . -name "zz_generated.deepcopy.go"
```

**Analysis Points:**
1. How does the CRD structure relate to the Go struct definitions?
2. Which permissions are in the generated ClusterRole?
3. What methods are in the DeepCopy files?

### 2.2 Run Code Generation (3 minutes)
**⚠️ WARNING**: Only do this if you have a clean git state and can easily revert changes!

```bash
# Check current git status
git status

# Run code generation
make generate

# See what changed
git diff --name-only

# Look at the differences
git diff config/rbac/role.yaml
```

**Observations to Make:**
1. Did any files change after running generation?
2. If changes occurred, what caused them?
3. How long did the generation process take?

---

## 🔍 Exercise 3: Marker Deep Dive (4 minutes)

### 3.1 Analyze a Specific Controller (2 minutes)
Pick one controller file and analyze its markers:

```bash
# Choose the DataScienceCluster controller
cat controllers/datasciencecluster_controller.go | grep "//+kubebuilder"

# Or use your favorite editor to examine the file
# Look at the markers at the top of the file
```

**Understanding Questions:**
1. What resources does this controller manage directly?
2. What Kubernetes core resources does it need access to?
3. Why does it need access to finalizers?

### 3.2 Connect Markers to Generated Content (2 minutes)
Trace how a marker becomes a manifest:

```bash
# Find a specific RBAC marker
grep -n "configmaps" controllers/datasciencecluster_controller.go

# Find the corresponding rule in the generated RBAC
grep -A5 -B5 "configmaps" config/rbac/role.yaml
```

**Connection Analysis:**
1. How does the marker syntax translate to YAML?
2. Are the verbs in the same order?
3. What additional metadata is added during generation?

---

## 📝 Exercise 4: Documentation Challenge (Optional Bonus)

If you finish early, try this challenge:

### Document Your Findings
Create a simple markdown file documenting:

1. **Marker Inventory**: List of all marker types you found
2. **Permission Summary**: What permissions ODH needs and why
3. **Generation Workflow**: Your understanding of the make generate process

```bash
# Create your analysis file
touch marker_analysis.md

# Use your favorite editor to document findings
```

---

## ✅ Exercise Checklist

By the end of these exercises, you should have:

- [ ] Found RBAC markers in controller files
- [ ] Located CRD markers in type definitions
- [ ] Discovered webhook markers (if any)
- [ ] Examined generated CRD files
- [ ] Looked at generated RBAC manifests
- [ ] Found generated DeepCopy methods
- [ ] Understood the relationship between markers and generated content
- [ ] Successfully run `make generate` (optional)

---

## 🤔 Reflection Questions

1. **Marker Purpose**: Why do you think ODH uses so many markers instead of writing manifests manually?

2. **Maintenance Benefits**: How do markers help when ODH needs to add new permissions or modify CRDs?

3. **Developer Experience**: What would happen if a developer forgot to run `make generate` after adding new markers?

4. **Code Organization**: How do the markers help keep the Go code and Kubernetes manifests in sync?

---

## 🚀 Bonus Exploration

If you want to go deeper:

### Advanced Marker Search
```bash
# Find all unique marker types
grep -r "//+kubebuilder:" . | sed 's/.*\/\/+kubebuilder:\([^:]*\).*/\1/' | sort | uniq

# Count marker usage
grep -r "//+kubebuilder:rbac" . | wc -l
grep -r "//+kubebuilder:validation" . | wc -l
```

### Understand the Build System
```bash
# Look at the Makefile targets
grep -A5 -B5 "generate:" Makefile

# Understand controller-gen
which controller-gen || echo "controller-gen not in PATH"
```

---

## 📚 Next Steps

Great job exploring ODH's use of Kubebuilder markers! Tomorrow in Day 10, we'll dive into ODH's controller architecture and see how all these generated manifests fit into the bigger picture.

**Key Takeaway**: Markers are the bridge between Go code and Kubernetes manifests, enabling ODH to maintain consistency and reduce manual errors across a complex operator codebase.