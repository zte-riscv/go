# Contribution Guidelines for ZTE & ByteDance Go RISC-V Repository
**English Version** | [中文版本](README-CN.md)

Repository: https://github.com/zte-riscv/go

## 1. Introduction
This specification applies to the Go RISC-V repository jointly maintained by ZTE and ByteDance. To ensure code quality and collaboration efficiency, all merge requests must comply with these guidelines.

## 2. Pre-Submission Preparation
### 2.1 Issue Tracking and Discussion
**Mandatory**:
- Create or claim a related Issue before commencing substantive work
- Clearly describe the change motivation, design approach, and expected impact in the Issue

**Recommended**:
- For significant changes (e.g., new instruction set support, architectural modifications), submit a design document for prior review

### 2.2 Branch Management
```bash
# Create a feature branch from the current development branch (e.g., go1.25.0-zte-dev):
git fetch origin
git checkout -b your_dev_branch origin/go1.25.0-zte-dev
```

## 3. Code Contribution Process
### 3.1 Development Workflow
**Issue Submission**:
- For bug reports: Document the problem, environment, and reproduction steps
- For features/optimizations: Describe the technical background and implementation strategy

**Pull Request Submission**:
- Provide task context in the PR description and reference related Issues
- Resolve all merge conflicts locally before submission

**CI and Review**:
- Automated test suites will execute upon PR submission (failed tests block merging)
- Require approvals from designated reviewers before merging, reviewers comment +1 or +2.

**Merge Policy**:
- All commits must be squashed into a single commit

### 3.2 Commit Message Standards
For squashed commits, follow this format:
```bash
git commit -m "package: concise change summary

Comprehensive change description covering:
- Change rationale
- Technical implementation
- Performance impacts (if applicable)
- Related issue references

Fixes #12345
Updates #67890"
```
- **Header**: Package prefix + brief description (<50 chars)
- **Body**: Complete explanation of changes
- **Footer**: Mandatory issue references

## 4. Review Process
### 4.1 Mandatory Requirements
- CI pipeline must pass
- Minimum approvals:
- One +2 from ZTE core maintainers
- One +2 from ByteDance core maintainers

### 4.2 Review Tiers
- `+1`: Preliminary approval (requires additional reviewer)
- `+2`: Final approval for merging

**For maintainers and reviewers: 
when approving the merge, please respond with either +1 or +2.**

### 4.3 Core Maintainers
| GitHub | Role | Org | Auth Level |
|--------|------|-----|-----------|
| @agiledragon | Maintainer | ZTE | +2 |
| @ctk-1998 | Reviewer | ZTE | +1 |
| @lxq015 | Reviewer | ZTE | +1 |
| @hehongjun20110618 | Maintainer | ZTE | +2 |
| @newborn22 | Maintainer | ZTE | +2 |
| @wenchangping | Reviewer | ZTE | +1 |
| @wangpc-pp | Maintainer | ByteDance | +2 |
| @BoyaoWang430 | Reviewer | ByteDance | +1 |

### 4.4 Review Checklist
- ✅ Go coding standards compliance
- ✅ Passing CI pipelines
- ✅ Corresponding test coverage
- ✅ Performance benchmarks (where applicable)

## 5. CI Pipeline Requirements
### 5.1 Test
- **Static analysis**: `go vet` for `/src`
- **Codegen validation**: Instruction verification via `asmcheck` (`/test/codegen/`)
- **Assembly tests**: Machine code validation (`/testdata/riscv64.s`)
- **Regression testing**: `all.bash` execution on x86/ARM
- **Application tests**: QEMU-RISCV validation (`/test/zte/*_test.go`)

## 6. Performance Reporting (Optional)
### 6.1 Benchmark Requirements
For performance-sensitive changes:

Test Environment:
```markdown
- Hardware: [e.g., SiFive Unmatched]
- CPU: [Cores, frequency, RISC-V extensions]
- Memory:
- Go Version:
- Test Date:
```
Benchmark Results, `benchstat` comparison:
```bash
benchstat old.txt new.txt
nameold time/opnew time/opdelta
Fibonacci125ms ± 2%98ms ± 3%-21.60%
Sort456ms ± 1%401ms ± 2%-12.06%
```

## 7. Appendix
Official Go Contribution Guide: https://golang.org/doc/contribute
