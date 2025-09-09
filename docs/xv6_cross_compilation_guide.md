# XV6交叉编译详细指南

## 什么是交叉编译？

### 📚 基本概念

**交叉编译**是指在一个平台上编译出能在另一个平台上运行的程序。

```
主机平台 (Host)     目标平台 (Target)
x86_64 Linux   →   RISC-V XV6
ARM64 Mac      →   RISC-V XV6  
Windows x64    →   RISC-V XV6
```

### 🔄 对比普通编译

```bash
# 普通编译（本地编译）
gcc hello.c -o hello           # x86程序 在 x86机器上运行

# 交叉编译  
riscv64-unknown-elf-gcc hello.c -o hello  # RISC-V程序 在 x86机器上编译
```

## 为什么XV6需要交叉编译？

### 🎯 架构差异

| 组件 | 架构 | 说明 |
|------|------|------|
| 开发机器 | x86_64/ARM64 | 我们写代码的地方 |
| XV6目标 | RISC-V | XV6运行的架构 |
| Syzkaller Manager | x86_64/ARM64 | 运行在开发机器上 |
| Syzkaller Executor | RISC-V | 运行在XV6虚拟机中 |

### 💡 关键理解

```
┌─────────────────┐    ┌─────────────────┐
│   开发机器       │    │   QEMU VM       │
│   x86_64        │    │   RISC-V        │
│                │    │                │
│ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │syz-manager  │ │◄──►│ │syz-executor │ │
│ │(Go binary)  │ │    │ │(C++ binary)│ │
│ │x86_64       │ │    │ │RISC-V      │ │
│ └─────────────┘ │    │ └─────────────┘ │
│                │    │                │
│                │    │   XV6 Kernel    │
└─────────────────┘    └─────────────────┘
```

## 安装RISC-V工具链

### Ubuntu/Debian
```bash
# 方法1：通过包管理器（推荐）
sudo apt update
sudo apt install gcc-riscv64-linux-gnu g++-riscv64-linux-gnu

# 检查安装
riscv64-linux-gnu-gcc --version

# 方法2：安装裸机工具链（如果方法1不行）
sudo apt install gcc-riscv64-unknown-elf g++-riscv64-unknown-elf
```

### macOS
```bash
# 使用Homebrew
brew tap riscv/riscv
brew install riscv-tools

# 或者安装特定的工具链
brew install riscv64-elf-gcc
```

### 从源码编译（所有平台）
```bash
# 下载并编译RISC-V工具链
git clone https://github.com/riscv/riscv-gnu-toolchain
cd riscv-gnu-toolchain
git submodule update --init --recursive

# 配置安装路径
./configure --prefix=/opt/riscv --with-arch=rv64gc --with-abi=lp64d
make -j$(nproc)

# 添加到PATH
export PATH="/opt/riscv/bin:$PATH"
```

## Makefile修改详解

### 🔧 为什么需要修改Makefile？

syzkaller的原始Makefile假设：
1. 使用系统默认编译器（通常是gcc/g++）
2. 编译目标是主机架构
3. 使用标准C/C++库

但XV6需要：
1. RISC-V交叉编译器
2. 编译目标是RISC-V架构  
3. 使用XV6的用户库

### 📝 具体修改

#### 1. 工具链设置
```makefile
# 原来（syzkaller默认）
CXX = g++
CC = gcc

# 修改后（XV6）
CROSS_COMPILE = riscv64-unknown-elf-
CXX = $(CROSS_COMPILE)g++
CC = $(CROSS_COMPILE)gcc
```

#### 2. 编译标志
```makefile
# 原来（Linux标准）
CFLAGS = -O2 -Wall -std=c++17

# 修改后（XV6裸机）
CFLAGS = -O2 -Wall -std=c++17
CFLAGS += -mcmodel=medium    # RISC-V内存模型
CFLAGS += -nostdinc          # 不使用标准头文件
CFLAGS += -nostdlib          # 不使用标准库
CFLAGS += -static            # 静态链接
CFLAGS += -fno-stack-protector  # 关闭栈保护
```

