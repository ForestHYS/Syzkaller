// Copyright 2024 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package vminfo

import (
	"github.com/google/syzkaller/pkg/flatrpc"
	"github.com/google/syzkaller/prog"
)

type xv6 struct {
	nopChecker
}

// XV6 is a simple teaching OS with limited syscalls
// Most syscalls should be supported by default since it's a minimal system
func (xv6) syscallCheck(ctx *checkContext, call *prog.Syscall) string {
	switch call.CallName {
	// XV6 basic file operations
	case "open", "openat":
		return supportedOpenat(ctx, call)

	// XV6 process operations - usually supported
	case "fork", "exec", "exit", "wait", "getpid":
		return ""

	// XV6 file system operations - basic support
	case "read", "write", "close", "dup", "pipe":
		return ""

	// XV6 memory operations - basic support
	case "sbrk":
		return ""

	// XV6 has basic sleep support
	case "sleep":
		return ""

	// XV6 may not support advanced syscalls
	case "socket", "bind", "listen", "accept", "connect":
		return "XV6 does not support networking syscalls"

	case "mount", "umount":
		return "XV6 does not support mount/umount"

	case "chroot", "chdir":
		// XV6 might have basic directory support
		return ""

	case "kill":
		// XV6 might have basic signal support
		return ""

	// Advanced features not typically in XV6
	case "mmap", "munmap", "mprotect":
		return "XV6 may not support advanced memory management"

	case "clone", "vfork":
		return "XV6 uses simple fork model"

	case "ioctl":
		return "XV6 has limited ioctl support"

	case "fcntl":
		return "XV6 has limited fcntl support"

	// Filesystem features
	case "stat", "fstat", "lstat":
		return "" // XV6 should support basic stat

	case "mkdir", "rmdir", "unlink", "link":
		return "" // Basic filesystem operations

	// XV6 typically doesn't support:
	case "epoll_create", "epoll_ctl", "epoll_wait":
		return "XV6 does not support epoll"

	case "select", "poll":
		return "XV6 does not support select/poll"

	case "sendfile", "splice":
		return "XV6 does not support sendfile/splice"

	case "signalfd", "eventfd", "timerfd_create":
		return "XV6 does not support fd-based event mechanisms"

	case "prctl", "ptrace":
		return "XV6 does not support process control/debugging"

	case "setuid", "setgid", "setreuid", "setregid":
		return "XV6 has simple user model"

	case "capget", "capset":
		return "XV6 does not support capabilities"

	case "acct", "quotactl":
		return "XV6 does not support accounting/quotas"

	case "syslog", "klogctl":
		return "XV6 does not support syslog"

	case "reboot", "sync":
		return "XV6 may not support reboot/sync"

	// Pseudo-syscalls that might not work
	case "syz_open_dev":
		return "XV6 has limited device support"

	case "syz_mount_image":
		return "XV6 does not support complex filesystem mounting"

	case "syz_read_part_table":
		return "XV6 does not support partition tables"

	default:
		// For unknown syscalls, assume they might not be supported
		// XV6 is a minimal OS, so be conservative
		if isXV6PseudoSyscall(call.CallName) {
			return "XV6 does not support complex pseudo-syscalls"
		}

		// Basic syscalls that should work in most Unix-like systems
		if isBasicUnixSyscall(call.CallName) {
			return ""
		}

		// Everything else is likely unsupported
		return "XV6 may not support this advanced syscall"
	}
}

// Check if this is a pseudo-syscall that XV6 definitely doesn't support
func isXV6PseudoSyscall(name string) bool {
	pseudoSyscalls := []string{
		"syz_", // Most syz_ pseudo-syscalls
	}

	for _, prefix := range pseudoSyscalls {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// Check if this is a basic Unix syscall that XV6 should support
func isBasicUnixSyscall(name string) bool {
	basicSyscalls := map[string]bool{
		// Process management
		"fork":    true,
		"exec":    true,
		"exit":    true,
		"wait":    true,
		"getpid":  true,
		"getppid": true,

		// File operations
		"open":  true,
		"read":  true,
		"write": true,
		"close": true,
		"lseek": true,
		"dup":   true,
		"dup2":  true,

		// Directory operations
		"chdir":  true,
		"mkdir":  true,
		"rmdir":  true,
		"unlink": true,
		"link":   true,

		// File metadata
		"stat":  true,
		"fstat": true,

		// Pipes
		"pipe": true,

		// Memory
		"sbrk": true,

		// Basic IPC
		"kill": true,

		// Time
		"sleep": true,
	}

	return basicSyscalls[name]
}

// XV6-specific machine info extraction (minimal)
func (xv6) machineInfos() []machineInfoFunc {
	// XV6 doesn't have complex machine info like Linux
	// Return minimal information
	return []machineInfoFunc{}
}

// XV6 doesn't have kernel modules
func (xv6) parseModules(files filesystem) ([]*KernelModule, error) {
	// XV6 is a monolithic kernel, no loadable modules
	return nil, nil
}

// XV6 has minimal features
func (xv6) checkFeature(feature flatrpc.Feature, files filesystem) string {
	switch feature {
	case flatrpc.FeatureCoverage:
		// XV6 doesn't have kcov coverage
		return "XV6 does not support kernel coverage"

	case flatrpc.FeatureComparisons:
		// XV6 doesn't have comparison coverage
		return "XV6 does not support comparison coverage"

	case flatrpc.FeatureExtraCoverage:
		// XV6 doesn't have extra coverage
		return "XV6 does not support extra coverage"

	case flatrpc.FeatureSandboxSetuid:
		// XV6 has simple user model
		return "XV6 has limited setuid support"

	case flatrpc.FeatureSandboxNamespace:
		// XV6 doesn't have namespaces
		return "XV6 does not support namespaces"

	case flatrpc.FeatureSandboxAndroid:
		// XV6 is not Android
		return "XV6 is not Android"

	case flatrpc.FeatureFault:
		// XV6 may not support fault injection
		return "XV6 may not support fault injection"

	case flatrpc.FeatureLeak:
		// XV6 may not support leak detection
		return "XV6 may not support leak detection"

	case flatrpc.FeatureNetInjection:
		// XV6 doesn't have networking
		return "XV6 does not support network injection"

	case flatrpc.FeatureNetDevices:
		// XV6 doesn't have network devices
		return "XV6 does not support network devices"

	case flatrpc.FeatureKCSAN:
		// XV6 doesn't have KCSAN
		return "XV6 does not support KCSAN"

	case flatrpc.FeatureDevlinkPCI:
		// XV6 doesn't have devlink
		return "XV6 does not support devlink"

	case flatrpc.FeatureUSBEmulation:
		// XV6 doesn't have USB emulation
		return "XV6 does not support USB emulation"

	case flatrpc.FeatureVhciInjection:
		// XV6 doesn't have VHCI
		return "XV6 does not support VHCI injection"

	case flatrpc.FeatureWifiEmulation:
		// XV6 doesn't have WiFi
		return "XV6 does not support WiFi emulation"

	default:
		// Unknown features are not supported
		return "XV6 does not support this feature"
	}
}
