# Day 15: Advanced CRD Features - Live Exercises

## Overview
These hands-on exercises help you explore advanced CRD features using the OpenDataHub operator codebase. You'll examine real schema validation, default value implementation, and subresource patterns.

**Total Time**: 15 minutes
**Prerequisites**: Access to ODH operator codebase at `/Users/suksubra/Documents/Work/RHOAI/opendatahub-operator`

---

## Exercise 1: CRD Schema Validation Analysis (5 minutes)

### Goal
Examine how ODH implements comprehensive schema validation in the DataScienceCluster CRD.

### Steps

**1. Examine the Generated CRD Schema (2 minutes)**
```bash
# Navigate to ODH operator directory
cd /Users/suksubra/Documents/Work/RHOAI/opendatahub-operator

# Look at the main DataScienceCluster CRD
cat config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml
```

**Key Things to Find:**
- [ ] OpenAPI v3 schema structure (`openAPIV3Schema`)
- [ ] Enum validation for management states
- [ ] Default values assignment
- [ ] Required vs optional field distinctions
- [ ] Complex nested object validation

**2. Find Validation Patterns (3 minutes)**
```bash
# Search for validation patterns in the CRD
grep -A 10 -B 2 "enum:" config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml

# Find default value assignments
grep -A 5 -B 2 "default:" config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml

# Look for pattern validation
grep -A 3 -B 2 "pattern:" config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml
```

### Questions to Answer
1. What are the valid values for component `managementState` fields?
2. Which fields have default values, and what are they?
3. How does ODH validate resource specifications (CPU, memory)?

---

## Exercise 2: Go Type Validation Markers (4 minutes)

### Goal
Understand how kubebuilder markers in Go source code generate CRD validation rules.

### Steps

**1. Examine Go Type Definitions (2 minutes)**
```bash
# Look at the main DataScienceCluster types
cat api/datasciencecluster/v1/datasciencecluster_types.go | head -100

# Search for kubebuilder validation markers
grep -n "+kubebuilder:validation" api/datasciencecluster/v1/datasciencecluster_types.go

# Find default value markers
grep -n "+kubebuilder:default" api/datasciencecluster/v1/datasciencecluster_types.go
```

**2. Compare Types Across Versions (2 minutes)**
```bash
# Compare v1 vs v2 API differences
diff -u api/datasciencecluster/v1/datasciencecluster_types.go \
        api/datasciencecluster/v2/datasciencecluster_types.go | head -50

# Look for version-specific markers
grep -n "+kubebuilder" api/datasciencecluster/v2/datasciencecluster_types.go | head -10
```

### Discovery Tasks
- [ ] Find examples of `+kubebuilder:validation:Enum` markers
- [ ] Identify fields with `+kubebuilder:default` values
- [ ] Locate `+kubebuilder:validation:Pattern` for format validation
- [ ] Compare how v1 and v2 handle the same concepts

### Questions to Answer
1. How do kubebuilder markers translate to OpenAPI schema rules?
2. What validation markers does ODH use most frequently?
3. How do default values in Go code appear in the generated CRD?

---

## Exercise 3: CRD Subresources and Printer Columns (3 minutes)

### Goal
Analyze how ODH implements status subresources and custom kubectl output.

### Steps

**1. Examine Subresource Configuration (1.5 minutes)**
```bash
# Find subresource definitions in CRDs
grep -A 10 -B 5 "subresources:" config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml

# Look at additional printer columns
grep -A 15 "additionalPrinterColumns:" config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml
```

**2. Analyze Status Structure (1.5 minutes)**
```bash
# Examine status type definition
grep -A 30 "type DataScienceClusterStatus struct" api/datasciencecluster/v1/datasciencecluster_types.go

# Find status condition patterns
grep -A 10 -B 5 "Conditions" api/datasciencecluster/v1/datasciencecluster_types.go
```

### Exploration Checklist
- [ ] Identify what subresources are enabled (status, scale)
- [ ] Find custom kubectl columns (Ready, Reason, etc.)
- [ ] Understand status condition types
- [ ] Locate JSONPath expressions for printer columns

### Questions to Answer
1. What information does `kubectl get dsc` display by default?
2. How are status conditions structured in ODH?
3. What makes the status subresource special compared to spec?

---

## Exercise 4: Multi-Version Support Analysis (3 minutes)

### Goal
Understand how ODH manages multiple CRD versions and conversion.

### Steps

**1. Compare Version Definitions (1.5 minutes)**
```bash
# Check version configuration in CRD
grep -A 20 "versions:" config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml

# Find storage version settings
grep -A 5 -B 5 "storage:" config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml

# Look for conversion strategy
grep -A 10 "conversion:" config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml
```

**2. Analyze Version Differences (1.5 minutes)**
```bash
# Check if both versions are served
grep "served:" config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml

# Look for version-specific features
ls api/datasciencecluster/*/
```

### Investigation Points
- [ ] Which versions are served vs storage versions?
- [ ] What conversion strategy is used?
- [ ] Are there schema differences between versions?
- [ ] How does ODH handle version migration?

### Questions to Answer
1. Which DataScienceCluster version is the storage version?
2. What conversion strategy does ODH use?
3. How do v1 and v2 schemas differ?

---

## Bonus Exploration: Component CRDs (Optional)

If you finish early, explore component-specific CRDs:

```bash
# List all ODH CRDs
ls config/crd/bases/ | grep -v datasciencecluster

# Pick one component CRD and analyze its patterns
cat config/crd/bases/components.platform.opendatahub.io_workbenches.yaml | grep -A 5 -B 5 "validation\|default\|enum"
```

---

## Verification Commands

Test your understanding with these verification commands:

```bash
# Generate fresh CRDs from source (if make is available)
make generate

# Validate CRD syntax
kubectl apply --dry-run=client -f config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml

# Check for kubebuilder markers
find api/ -name "*.go" -exec grep -l "+kubebuilder:" {} \;
```

---

## Key Observations to Document

As you complete these exercises, note:

### Schema Validation Patterns
- How ODH uses enum validation for management states
- Pattern validation for resource specifications
- Required vs optional field strategies
- Default value assignment patterns

### Subresource Implementation
- Status subresource configuration
- Custom kubectl output design
- Condition-based status reporting
- JSONPath expressions for fields

### Version Management
- Multi-version serving strategy
- Storage version selection
- Conversion approach (None vs Webhook)
- Schema evolution patterns

### Kubebuilder Integration
- Marker-based code generation
- Go type to OpenAPI translation
- Validation rule propagation
- Default value handling

---

## Troubleshooting

**If you can't find the ODH codebase:**
```bash
# Verify the path exists
ls /Users/suksubra/Documents/Work/RHOAI/opendatahub-operator/

# If not available, you can reference the online repository
# or use generic examples from kubebuilder documentation
```

**If CRD files seem empty:**
```bash
# CRDs might need generation
cd /Users/suksubra/Documents/Work/RHOAI/opendatahub-operator/
make generate
```

**Quick Reference Commands:**
```bash
# Find all validation markers
grep -r "+kubebuilder:validation" api/

# Find all CRD files
find . -name "*.yaml" -path "*/crd/*"

# Search for specific patterns
grep -r "managementState" api/ | head -5
```

These exercises provide hands-on experience with the advanced CRD features that make ODH operator robust and user-friendly. The patterns you discover will inform your understanding of production operator design.