// Copyright 2024 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package report

import (
	"regexp"
	"strings"
)

type xv6 struct {
	*config
	// XV6-specific crash patterns
	kernelPanicRe *regexp.Regexp
	assertFailRe  *regexp.Regexp
	segfaultRe    *regexp.Regexp
	oopsRe        *regexp.Regexp
}

// XV6 crash patterns - XV6 is much simpler than Linux
var xv6CrashPatterns = []*regexp.Regexp{
	// Kernel panic patterns
	regexp.MustCompile(`panic: (.+)`),
	regexp.MustCompile(`PANIC: (.+)`),

	// Assertion failures
	regexp.MustCompile(`assertion failed: (.+)`),
	regexp.MustCompile(`assert\((.+)\) failed`),

	// Page faults and memory errors
	regexp.MustCompile(`page fault: (.+)`),
	regexp.MustCompile(`segmentation fault: (.+)`),
	regexp.MustCompile(`invalid memory access: (.+)`),

	// Stack overflow
	regexp.MustCompile(`stack overflow`),

	// Deadlock detection (if XV6 has it)
	regexp.MustCompile(`deadlock detected`),

	// General errors
	regexp.MustCompile(`kernel error: (.+)`),
	regexp.MustCompile(`fatal error: (.+)`),
}

func ctorXV6(cfg *config) (reporterImpl, []string, error) {
	ctx := &xv6{
		config:        cfg,
		kernelPanicRe: regexp.MustCompile(`panic: (.+)`),
		assertFailRe:  regexp.MustCompile(`assertion failed: (.+)`),
		segfaultRe:    regexp.MustCompile(`segmentation fault: (.+)`),
		oopsRe:        regexp.MustCompile(`kernel error: (.+)`),
	}

	return ctx, nil, nil
}

func (ctx *xv6) ContainsCrash(output []byte) bool {
	// Check for XV6-specific crash patterns
	for _, re := range xv6CrashPatterns {
		if re.Match(output) {
			return true
		}
	}
	return false
}

func (ctx *xv6) Parse(output []byte) *Report {
	rep := &Report{
		Output: output,
	}

	// XV6 output is usually much simpler than Linux
	// Look for panic, assertion failures, etc.

	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Check for kernel panic
		if match := ctx.kernelPanicRe.FindStringSubmatch(line); match != nil {
			rep.Title = "XV6 kernel panic: " + match[1]
			rep.Report = extractXV6Context(lines, i, 10)
			rep.StartPos = len(strings.Join(lines[:i], "\n"))
			rep.EndPos = rep.StartPos + len(rep.Report)
			return rep
		}

		// Check for assertion failures
		if match := ctx.assertFailRe.FindStringSubmatch(line); match != nil {
			rep.Title = "XV6 assertion failed: " + match[1]
			rep.Report = extractXV6Context(lines, i, 10)
			rep.StartPos = len(strings.Join(lines[:i], "\n"))
			rep.EndPos = rep.StartPos + len(rep.Report)
			return rep
		}

		// Check for segfaults
		if match := ctx.segfaultRe.FindStringSubmatch(line); match != nil {
			rep.Title = "XV6 segmentation fault: " + match[1]
			rep.Report = extractXV6Context(lines, i, 10)
			rep.StartPos = len(strings.Join(lines[:i], "\n"))
			rep.EndPos = rep.StartPos + len(rep.Report)
			return rep
		}

		// Check for other kernel errors
		if match := ctx.oopsRe.FindStringSubmatch(line); match != nil {
			rep.Title = "XV6 kernel error: " + match[1]
			rep.Report = extractXV6Context(lines, i, 10)
			rep.StartPos = len(strings.Join(lines[:i], "\n"))
			rep.EndPos = rep.StartPos + len(rep.Report)
			return rep
		}

		// Look for stack traces (XV6 might print simple stack traces)
		if strings.Contains(line, "backtrace:") || strings.Contains(line, "stack trace:") {
			rep.Title = "XV6 stack trace"
			rep.Report = extractXV6StackTrace(lines, i)
			rep.StartPos = len(strings.Join(lines[:i], "\n"))
			rep.EndPos = rep.StartPos + len(rep.Report)
			return rep
		}
	}

	// If no specific crash pattern found but output looks suspicious
	if containsSuspiciousXV6Output(string(output)) {
		rep.Title = "XV6 suspicious output"
		rep.Report = output
		rep.StartPos = 0
		rep.EndPos = len(output)
	}

	return rep
}

