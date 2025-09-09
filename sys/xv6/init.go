// Copyright 2024 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package xv6

import (
	"github.com/google/syzkaller/prog"
	"github.com/google/syzkaller/sys/targets"
)

func InitTarget(target *prog.Target) {
	switch target.Arch {
	case targets.RiscV64:
		// XV6 primarily runs on RISC-V 64-bit
		target.PageSize = 4096
		target.DataOffset = 0x1000000
		target.NumPages = 512 // XV6 has limited memory
	case targets.I386:
		// XV6 also has x86 support (legacy)
		target.PageSize = 4096
		target.DataOffset = 0x800000
		target.NumPages = 512
	default:
		// Default to RISC-V settings
		target.PageSize = 4096
		target.DataOffset = 0x1000000
		target.NumPages = 512
	}

	// XV6 uses simple data mapping similar to POSIX
	target.MakeDataMmap = targets.MakePosixMmap(target, false, false)
}
