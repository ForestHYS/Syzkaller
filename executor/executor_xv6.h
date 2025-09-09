// Copyright 2024 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

// XV6 is a simple Unix-like teaching operating system.
// This file contains xv6-specific implementation for syzkaller executor.

// XV6 system headers (when compiled for XV6)
#ifdef XV6_BUILD
#include "kernel/types.h"
#include "kernel/stat.h"
#include "kernel/syscall.h"
#include "user/user.h"
#else
// Fallback headers for development/testing
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <sys/types.h>
#endif

// XV6 specific constants
#define XV6_PAGE_SIZE 4096
#define XV6_MAX_ARGS 6  // XV6 syscalls typically have fewer args

// XV6 system call numbers (matching kernel/syscall.h)
#define XV6_SYS_fork    1
#define XV6_SYS_exit    2
#define XV6_SYS_wait    3
#define XV6_SYS_pipe    4
#define XV6_SYS_read    5
#define XV6_SYS_kill    6
#define XV6_SYS_exec    7
#define XV6_SYS_fstat   8
#define XV6_SYS_chdir   9
#define XV6_SYS_dup     10
#define XV6_SYS_getpid  11
#define XV6_SYS_sbrk    12
#define XV6_SYS_sleep   13
#define XV6_SYS_uptime  14
#define XV6_SYS_open    15
#define XV6_SYS_write   16
#define XV6_SYS_mknod   17
#define XV6_SYS_unlink  18
#define XV6_SYS_link    19
#define XV6_SYS_mkdir   20
#define XV6_SYS_close   21

// XV6 file flags
#define XV6_O_RDONLY  0x000
#define XV6_O_WRONLY  0x001
#define XV6_O_RDWR    0x002
#define XV6_O_CREATE  0x200
#define XV6_O_TRUNC   0x400

// XV6 doesn't have signal handling like Linux
static void os_init(int argc, char** argv, void* data, size_t data_size)
{
	// XV6 initialization is much simpler than Linux
	// Just ensure we have enough memory allocated
	
#ifdef XV6_BUILD
	// Use XV6's sbrk to allocate memory for data segment
	if (sbrk(data_size) == (void*)-1) {
		printf("syz-executor: failed to allocate memory\n");
		exit(1);
	}
#else
	// For development builds, just mark data as used
	(void)data;
	(void)data_size;
#endif
	
	// XV6 doesn't have complex process setup like Linux
	// No namespaces, no cgroups, no complex signal handling
}

// XV6 system call execution - the core function
static intptr_t execute_syscall(const call_t* c, intptr_t a[XV6_MAX_ARGS])
{
	// Direct system call mapping for XV6
	// We map syzkaller syscall numbers to XV6 syscall numbers
	
	switch (c->sys_nr) {
	case XV6_SYS_fork:
		return fork();
		
	case XV6_SYS_exit:
		exit((int)a[0]);
		return 0; // Never reached
		
	case XV6_SYS_wait:
		return wait((int*)a[0]);
		
	case XV6_SYS_pipe:
		return pipe((int*)a[0]);
		
	case XV6_SYS_read:
		return read((int)a[0], (void*)a[1], (int)a[2]);
		
	case XV6_SYS_write:
		return write((int)a[0], (void*)a[1], (int)a[2]);
		
	case XV6_SYS_open:
		return open((char*)a[0], (int)a[1]);
		
	case XV6_SYS_close:
		return close((int)a[0]);
		
	case XV6_SYS_kill:
		return kill((int)a[0]);
		
	case XV6_SYS_exec:
		return exec((char*)a[0], (char**)a[1]);
		
	case XV6_SYS_fstat:
		return fstat((int)a[0], (struct stat*)a[1]);
		
	case XV6_SYS_chdir:
		return chdir((char*)a[0]);
		
	case XV6_SYS_dup:
		return dup((int)a[0]);
		
	case XV6_SYS_getpid:
		return getpid();
		
	case XV6_SYS_sbrk:
		return (intptr_t)sbrk((int)a[0]);
		
	case XV6_SYS_sleep:
		return sleep((int)a[0]);
		
	case XV6_SYS_uptime:
		return uptime();
		
	case XV6_SYS_mknod:
		return mknod((char*)a[0], (short)a[1], (short)a[2]);
		
	case XV6_SYS_unlink:
		return unlink((char*)a[0]);
		
	case XV6_SYS_link:
		return link((char*)a[0], (char*)a[1]);
		
	case XV6_SYS_mkdir:
		return mkdir((char*)a[0]);
		
	default:
		// Unsupported system call
		return -1;
	}
}