func (ctx *xv6) Symbolize(rep *Report) error {
	// XV6 symbolization is much simpler than Linux
	// XV6 usually doesn't have complex symbol resolution
	// For now, just try basic address-to-symbol mapping if available

	if ctx.kernelDirs.Obj == "" {
		return nil
	}

	// TODO: Implement XV6-specific symbolization
	// XV6 might have simpler symbol files or debugging info
	// This would depend on XV6's debugging capabilities

	return nil
}

// Extract context around a crash line
func extractXV6Context(lines []string, crashLine, contextLines int) []byte {
	start := crashLine - contextLines
	if start < 0 {
		start = 0
	}

	end := crashLine + contextLines + 1
	if end > len(lines) {
		end = len(lines)
	}

	context := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		if i == crashLine {
			context = append(context, ">>> "+lines[i]+" <<<")
		} else {
			context = append(context, lines[i])
		}
	}

	return []byte(strings.Join(context, "\n"))
}

// Extract stack trace for XV6
func extractXV6StackTrace(lines []string, startLine int) []byte {
	trace := make([]string, 0)

	// Add the stack trace header
	if startLine < len(lines) {
		trace = append(trace, lines[startLine])
	}

	// Look for subsequent lines that look like stack frames
	for i := startLine + 1; i < len(lines) && i < startLine+20; i++ {
		line := strings.TrimSpace(lines[i])

		// XV6 stack traces might look like:
		// - hex addresses
		// - function names
		// - simple address+offset format
		if line == "" {
			break
		}

		if isXV6StackFrame(line) {
			trace = append(trace, lines[i])
		} else {
			break
		}
	}

	return []byte(strings.Join(trace, "\n"))
}

// Check if a line looks like an XV6 stack frame
func isXV6StackFrame(line string) bool {
	line = strings.TrimSpace(line)

	// XV6 stack frames might be simple hex addresses
	if len(line) > 2 && strings.HasPrefix(line, "0x") {
		return true
	}

	// Or function names with addresses
	if strings.Contains(line, "+") && strings.Contains(line, "0x") {
		return true
	}

	// Or simple function names
	if strings.Contains(line, "()") {
		return true
	}

	return false
}

// Check if output contains suspicious XV6 patterns
func containsSuspiciousXV6Output(output string) bool {
	suspicious := []string{
		"trap",
		"interrupt",
		"exception",
		"fault",
		"error",
		"warning",
		"corruption",
		"invalid",
		"illegal",
		"unexpected",
	}

	lowercaseOutput := strings.ToLower(output)
	for _, pattern := range suspicious {
		if strings.Contains(lowercaseOutput, pattern) {
			return true
		}
	}

	return false
}

// Classify the type of XV6 crash
func classifyXV6Crash(title string) string {
	title = strings.ToLower(title)

	if strings.Contains(title, "panic") {
		return "kernel-panic"
	}
	if strings.Contains(title, "assertion") {
		return "assertion-failure"
	}
	if strings.Contains(title, "segmentation") || strings.Contains(title, "segfault") {
		return "memory-error"
	}
	if strings.Contains(title, "stack") {
		return "stack-error"
	}
	if strings.Contains(title, "deadlock") {
		return "deadlock"
	}

	return "unknown"
}

// Get relevant XV6 source files for a crash
func getXV6RelevantFiles(report *Report) []string {
	// XV6 has a smaller codebase, so this is simpler
	coreFiles := []string{
		"kernel/main.c",
		"kernel/vm.c",
		"kernel/proc.c",
		"kernel/syscall.c",
		"kernel/trap.c",
		"kernel/fs.c",
		"kernel/bio.c",
		"kernel/sleeplock.c",
		"kernel/spinlock.c",
	}

	// For specific crash types, suggest more specific files
	if report.Title != "" {
		title := strings.ToLower(report.Title)

		if strings.Contains(title, "vm") || strings.Contains(title, "memory") || strings.Contains(title, "page") {
			return []string{"kernel/vm.c", "kernel/kalloc.c"}
		}
		if strings.Contains(title, "proc") || strings.Contains(title, "process") {
			return []string{"kernel/proc.c", "kernel/swtch.S"}
		}
		if strings.Contains(title, "syscall") {
			return []string{"kernel/syscall.c", "kernel/sysproc.c"}
		}
		if strings.Contains(title, "fs") || strings.Contains(title, "file") {
			return []string{"kernel/fs.c", "kernel/file.c", "kernel/bio.c"}
		}
		if strings.Contains(title, "lock") {
			return []string{"kernel/spinlock.c", "kernel/sleeplock.c"}
		}
	}

	return coreFiles
}
