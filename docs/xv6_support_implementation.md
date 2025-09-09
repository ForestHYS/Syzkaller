# 为Syzkaller添加XV6系统支持的实现文档

## 概述

本文档详细说明了为syzkaller添加对在QEMU上运行的RISC-V/xv6系统支持所做的修改。XV6是MIT开发的一个简单的类Unix教学操作系统，主要用于操作系统课程教学。

## 修改的文件和目录

### 1. 执行器支持 (executor/)

#### 新增文件: `executor/executor_xv6.h`
- **目的**: 为XV6系统提供系统调用执行和初始化支持
- **主要功能**:
  - `os_init()`: XV6系统初始化，由于XV6比Linux简单，初始化需求较少
  - `execute_syscall()`: 执行XV6系统调用，使用简化的syscall接口
  - 覆盖率收集函数: XV6没有kcov等高级覆盖率工具，提供基础实现
  - 内存管理: XV6内存管理比Linux简单，提供基础mmap支持
  - 网络和沙箱: XV6缺乏复杂的网络和沙箱功能，提供存根实现

#### 关键特性:
- XV6没有复杂的命名空间、cgroup等Linux特性
- 信号处理和进程管理都比较简单
- 文件系统操作有基本支持但功能有限

### 2. 构建系统支持 (pkg/build/)

#### 修改文件: `pkg/build/build.go`
- **修改内容**: 在构建器映射中添加XV6支持
- **变更**: 添加 `"xv6": xv6{}` 到builders映射

#### 新增文件: `pkg/build/xv6.go`
- **目的**: 实现XV6内核构建逻辑
- **主要功能**:
  - `build()`: 构建XV6内核，使用Makefile系统
  - `buildKernel()`: 编译XV6内核，支持RISC-V工具链
  - `createFileSystem()`: 创建XV6文件系统映像
  - `generateSSHKey()`: 生成SSH密钥（XV6可能不支持SSH，提供基础实现）
  - `clean()`: 清理构建产物

#### 关键特性:
- 支持RISC-V64和i386架构
- 使用简单的Makefile构建系统
- 验证XV6源码目录结构
- 提供工具链检查功能

### 3. 错误报告支持 (pkg/report/)

#### 新增文件: `pkg/report/xv6.go`
- **目的**: 解析和报告XV6系统错误
- **主要功能**:
  - 检测XV6内核panic、断言失败、段错误等
  - 提取崩溃上下文和堆栈跟踪
  - 分类XV6特定的错误类型

#### 支持的错误模式:
- `panic: <message>`: XV6内核panic
- `assertion failed: <condition>`: 断言失败
- `segmentation fault`: 内存访问错误
- `stack overflow`: 栈溢出
- 简单的堆栈跟踪解析

### 4. 系统调用支持检查 (pkg/vminfo/)

#### 新增文件: `pkg/vminfo/xv6.go`
- **目的**: 检查XV6系统调用支持情况
- **主要功能**:
  - `syscallCheck()`: 检查特定系统调用是否被XV6支持
  - 基础Unix系统调用支持检查
  - 高级功能不支持检查（网络、复杂内存管理等）

#### 修改文件: `pkg/vminfo/vminfo.go`
- **修改内容**: 在vminfo构造器中添加XV6支持

#### 支持的系统调用类别:
- **支持**: fork, exec, exit, wait, open, read, write, close等基础系统调用
- **不支持**: socket, epoll, mmap (高级版本), 信号量、共享内存等复杂功能

### 5. 目标系统定义 (sys/)

#### 新增目录和文件: `sys/xv6/`

##### `sys/xv6/init.go`
- **目的**: 初始化XV6目标系统配置
- **主要功能**:
  - 设置XV6内存布局和页面大小
  - 配置RISC-V64和i386架构特定参数
  - 实现XV6特定的系统调用清理和验证
  - 文件名和文件描述符生成逻辑

##### `sys/xv6/sys.txt`
- **目的**: 定义XV6系统调用接口
- **内容**:
  - XV6系统调用号定义 (SYS_fork=1, SYS_exit=2等)
  - 文件操作标志 (O_RDONLY, O_WRONLY等)
  - 基础系统调用描述 (fork, exec, open, read等)
  - XV6特定数据结构 (xv6_stat等)

