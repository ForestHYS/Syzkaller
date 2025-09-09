# XV6 Syzkaller 完整实现方案

## 🎯 目标
实现一个可工作的XV6 fuzzing系统，能够在QEMU中运行XV6并进行系统调用模糊测试。

## 📋 实现状态

### ✅ 已完成（基础框架）
1. **Go语言框架** - 所有syzkaller组件已实现
2. **Executor头文件** - `executor/executor_xv6.h` 现在有正确的系统调用映射
3. **系统调用描述** - `sys/xv6/sys.txt` 定义了XV6系统调用
4. **构建支持** - pkg/build可以处理XV6编译
5. **QEMU配置** - vm/qemu支持XV6启动

### 🔧 需要完成的关键组件

## 第一阶段：基础实现（1-2周）

### 1. 修复Executor编译

#### A. 创建XV6编译配置
```makefile
# executor/Makefile.xv6
XV6_PATH ?= ../xv6-riscv
CC = riscv64-unknown-elf-gcc
CFLAGS = -Wall -Werror -O -fno-omit-frame-pointer -ggdb
CFLAGS += -DGOOS_xv6 -DGOARCH_riscv64 -DXV6_BUILD
CFLAGS += -mcmodel=medium -nostdinc -fno-stack-protector
CFLAGS += -fno-common -fno-builtin
CFLAGS += -I$(XV6_PATH)/kernel -I$(XV6_PATH)/user

executor-xv6: executor.cc executor_xv6.h
	$(CC) $(CFLAGS) -T $(XV6_PATH)/user/user.ld \
		-o syz-executor-xv6 executor.cc \
		$(XV6_PATH)/user/ulib.c $(XV6_PATH)/user/usys.pl
```

#### B. 适配XV6用户空间
```c
// executor/xv6_user.h - XV6用户空间适配
#ifndef XV6_USER_H
#define XV6_USER_H

// XV6用户空间系统调用包装
int fork(void);
int exit(int) __attribute__((noreturn));
int wait(int*);
int pipe(int*);
int write(int, const void*, int);
int read(int, void*, int);
int close(int);
int kill(int);
int exec(const char*, char**);
int open(const char*, int);
int mknod(const char*, short, short);
int unlink(const char*);
int fstat(int fd, struct stat*);
int link(const char*, const char*);
int mkdir(const char*);
int chdir(const char*);
int dup(int);
int getpid(void);
char* sbrk(int);
int sleep(int);
int uptime(void);

#endif
```

### 2. 建立通信机制

#### A. 串口通信协议
```c
// executor/xv6_comm.c - XV6通信模块
#define COMM_PORT 1  // 串口设备

void comm_init() {
    // 初始化串口通信
}

int comm_send_result(int syscall_nr, int result, int error) {
    char buf[128];
    snprintf(buf, sizeof(buf), "RESULT:%d:%d:%d\n", syscall_nr, result, error);
    return write(COMM_PORT, buf, strlen(buf));
}

int comm_receive_command(char* buf, int size) {
    return read(COMM_PORT, buf, size);
}
```

#### B. 修改VM配置支持串口
```go
// vm/qemu/qemu.go 中XV6配置
"xv6/riscv64": {
    Qemu:     "qemu-system-riscv64",
    QemuArgs: "-machine virt -cpu rv64 -nographic",
    NetDev:   "", // XV6不需要网络
    RngDev:   "",
    UseNewQemuImageOptions: true,
    CmdLine: []string{
        "console=ttyS0",
    },
    // 添加串口重定向
    ExtraArgs: []string{
        "-chardev", "stdio,id=char0,mux=on,signal=off",
        "-serial", "chardev:char0",
    },
},
```

### 3. 最小可行测试

#### A. 简单的测试程序
```c
// test/xv6_simple_test.c
int main() {
    printf("XV6 Executor Test Starting\n");
    
    // 测试基础系统调用
    int pid = fork();
    if (pid == 0) {
        printf("Child process: %d\n", getpid());
        exit(0);
    } else {
        int status;
        wait(&status);
        printf("Parent process: child exited\n");
    }
    
    // 测试文件操作
    int fd = open("test.txt", O_CREATE | O_WRONLY);
    if (fd >= 0) {
        write(fd, "Hello XV6\n", 10);
        close(fd);
        printf("File test completed\n");
    }
    
    printf("XV6 Executor Test Completed\n");
    return 0;
}
```

## 第二阶段：集成测试（2-3周）

### 1. Syzkaller集成

