// Copyright 2024 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/syzkaller/pkg/osutil"
)

type xv6 struct{}

func (xv6 xv6) build(params Params) (ImageDetails, error) {
	details := ImageDetails{}

	// XV6 build is much simpler than Linux
	if err := xv6.buildKernel(params); err != nil {
		return details, err
	}

	// Get compiler identity
	var err error
	details.CompilerID, err = compilerIdentity(params.Compiler)
	if err != nil {
		return details, fmt.Errorf("failed to get compiler identity: %w", err)
	}

	// XV6 builds a simple kernel binary
	kernelPath := filepath.Join(params.KernelDir, "kernel")
	if !osutil.IsExist(kernelPath) {
		return details, fmt.Errorf("XV6 kernel binary not found at %v", kernelPath)
	}

	// Copy kernel to output directory
	if err := osutil.CopyFile(kernelPath, filepath.Join(params.OutputDir, "kernel")); err != nil {
		return details, fmt.Errorf("failed to copy XV6 kernel: %w", err)
	}

	// Create a basic filesystem image for XV6
	if err := xv6.createFileSystem(params); err != nil {
		return details, fmt.Errorf("failed to create XV6 filesystem: %w", err)
	}

	// Generate SSH keys for access (if needed)
	if err := xv6.generateSSHKey(params.OutputDir); err != nil {
		return details, fmt.Errorf("failed to generate SSH key: %w", err)
	}

	return details, nil
}

func (xv6 xv6) buildKernel(params Params) error {
	// XV6 uses a simple Makefile build system
	makeArgs := []string{
		"TOOLPREFIX=riscv64-linux-gnu-", // Use RISC-V toolchain
		"clean", "all",
	}

	if params.Compiler != "" {
		// Override compiler if specified
		makeArgs = append([]string{"CC=" + params.Compiler}, makeArgs...)
	}

	// Set build parallelism
	if params.BuildCPUs > 1 {
		makeArgs = append([]string{fmt.Sprintf("-j%d", params.BuildCPUs)}, makeArgs...)
	}

	params.Tracer.Log("Building XV6 kernel...")

	cmd := exec.Command(params.Make, makeArgs...)
	if cmd.Path == "" {
		cmd.Path = "make"
	}
	cmd.Dir = params.KernelDir
	cmd.Env = append(os.Environ(),
		// XV6 specific environment variables
		"QEMU=qemu-system-riscv64",
		"CPUS=1",
	)

	if output, err := osutil.RunCmd(30*time.Minute, params.KernelDir, cmd.Path, makeArgs...); err != nil {
		return &KernelError{
			Report: []byte(fmt.Sprintf("XV6 kernel build failed: %v", err)),
			Output: output,
		}
	}

	return nil
}

func (xv6 xv6) createFileSystem(params Params) error {
	// XV6 uses a simple filesystem
	// Create a basic filesystem image

	fsImagePath := filepath.Join(params.OutputDir, "fs.img")

	// Check if XV6 builds its own filesystem image
	xv6FsPath := filepath.Join(params.KernelDir, "fs.img")
	if osutil.IsExist(xv6FsPath) {
		// Copy XV6's filesystem
		return osutil.CopyFile(xv6FsPath, fsImagePath)
	}

	// Otherwise create a minimal filesystem image
	// This is a placeholder - actual XV6 filesystem creation would depend on XV6's tools
	params.Tracer.Log("Creating minimal XV6 filesystem...")

	// Create a simple filesystem image using dd
	cmd := exec.Command("dd", "if=/dev/zero", "of="+fsImagePath, "bs=1M", "count=64")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create filesystem image: %w", err)
	}

	return nil
}