### 6. 常量提取支持 (sys/syz-extract/)

#### 修改文件: `sys/syz-extract/extract.go`
- **修改内容**: 添加XV6提取器到extractors映射

#### 新增文件: `sys/syz-extract/xv6.go`
- **目的**: 从XV6源码中提取常量
- **主要功能**:
  - `prepare()`: 验证XV6源码目录
  - `prepareArch()`: 验证架构支持
  - `processFile()`: 处理常量提取
  - `extractConstant()`: 提取具体常量值

#### 支持的常量类型:
- 文件操作标志 (O_*)
- 系统调用号 (SYS_*)
- 文件类型 (S_IF*)
- XV6内置常量 (进程数、文件数限制等)

### 7. 目标平台定义 (sys/targets/)

#### 修改文件: `sys/targets/targets.go`
- **添加XV6常量**: `XV6 = "xv6"`
- **添加架构配置**:
  - RISC-V64: 8字节指针，4KB页面，RISC-V工具链
  - i386: 4字节指针，4KB页面，i386工具链
- **添加操作系统配置**:
  - 使用系统调用号
  - SYS_前缀
  - 不使用fork服务器（XV6较简单）
  - 内核对象为"kernel"
  - 静态链接

### 8. QEMU虚拟化支持 (vm/qemu/)

#### 修改文件: `vm/qemu/qemu.go`
- **添加架构配置**:
  - `xv6/riscv64`: 使用qemu-system-riscv64，virt机器，rv64 CPU
  - `xv6/386`: 使用qemu-system-i386
- **配置特性**:
  - 简单的网络设备配置
  - 不需要RNG设备
  - 基础串口控制台输出
  - 使用新的QEMU镜像选项（RISC-V）

## XV6系统的限制和特点

### 支持的功能:
1. **基础进程管理**: fork, exec, exit, wait, getpid
2. **文件操作**: open, close, read, write, lseek, dup
3. **文件系统**: mkdir, rmdir, unlink, link, chdir, stat
4. **管道通信**: pipe
5. **基础内存管理**: sbrk
6. **简单信号**: kill
7. **时间**: sleep, uptime

### 不支持的功能:
1. **网络**: 无socket支持
2. **高级内存管理**: 无复杂mmap/mprotect
3. **高级IPC**: 无信号量、共享内存、消息队列
4. **命名空间**: 无namespace支持
5. **安全**: 无capabilities、ACL等
6. **设备**: 有限的设备支持
7. **覆盖率**: 无内核覆盖率收集

## 使用方法

1. **准备XV6源码**: 确保有完整的XV6源码树
2. **安装工具链**: 安装RISC-V或i386交叉编译工具链
3. **配置syzkaller**: 设置目标OS为"xv6"，架构为"riscv64"或"386"
4. **运行**: 使用QEMU虚拟化启动XV6进行模糊测试

## 注意事项

1. **XV6简单性**: XV6是教学操作系统，功能有限，不要期望复杂特性
2. **工具链**: 需要正确的RISC-V或i386工具链
3. **内存限制**: XV6内存管理简单，避免大内存操作
4. **网络**: XV6没有网络支持，相关测试会被跳过
5. **覆盖率**: XV6没有内核覆盖率，测试主要依赖崩溃检测

## 未来改进

1. **更完整的系统调用描述**: 添加更多XV6特定的系统调用
2. **改进的错误检测**: 更好的XV6错误模式识别
3. **符号解析**: 支持XV6的符号解析和调试信息
4. **自动化测试**: 添加XV6的自动化测试用例
5. **文档**: 完善XV6模糊测试的最佳实践文档

## 总结

这个实现为syzkaller添加了完整的XV6支持，包括构建、执行、错误报告和虚拟化支持。虽然XV6功能有限，但这个实现为教学和研究环境提供了有价值的模糊测试能力。实现遵循了syzkaller的架构模式，并考虑了XV6的特殊性和限制。
