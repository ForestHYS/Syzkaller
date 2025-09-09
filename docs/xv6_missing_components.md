# XV6 Syzkaller支持 - 缺失的关键组件

## 当前状态
目前的实现只是一个**框架原型**，不能直接运行。主要问题：

### 1. Executor架构问题
**问题**: `executor_xv6.h`试图在Linux主机上调用XV6系统调用
**解决方案需要**:
- 交叉编译executor到XV6
- 或者使用代理/桥接机制
- 或者在QEMU内运行executor

### 2. 缺失的核心文件

#### A. XV6系统调用桥接
```c
// 需要实现：executor/xv6_syscall_bridge.c
// 将syzkaller系统调用映射到XV6系统调用

// 例如：Linux的open()映射到XV6的open()
static int xv6_open(const char* path, int flags) {
    // 需要通过某种机制调用XV6内核
    // 可能通过QEMU monitor或其他方式
}
```

#### B. XV6头文件支持
```c
// 需要：include/xv6/
// - syscall.h (XV6系统调用定义)  
// - types.h (XV6类型定义)
// - stat.h (XV6文件状态结构)
// - fcntl.h (XV6文件控制)
```

#### C. 构建系统集成
```makefile
# 需要修改Makefile支持XV6目标
xv6-executor:
	riscv64-linux-gnu-gcc -o syz-executor executor.cc \
		-DGOOS_xv6 -DGOARCH_riscv64 \
		-I./xv6-headers \
		-static
```

### 3. 运行时通信机制

#### A. VM通信协议
当前缺少syzkaller manager与XV6 VM的通信方式：
- 如何将测试程序传输到XV6
- 如何从XV6获取执行结果
- 如何检测XV6崩溃

#### B. 可能的解决方案
1. **串口通信**: 通过串口传输命令和结果
2. **9P文件系统**: 使用QEMU的9P共享文件系统
3. **网络**: 如果XV6支持网络（通常不支持）
4. **QEMU Monitor**: 通过QEMU monitor接口

### 4. XV6特定的限制处理

#### A. 内存限制
```go
// pkg/vminfo/xv6.go需要更准确的限制
func (xv6) checkMemoryLimits() {
    // XV6通常只有128MB内存
    // 需要限制测试程序大小
    // 需要限制并发执行数量
}
```

#### B. 文件系统限制
```go
// XV6文件系统很简单
func (xv6) checkFilesystemLimits() {
    // 文件名限制: 14字符
    // 文件数量限制
    // 目录深度限制
}
```

## 实现优先级

### 第一阶段：基础通信
1. **实现串口通信机制**
   ```go
   // pkg/vmimpl/qemu_xv6.go
   func (inst *instance) communicateViaSerial() error {
       // 通过QEMU串口与XV6通信
   }
   ```

2. **简化的executor**
   ```c
   // executor/simple_xv6_executor.c
   // 直接在XV6上编译运行的简化版本
   ```

### 第二阶段：系统调用映射
```c
// executor/xv6_syscalls.h
#define XV6_SYS_fork    1
#define XV6_SYS_exit    2
#define XV6_SYS_wait    3
// ... 等等

// 映射函数
static long xv6_syscall(int nr, long a0, long a1, long a2, long a3, long a4, long a5) {
    // 实际的XV6系统调用
    // 这需要在XV6环境中编译
    asm volatile(
        "li a7, %1\n"
        "mv a0, %2\n"
        "mv a1, %3\n"
        "mv a2, %4\n"
        "mv a3, %5\n"
        "mv a4, %6\n"
        "mv a5, %7\n"
        "ecall\n"
        "mv %0, a0\n"
        : "=r"(ret)
        : "r"(nr), "r"(a0), "r"(a1), "r"(a2), "r"(a3), "r"(a4), "r"(a5)
        : "a0", "a1", "a2", "a3", "a4", "a5", "a7"
    );
    return ret;
}
```

### 第三阶段：完整集成
1. **崩溃检测**
2. **覆盖率收集**（如果可能）
3. **自动化测试**

## 当前executor_xv6.h的问题

### 为什么有这么多TODO？

1. **概念性问题**: 试图在Linux上调用XV6函数
   ```c
   // 这些在Linux上不存在或不兼容：
   prctl_set_vma = prctl_set_vma_anon;  // XV6没有prctl
   syscall(SYS_sigaction, SIGSEGV, &sa, nullptr);  // XV6的信号不同
   ```

2. **未定义的符号**: 
   ```c
   // 这些在当前代码中未定义：
   segv_handler  // 需要实现
   call_t        // 需要定义
   cover_t       // 需要定义
   kMaxArgs      // 需要定义
   ```

3. **架构不匹配**:
   ```c
   // 试图使用Linux的mmap调用XV6
   mmap((void*)a0, a1, PROT_READ | PROT_WRITE, MAP_ANON | MAP_PRIVATE, -1, 0);
   ```

## 正确的实现路径

### 方案A：在XV6内运行executor（推荐）
1. 将executor交叉编译到RISC-V
2. 在XV6启动时加载executor
3. 通过串口与外部syzkaller通信

### 方案B：代理模式
1. 在Linux上运行代理executor
2. 通过QEMU monitor或串口与XV6通信
3. 将系统调用转发到XV6

### 方案C：混合模式
1. 轻量级的XV6内执行器
2. Linux上的控制器
3. 通过共享文件系统交换数据

## 下一步行动计划

如果要真正实现可运行的XV6 fuzzing，建议：

1. **先实现简单的XV6程序执行机制**
2. **建立基础通信协议**
3. **逐步添加系统调用支持**
4. **最后完善错误检测和报告**

现在的代码更像是一个"占位符"框架，需要大量具体实现才能工作。