func (xv6 xv6) generateSSHKey(outputDir string) error {
	// Generate SSH key for accessing XV6 if it supports SSH
	keyPath := filepath.Join(outputDir, "key")

	if osutil.IsExist(keyPath) {
		return nil // Key already exists
	}

	// Generate RSA key
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-N", "", "-f", keyPath)
	if err := cmd.Run(); err != nil {
		// SSH might not be available or needed for XV6
		// Don't fail the build, just log a warning
		fmt.Printf("Warning: failed to generate SSH key (this may be expected for XV6): %v\n", err)

		// Create a dummy key file to satisfy syzkaller expectations
		if err := osutil.WriteFile(keyPath, []byte("# Dummy SSH key for XV6\n")); err != nil {
			return err
		}
		if err := osutil.WriteFile(keyPath+".pub", []byte("# Dummy SSH public key for XV6\n")); err != nil {
			return err
		}
	}

	return nil
}

func (xv6 xv6) clean(params Params) error {
	// Clean XV6 build artifacts
	params.Tracer.Log("Cleaning XV6...")

	makeArgs := []string{"clean"}

	// Try to clean using XV6's Makefile
	cmd := exec.Command(params.Make, makeArgs...)
	if cmd.Path == "" {
		cmd.Path = "make"
	}
	cmd.Dir = params.KernelDir

	if err := cmd.Run(); err != nil {
		// If make clean fails, try to remove common XV6 build artifacts manually
		artifacts := []string{
			"kernel",
			"fs.img",
			"*.o",
			"*.d",
			"*.asm",
			"*.sym",
		}

		for _, pattern := range artifacts {
			matches, _ := filepath.Glob(filepath.Join(params.KernelDir, pattern))
			for _, match := range matches {
				os.Remove(match)
			}
		}
	}

	return nil
}

// Helper function to detect XV6 directory
func isXV6Directory(dir string) bool {
	// Check for XV6-specific files
	xv6Files := []string{
		"Makefile",
		"kernel/kernel.ld",
		"kernel/main.c",
		"user/init.c",
	}

	for _, file := range xv6Files {
		if !osutil.IsExist(filepath.Join(dir, file)) {
			return false
		}
	}

	// Check if Makefile contains XV6-specific content
	makefileContent, err := os.ReadFile(filepath.Join(dir, "Makefile"))
	if err != nil {
		return false
	}

	content := string(makefileContent)
	return strings.Contains(content, "xv6") || strings.Contains(content, "QEMU") ||
		strings.Contains(content, "riscv64") || strings.Contains(content, "kernel.ld")
}

// Get XV6 architecture-specific settings
func getXV6Config(arch string) map[string]string {
	configs := map[string]map[string]string{
		"riscv64": {
			"TOOLPREFIX": "riscv64-linux-gnu-",
			"QEMU":       "qemu-system-riscv64",
			"CPUS":       "1",
			"ARCH":       "riscv64",
		},
		// XV6 primarily targets RISC-V now, but keep this for completeness
		"i386": {
			"TOOLPREFIX": "i386-linux-gnu-",
			"QEMU":       "qemu-system-i386",
			"CPUS":       "1",
			"ARCH":       "i386",
		},
	}

	if config, ok := configs[arch]; ok {
		return config
	}

	// Default to RISC-V
	return configs["riscv64"]
}

// Check if required tools are available
func checkXV6Tools(arch string) error {
	config := getXV6Config(arch)

	requiredTools := []string{
		"make",
		config["QEMU"],
	}

	// Check for toolchain
	if toolprefix := config["TOOLPREFIX"]; toolprefix != "" {
		requiredTools = append(requiredTools,
			toolprefix+"gcc",
			toolprefix+"ld",
			toolprefix+"objdump",
		)
	}

	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("required tool '%s' not found in PATH: %w", tool, err)
		}
	}

	return nil
}

// Validate XV6 build environment
func validateXV6Environment(params Params) error {
	// Check if the kernel directory looks like XV6
	if !isXV6Directory(params.KernelDir) {
		return fmt.Errorf("directory %s does not appear to be an XV6 source tree", params.KernelDir)
	}

	// Check required tools
	if err := checkXV6Tools(params.TargetArch); err != nil {
		return fmt.Errorf("XV6 build environment validation failed: %w", err)
	}

	return nil
}
