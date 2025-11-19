# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **30-day structured study plan** for mastering Kubernetes Operator development using the OpenDataHub (ODH) Operator as a real-world reference. It's a learning repository, not a development codebase.

**Technology Focus**: Kubernetes Operators, Kubebuilder, Controller-Runtime, Go programming, OLM (Operator Lifecycle Manager)

## Repository Structure

- **Daily Learning Modules**: `day01/` through `day30/` - Each contains structured learning materials
- **Quick Reference**: `quick_reference.md` - Commands, patterns, and troubleshooting guide
- **Study Plan**: `30_day_study_plan.md` - Complete overview of the 30-day curriculum

### Daily Module Structure
Each `dayXX/` directory contains:
- `README.md` - Overview, goals, and time allocation
- `dayXX_complete_guide.md` - Comprehensive study guide (40-45 minutes)
- `dayXX_live_exercises.md` - Hands-on exercises (15-20 minutes)
- `dayXX_exercises.md` - Alternative offline exercises
- Supporting materials (CRDs, reference docs, code examples)

## Study Methodology

**Daily Routine (1 hour/day)**:
1. Navigate to current day's folder
2. Read the day's README for overview and goals
3. Follow the study guide for concepts (40-50 minutes)
4. Complete hands-on exercises (10-20 minutes)

## Study Context

When working in this repository:

1. **Follow Structure**: Respect the systematic progression through concepts
2. **Reference ODH Code**: Use `/Users/suksubra/Documents/Work/RHOAI/opendatahub-operator` as the primary codebase reference
3. **Focus on Learning**: Each day is designed as a focused 1-hour learning session

## Key Learning Path

### Week 1: Foundations (Days 1-7)
- Kubernetes API fundamentals
- Custom Resource Definitions (CRDs)
- Controllers and reconciliation loops
- Go programming for operators
- Client-go library basics
- Controller-runtime framework

### Week 2: Kubebuilder & Controller Framework (Days 8-14)
- Kubebuilder introduction and code generation
- ODH controller architecture deep dive
- Component management patterns
- Reconciler implementation
- Event watching and filtering

### Week 3: CRDs, Webhooks & OLM (Days 15-21)
- Advanced CRD features and validation
- Admission webhooks (mutating/validating)
- Operator Lifecycle Manager (OLM) packaging
- Bundle creation and scorecard testing

### Week 4: Integration & Production (Days 22-28)
- Testing strategies for operators
- Monitoring and observability
- Manifest management with Kustomize
- Security, RBAC, and production deployment
- CI/CD and release automation

### Bonus Days (Days 29-30)
- Service mesh integration
- Capstone project planning

## Essential Commands for Study

### Learning Progress Management
```bash
# Navigate to study materials
ls day*/

# Navigate to current day
cd dayXX/

# Quick reference lookup
grep -n "pattern" quick_reference.md
```

### ODH Operator Development (for reference)
```bash
# From ODH operator directory
make generate          # Generate code and manifests
make build             # Build operator binary
make test              # Run tests
make install           # Install CRDs
make run               # Run locally (outside cluster)

# OLM Bundle operations
make bundle            # Generate bundle
make bundle-build      # Build bundle image
operator-sdk bundle validate ./bundle  # Validate bundle
operator-sdk scorecard bundle          # Run scorecard tests
```

### Kubernetes Commands (for exercises)
```bash
# CRD inspection
kubectl get crd | grep opendatahub
kubectl describe crd datascienceclusters.datasciencecluster.opendatahub.io

# Resource management
kubectl get datasciencecluster -o yaml
kubectl describe datasciencecluster <name>

# Debugging
kubectl logs -n opendatahub-operator-system deployment/opendatahub-operator-controller-manager
kubectl get events --sort-by='.metadata.creationTimestamp'
```

## Study Tips

### Code Reading Approach
1. Start with type definitions in ODH `api/` directories
2. Follow controller setup in `controllers/`
3. Trace reconciliation logic step by step
4. Check tests for usage examples
5. Use `quick_reference.md` for patterns and troubleshooting

### Study Focus
- Each session is designed for 1-hour focused learning
- Take notes on key insights and questions as needed
- Reference ODH codebase for practical examples

### Integration with ODH Codebase
- Always reference real ODH code examples when studying concepts
- Use the operator codebase at `/Users/suksubra/Documents/Work/RHOAI/opendatahub-operator`
- Cross-reference study materials with actual implementation patterns
- Apply theoretical concepts to real-world operator code

## Learning Objectives

By completion, this study plan develops expertise in:
- **Technical Skills**: Kubernetes operator development, Kubebuilder, Controller-runtime, CRD design, webhooks, OLM packaging
- **ODH-Specific Knowledge**: Component architecture, action-based reconciliation, manifest management, multi-CRD coordination
- **Production Skills**: Testing strategies, monitoring, security, CI/CD, performance optimization

## Notes for Claude Code

- This is a **learning repository**, not a development project
- Prioritize understanding and educational support over code generation
- Help maintain study structure and learning flow
- Reference the actual ODH operator codebase for real-world examples
- Support systematic learning progression through the 30-day curriculum
- Assist with understanding complex operator patterns and Kubernetes concepts