#### A. 修改manager支持XV6
```go
// pkg/manager/manager.go 中添加XV6特殊处理
func (mgr *Manager) createInstance() {
    if mgr.cfg.TargetOS == "xv6" {
        // XV6特殊的实例创建逻辑
        // 处理串口通信
        // 处理有限的系统调用集
    }
}
```

#### B. XV6特定的测试生成
```go
// pkg/fuzzer/xv6_fuzzer.go
func (fuzzer *XV6Fuzzer) generateProgram() *prog.Prog {
    // 生成适合XV6的测试程序
    // 考虑XV6的限制：
    // - 最多16个文件描述符
    // - 简单的文件系统
    // - 有限的进程数
}
```

### 2. 错误检测改进

#### A. XV6特定的崩溃检测
```go
// pkg/report/xv6.go 改进
func (ctx *xv6) detectPanic(output []byte) *Report {
    // 检测XV6特有的panic模式
    // "panic: " - 内核panic
    // "trap " - 陷阱错误
    // "cpu halt" - CPU停止
}
```

## 第三阶段：优化和完善（3-4周）

### 1. 性能优化

#### A. 减少测试程序复杂度
```go
// 限制XV6测试的复杂度
const (
    XV6_MAX_PROCS = 4        // 最多4个进程
    XV6_MAX_FILES = 8        // 最多8个文件
    XV6_MAX_SYSCALLS = 20    // 每个程序最多20个系统调用
    XV6_MEMORY_LIMIT = 1024  // 1KB内存限制
)
```

#### B. 快速重启机制
```go
// 由于XV6启动快，可以频繁重启来清理状态
func (pool *XV6Pool) resetVM() {
    // 快速重启XV6实例
    // XV6启动通常只需要1-2秒
}
```

### 2. 覆盖率收集（可选）

#### A. 简单的函数覆盖率
```c
// 在XV6内核中添加简单的覆盖率收集
static uint64 coverage_bitmap[1024];
static int coverage_enabled = 0;

void record_function(uint64 pc) {
    if (coverage_enabled) {
        int idx = (pc >> 4) & 1023;
        coverage_bitmap[idx] = 1;
    }
}
```

## 💻 实际使用示例

### 1. 编译Syzkaller
```bash
# 1. 确保有RISC-V工具链
sudo apt install gcc-riscv64-linux-gnu

# 2. 获取XV6源码
git clone https://github.com/mit-pdos/xv6-riscv.git

# 3. 编译syzkaller（包括XV6支持）
make manager fuzzer
make executor-xv6 XV6_PATH=./xv6-riscv

# 4. 编译XV6
cd xv6-riscv && make && cd ..
```

### 2. 配置文件
```json
{
    "target": "xv6/riscv64",
    "http": "127.0.0.1:56741",
    "workdir": "./workdir",
    "kernel_obj": "./xv6-riscv/kernel",
    "kernel_src": "./xv6-riscv",
    "image": "./xv6-riscv/fs.img",
    "procs": 1,
    "type": "qemu",
    "vm": {
        "count": 1,
        "cpu": 1,
        "mem": 128,
        "kernel": "./xv6-riscv/kernel",
        "qemu": "qemu-system-riscv64",
        "qemu_args": "-machine virt -cpu rv64 -nographic"
    }
}
```

### 3. 运行
```bash
./bin/syz-manager -config=xv6-config.json
```

## 🔍 预期结果

### 成功指标：
1. **启动成功** - XV6在QEMU中正常启动
2. **executor运行** - syz-executor在XV6中成功执行
3. **系统调用执行** - 基础系统调用（open, read, write, fork等）正常工作
4. **错误检测** - 能够检测到XV6内核panic和崩溃
5. **持续运行** - 能够连续运行测试程序而不挂起

### 性能预期：
- **启动时间**: 2-3秒（XV6启动快）
- **测试速度**: 10-50个程序/秒（取决于XV6性能）
- **稳定性**: 连续运行几小时不崩溃

## 🚀 开始实现

要开始实现，建议按以下顺序：

1. **先测试当前框架** - 编译并尝试运行现有代码
2. **修复编译错误** - 解决XV6特定的编译问题
3. **实现基础通信** - 建立syzkaller与XV6的通信
4. **逐步添加功能** - 一个系统调用一个系统调用地测试

**估计总工作量**: 6-8周的专门开发时间，可以产生一个基本可用的XV6 fuzzer。

这是一个有挑战性但完全可行的项目！
