# XV6 Syzkaller支持 - 当前状态

## ⚠️ 重要声明
**当前的XV6支持是一个原型框架，不能直接运行fuzz测试。**

## 当前完成的工作

### ✅ 已实现的框架组件
1. **基础架构支持** - 所有必要的Go代码框架已完成
2. **构建系统集成** - pkg/build支持XV6内核编译
3. **目标系统定义** - sys/targets中已定义XV6
4. **QEMU虚拟化配置** - vm/qemu支持XV6启动
5. **错误报告框架** - pkg/report可以解析XV6错误
6. **系统调用描述** - sys/xv6/sys.txt定义了基础系统调用

### 🏗️ 框架性质的executor
`executor/executor_xv6.h`目前只是一个**框架占位符**，包含：
- 函数签名和接口定义
- 空的实现存根
- 大量的TODO注释

## ❌ 缺失的关键实现

### 1. 根本性架构问题
```
当前executor试图：
Host Linux System → XV6 syscalls
这在逻辑上不可能工作
```

**需要的解决方案：**
- 在XV6内部运行executor，或
- 实现系统调用代理机制，或  
- 通过QEMU接口桥接系统调用

### 2. 缺失的核心组件

#### A. 真实的系统调用实现
```c
// 当前（占位符）：
return syscall(c->sys_nr, a[0], a[1], a[2], a[3], a[4], a[5]);

// 需要的实现：
// 选项1: XV6内部执行
asm volatile("ecall" : "=r"(ret) : "r"(syscall_nr), "r"(a0), ...);

// 选项2: 通过QEMU bridge
ret = qemu_xv6_syscall(c->sys_nr, a[0], a[1], ...);
```

#### B. 通信机制
目前没有syzkaller与XV6虚拟机的通信方式：
- 如何传输测试程序到XV6
- 如何接收执行结果
- 如何检测崩溃

#### C. 内存管理
```c
// 当前（不正确）：
mmap((void*)a0, a1, PROT_READ | PROT_WRITE, MAP_ANON | MAP_PRIVATE, -1, 0);

// XV6实际上：
// 可能根本不支持mmap，或者有完全不同的接口
```

### 3. XV6特定限制未处理
- 内存限制（通常只有128MB）
- 文件系统限制（简单的文件系统）
- 进程数限制
- 系统调用参数限制

## 🎯 要让它真正工作需要什么

### 最小可行产品 (MVP) 需要：

1. **选择执行模式**：
   ```
   选项A: XV6内executor
   ├── 将executor交叉编译到RISC-V
   ├── 在XV6启动时自动加载
   └── 通过串口与外部syzkaller通信

   选项B: 代理模式  
   ├── Linux主机上运行代理executor
   ├── 通过QEMU monitor接口
   └── 转发系统调用到XV6
   ```

2. **实现基础通信**：
   ```c
   // 串口通信协议
   write_serial("EXEC:open:/tmp/test:O_RDWR\n");
   response = read_serial(); // "RESULT:3\n" (fd=3)
   ```

3. **系统调用映射**：
   ```c
   // 将syzkaller调用映射到XV6
   int xv6_open(const char* path, int flags) {
       // 实际调用XV6的open系统调用
       // 这需要在XV6环境中运行，或通过代理
   }
   ```

### 估计工作量：

- **MVP实现**: 2-4周（选择代理模式）
- **完整实现**: 2-3个月
- **生产就绪**: 6个月+

## 🚀 如果你想尝试运行

### 当前状态测试：
```bash
# 1. 编译syzkaller（会成功）
make manager fuzzer

# 2. 尝试运行（会失败）
./bin/syz-manager -config=xv6-config.json
# 错误：executor编译失败或运行时崩溃
```

### 快速原型建议：
1. **从简单的XV6程序开始**
2. **手动测试单个系统调用**
3. **建立串口通信机制**
4. **逐步集成到syzkaller**

## 📚 相关资源

- [XV6源码](https://github.com/mit-pdos/xv6-riscv)
- [Syzkaller文档](https://github.com/google/syzkaller/tree/master/docs)
- [QEMU RISC-V文档](https://www.qemu.org/docs/master/system/target-riscv.html)

## 💡 替代方案

如果你主要想学习/研究fuzzing，建议：
1. **先使用现有的Linux支持**熟悉syzkaller
2. **然后考虑简化的操作系统**（如minimal Linux）
3. **最后再考虑XV6这样的特殊系统**

XV6支持虽然有教育价值，但实现复杂度很高，不适合作为syzkaller的入门项目。