// XV6 doesn't have kcov, so coverage collection is minimal
static void cover_open(cover_t* cov, bool extra)
{
	// XV6 doesn't have kernel coverage infrastructure
	// All fields set to indicate no coverage
	if (cov) {
		cov->fd = -1;
		cov->mmap_alloc_size = 0;
		cov->data = NULL;
		cov->data_end = NULL;
		cov->data_offset = 0;
		cov->pc_offset = 0;
	}
}

static void cover_enable(cover_t* cov, bool collect_comps, bool extra)
{
	// No-op for XV6 - no coverage support
	(void)cov; (void)collect_comps; (void)extra;
}

static void cover_reset(cover_t* cov)
{
	// No-op for XV6
	(void)cov;
}

static void cover_collect(cover_t* cov)
{
	// No-op for XV6
	(void)cov;
}

static bool cover_check(uint32 pc)
{
	// No coverage checking for XV6
	(void)pc;
	return true;
}

static bool cover_check(uint64 pc)
{
	// No coverage checking for XV6
	(void)pc;
	return true;
}

// Simplified function implementations for XV6
static void setup_control_pipes() { /* XV6 doesn't need complex pipes */ }
static void setup_common() { /* XV6 has no network */ }
static void setup_fault() { /* XV6 has no fault injection */ }
static void setup_leak() { /* XV6 has no leak detection */ }
static void install_segv_handler() { /* XV6 has limited signal support */ }
static void setup_usb() { /* XV6 has no USB */ }
static void setup_sysctl() { /* XV6 has no sysctl */ }
static void setup_binfmt_misc() { /* XV6 has no binfmt_misc */ }
static void setup_net() { /* XV6 has no networking */ }
static void setup_sandbox() { /* XV6 has no sandboxing */ }

// XV6 namespace functions (all no-ops)
static void use_net_namespace() { /* XV6 has no namespaces */ }
static void use_pid_namespace() { /* XV6 has no namespaces */ }
static void use_uts_namespace() { /* XV6 has no namespaces */ }
static void use_ipc_namespace() { /* XV6 has no namespaces */ }
static void use_user_namespace() { /* XV6 has no namespaces */ }
static void use_cgroup_namespace() { /* XV6 has no namespaces */ }
static void use_time_namespace() { /* XV6 has no namespaces */ }
static void use_tmpdir() { /* XV6 has basic filesystem */ }
static void use_sysctl() { /* XV6 has no sysctl */ }
static void use_cgroups() { /* XV6 has no cgroups */ }
static void use_tmpfile() { /* XV6 has basic files */ }
static void drop_caps() { /* XV6 has no capabilities */ }

// Basic file operations for XV6
static void write_file(const char* file, const char* what, ...)
{
	// Simple file write for XV6
	int fd = open(file, XV6_O_WRONLY | XV6_O_CREATE);
	if (fd >= 0) {
		write(fd, what, strlen(what));
		close(fd);
	}
}

// Network functions (not supported)
static int read_tun_fd(int tunfd) { (void)tunfd; return -1; }
static int wait_for_loop(int pid) { (void)pid; return 0; }

// Memory allocation for XV6 (simplified)
static long syz_mmap(volatile long a0, volatile long a1)
{
	// XV6 doesn't have mmap like Linux
	// Use sbrk for memory allocation
	(void)a0; // ignore requested address
	return (long)sbrk((int)a1);
}

// End of XV6 executor implementation
