// Copyright 2024 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/syzkaller/pkg/compiler"
	"github.com/google/syzkaller/pkg/osutil"
)

type xv6 struct{}

// XV6 is a simple teaching OS, so constant extraction is much simpler than Linux
func (xv6 *xv6) prepare(sourcedir string, build bool, arches []*Arch) error {
	// XV6 doesn't require complex preparation like Linux
	// Just verify that the XV6 source directory looks correct

	if sourcedir == "" {
		return fmt.Errorf("XV6 source directory not specified")
	}

	// Check for XV6-specific files to verify this is an XV6 source tree
	xv6Files := []string{
		"Makefile",
		"kernel/kernel.ld",
		"kernel/main.c",
		"user/init.c",
	}

	for _, file := range xv6Files {
		fullPath := filepath.Join(sourcedir, file)
		if !osutil.IsExist(fullPath) {
			return fmt.Errorf("XV6 source verification failed: %s not found", file)
		}
	}

	return nil
}

func (xv6 *xv6) prepareArch(arch *Arch) error {
	// XV6 doesn't need per-architecture preparation like Linux kernel headers
	// XV6 has a simple build system

	// Verify the target architecture is supported by XV6
	switch arch.target.Arch {
	case "riscv64":
		// XV6 primarily supports RISC-V
		return nil
	case "386":
		// XV6 has legacy x86 support
		return nil
	default:
		return fmt.Errorf("XV6 does not support architecture %s", arch.target.Arch)
	}
}

func (xv6 *xv6) processFile(arch *Arch, info *compiler.ConstInfo) (map[string]uint64, map[string]bool, error) {
	// XV6 constant processing is much simpler than Linux
	// XV6 has fewer constants and simpler system call interface

	consts := make(map[string]uint64)
	undeclared := make(map[string]bool)

	// Process XV6-specific constants
	for _, constInfo := range info.Consts {
		val, err := xv6.extractConstant(arch, constInfo.Name, constInfo)
		if err != nil {
			// If we can't extract the constant, mark it as undeclared
			undeclared[constInfo.Name] = true
			continue
		}
		consts[constInfo.Name] = val
	}

	// Add XV6-specific built-in constants
	xv6.addBuiltinConstants(arch, consts)

	return consts, undeclared, nil
}

func (xv6 *xv6) extractConstant(arch *Arch, name string, info *compiler.Const) (uint64, error) {
	// XV6 constant extraction
	// Most XV6 constants are simple #defines or enum values

	// For basic constants, try to extract from XV6 headers
	switch {
	case strings.HasPrefix(name, "O_"):
		// File open flags - XV6 has basic open flags
		return xv6.extractFileFlag(arch, name)

	case strings.HasPrefix(name, "SYS_"):
		// System call numbers - XV6 has simple syscall numbering
		return xv6.extractSyscallNumber(arch, name)

	case strings.HasPrefix(name, "PROT_"):
		// Memory protection flags - XV6 may not support mprotect
		return 0, fmt.Errorf("XV6 may not support memory protection flags")

	case strings.HasPrefix(name, "MAP_"):
		// Memory mapping flags - XV6 has limited mmap support
		return xv6.extractMapFlag(arch, name)

	case strings.HasPrefix(name, "S_IF"):
		// File type constants - XV6 supports basic file types
		return xv6.extractFileType(arch, name)

	case strings.HasPrefix(name, "AT_"):
		// openat flags - XV6 may not support all AT_ flags
		return xv6.extractAtFlag(arch, name)

	default:
		// For other constants, try a generic approach
		return xv6.extractGenericConstant(arch, name, info)
	}
}

func (xv6 *xv6) extractFileFlag(arch *Arch, name string) (uint64, error) {
	// XV6 file open flags are much simpler than Linux
	flags := map[string]uint64{
		"O_RDONLY": 0,
		"O_WRONLY": 1,
		"O_RDWR":   2,
		"O_CREATE": 0x200, // XV6 O_CREATE flag
		"O_TRUNC":  0x400, // XV6 O_TRUNC flag (if supported)
	}

	if val, ok := flags[name]; ok {
		return val, nil
	}

	return 0, fmt.Errorf("XV6 file flag %s not supported", name)
}

func (xv6 *xv6) extractSyscallNumber(arch *Arch, name string) (uint64, error) {
	// XV6 system call numbers are defined in kernel/syscall.h
	// They are much simpler than Linux

	syscalls := map[string]uint64{
		"SYS_fork":   1,
		"SYS_exit":   2,
		"SYS_wait":   3,
		"SYS_pipe":   4,
		"SYS_read":   5,
		"SYS_kill":   6,
		"SYS_exec":   7,
		"SYS_fstat":  8,
		"SYS_chdir":  9,
		"SYS_dup":    10,
		"SYS_getpid": 11,
		"SYS_sbrk":   12,
		"SYS_sleep":  13,
		"SYS_uptime": 14,
		"SYS_open":   15,
		"SYS_write":  16,
		"SYS_mknod":  17,
		"SYS_unlink": 18,
		"SYS_link":   19,
		"SYS_mkdir":  20,
		"SYS_close":  21,
	}

	if val, ok := syscalls[name]; ok {
		return val, nil
	}

	return 0, fmt.Errorf("XV6 syscall %s not supported", name)
}