#### 3. 包含路径
```makefile
# 添加XV6头文件路径
CFLAGS += -I$(XV6_PATH)/kernel  # XV6内核头文件
CFLAGS += -I$(XV6_PATH)/user    # XV6用户库头文件
```

#### 4. 链接设置
```makefile
# 使用XV6的链接脚本和用户库
LDFLAGS = -T $(XV6_PATH)/user/user.ld  # XV6链接脚本
XV6_LIBS = $(XV6_PATH)/user/ulib.o $(XV6_PATH)/user/usys.o  # XV6用户库
```

## 使用新的构建系统

### 🚀 快速开始

```bash
# 1. 确保有RISC-V工具链
riscv64-unknown-elf-gcc --version

# 2. 获取XV6源码
git clone https://github.com/mit-pdos/xv6-riscv.git

# 3. 编译XV6（准备用户库）
cd xv6-riscv
make
cd ..

# 4. 编译XV6 executor
cd executor
make -f Makefile.xv6 XV6_PATH=../xv6-riscv

# 5. 检查结果
ls -la syz-executor-xv6
file syz-executor-xv6  # 应该显示RISC-V binary
```

### 🔍 故障排除

#### 问题1：找不到交叉编译器
```bash
# 错误信息
make: riscv64-unknown-elf-gcc: Command not found

# 解决方案
# 确保工具链已安装并在PATH中
which riscv64-unknown-elf-gcc
export PATH="/opt/riscv/bin:$PATH"
```

#### 问题2：找不到XV6头文件
```bash
# 错误信息
fatal error: 'kernel/types.h' file not found

# 解决方案
# 确保XV6路径正确且已编译
make -f Makefile.xv6 XV6_PATH=/correct/path/to/xv6-riscv
```

#### 问题3：链接错误
```bash
# 错误信息
undefined reference to `printf`

# 解决方案
# 确保使用XV6的用户库
# 检查XV6用户库是否存在
ls -la ../xv6-riscv/user/ulib.o
ls -la ../xv6-riscv/user/usys.o
```

## 验证交叉编译结果

### 🧪 检查二进制文件

```bash
# 检查文件类型
file syz-executor-xv6
# 输出应该是：ELF 64-bit LSB executable, UCB RISC-V, version 1 (SYSV), statically linked, not stripped

# 检查架构
readelf -h syz-executor-xv6 | grep Machine
# 输出应该是：Machine: RISC-V

# 检查大小
ls -lh syz-executor-xv6
# XV6 executor应该比较小，通常几百KB
```

### 🔬 反汇编检查

```bash
# 生成反汇编
make -f Makefile.xv6 disasm

# 查看入口点
head -20 syz-executor-xv6.asm
# 应该看到RISC-V指令，如：addi, auipc等
```

## 集成到主构建系统

### 📁 修改主Makefile

在主Makefile中添加XV6支持：

```makefile
# 在executor规则中添加XV6特殊处理
executor: descriptions
ifeq ($(TARGETOS),xv6)
	$(MAKE) -C executor -f Makefile.xv6 install XV6_PATH=$(XV6_PATH)
else
	# 原有的executor编译逻辑
	...
endif
```

### 🎯 使用方法

```bash
# 设置环境变量
export XV6_PATH=/path/to/xv6-riscv

# 编译XV6 executor
make executor TARGETOS=xv6 TARGETARCH=riscv64

# 或者直接使用XV6 Makefile
cd executor
make -f Makefile.xv6
```

## 下一步：运行时集成

交叉编译只是第一步，接下来需要：

1. **通信机制** - syzkaller manager与XV6中的executor通信
2. **程序传输** - 如何将测试程序传入XV6
3. **结果收集** - 如何从XV6获取执行结果
4. **崩溃检测** - 如何检测XV6内核崩溃

这些将在后续的集成阶段完成。

## 📚 参考资料

- [RISC-V工具链文档](https://github.com/riscv/riscv-gnu-toolchain)
- [XV6源码](https://github.com/mit-pdos/xv6-riscv)
- [交叉编译原理](https://en.wikipedia.org/wiki/Cross_compiler)
- [GNU Make手册](https://www.gnu.org/software/make/manual/)

交叉编译是实现XV6 fuzzing的关键第一步，掌握这个基础后，后续的集成工作会变得清晰很多！