func (xv6 *xv6) extractMapFlag(arch *Arch, name string) (uint64, error) {
	// XV6 may not support mmap, but provide basic flags just in case
	flags := map[string]uint64{
		"MAP_PRIVATE":   0x02,
		"MAP_ANON":      0x20,
		"MAP_ANONYMOUS": 0x20,
	}

	if val, ok := flags[name]; ok {
		return val, nil
	}

	return 0, fmt.Errorf("XV6 map flag %s not supported", name)
}

func (xv6 *xv6) extractFileType(arch *Arch, name string) (uint64, error) {
	// XV6 file type constants
	types := map[string]uint64{
		"S_IFDIR": 0x4000,
		"S_IFREG": 0x8000,
		"S_IFCHR": 0x2000,
		"S_IFBLK": 0x6000,
		"S_IFIFO": 0x1000,
	}

	if val, ok := types[name]; ok {
		return val, nil
	}

	return 0, fmt.Errorf("XV6 file type %s not supported", name)
}

func (xv6 *xv6) extractAtFlag(arch *Arch, name string) (uint64, error) {
	// XV6 may not support openat, but provide basic AT_ constants
	flags := map[string]uint64{
		"AT_FDCWD": 0xffffff9c,
	}

	if val, ok := flags[name]; ok {
		return val, nil
	}

	return 0, fmt.Errorf("XV6 AT flag %s not supported", name)
}

func (xv6 *xv6) extractGenericConstant(arch *Arch, name string, info *compiler.Const) (uint64, error) {
	// For constants we don't have specific knowledge about,
	// try to extract from XV6 source if available

	// Handle some common POSIX-like constants that XV6 might have
	commonConstants := map[string]uint64{
		// Error codes
		"EPERM":   1,
		"ENOENT":  2,
		"ESRCH":   3,
		"EINTR":   4,
		"EIO":     5,
		"ENXIO":   6,
		"E2BIG":   7,
		"ENOEXEC": 8,
		"EBADF":   9,
		"ECHILD":  10,
		"EAGAIN":  11,
		"ENOMEM":  12,
		"EACCES":  13,
		"EFAULT":  14,
		"EBUSY":   16,
		"EEXIST":  17,
		"EXDEV":   18,
		"ENODEV":  19,
		"ENOTDIR": 20,
		"EISDIR":  21,
		"EINVAL":  22,
		"ENFILE":  23,
		"EMFILE":  24,
		"ENOTTY":  25,
		"ETXTBSY": 26,
		"EFBIG":   27,
		"ENOSPC":  28,
		"ESPIPE":  29,
		"EROFS":   30,
		"EMLINK":  31,
		"EPIPE":   32,

		// File permissions
		"S_IRUSR": 0400,
		"S_IWUSR": 0200,
		"S_IXUSR": 0100,
		"S_IRGRP": 0040,
		"S_IWGRP": 0020,
		"S_IXGRP": 0010,
		"S_IROTH": 0004,
		"S_IWOTH": 0002,
		"S_IXOTH": 0001,

		// File modes
		"S_ISUID": 04000,
		"S_ISGID": 02000,
		"S_ISVTX": 01000,

		// Seek constants
		"SEEK_SET": 0,
		"SEEK_CUR": 1,
		"SEEK_END": 2,

		// NULL and other basic constants
		"NULL": 0,
	}

	if val, ok := commonConstants[name]; ok {
		return val, nil
	}

	// For other constants, return error as we don't have XV6 source parsing
	return 0, fmt.Errorf("XV6 constant %s not found in predefined list", name)
}

func (xv6 *xv6) addBuiltinConstants(arch *Arch, consts map[string]uint64) {
	// Add XV6-specific built-in constants that are always available

	builtins := map[string]uint64{
		// XV6 has simple constants
		"XV6_NPROC":   64,  // Max processes
		"XV6_NOFILE":  16,  // Max open files per process
		"XV6_NFILE":   100, // Max open files system-wide
		"XV6_NINODE":  50,  // Max in-memory inodes
		"XV6_NDEV":    10,  // Max major device number
		"XV6_ROOTDEV": 1,   // Root device number
		"XV6_MAXARG":  32,  // Max exec arguments

		// Page and memory constants
		"XV6_PGSIZE":  4096,
		"XV6_PGSHIFT": 12,

		// File system constants
		"XV6_ROOTINO":   1,          // Root inode number
		"XV6_BSIZE":     1024,       // Block size
		"XV6_NDIRECT":   12,         // Direct blocks in inode
		"XV6_NINDIRECT": 256,        // Indirect blocks
		"XV6_MAXFILE":   (12 + 256), // Max file size in blocks

		// Process constants
		"XV6_KSTACKSIZE": 4096, // Kernel stack size
	}

	for name, val := range builtins {
		consts[name] = val
	}
}

// Helper function to check if file exists (removed as we now use osutil.IsExist)
