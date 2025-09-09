# XV6 Syzkaller 设置指南

## 概述

本指南说明如何配置syzkaller以测试在QEMU上运行的XV6操作系统。

## 前置条件

### 1. 工具链
```bash
# RISC-V工具链（推荐）
sudo apt install gcc-riscv64-linux-gnu

# 或者 i386工具链（如果使用x86版本的XV6）
sudo apt install gcc-multilib
```

### 2. QEMU
```bash
sudo apt install qemu-system-riscv64 qemu-system-i386
```

### 3. XV6源码
从MIT获取XV6源码：
```bash
git clone https://github.com/mit-pdos/xv6-riscv.git
cd xv6-riscv
make
```

## Syzkaller配置

### 配置文件示例 (config.json)
```json
{
    "target": "xv6/riscv64",
    "http": "127.0.0.1:56741",
    "workdir": "./workdir",
    "kernel_obj": "/path/to/xv6-riscv/kernel",
    "kernel_src": "/path/to/xv6-riscv",
    "image": "/path/to/xv6-riscv/fs.img",
    "procs": 1,
    "type": "qemu",
    "vm": {
        "count": 1,
        "cpu": 1,
        "mem": 128,
        "kernel": "/path/to/xv6-riscv/kernel",
        "qemu": "qemu-system-riscv64",
        "qemu_args": "-machine virt -cpu rv64"
    }
}
```

### i386版本配置
```json
{
    "target": "xv6/386",
    "http": "127.0.0.1:56741", 
    "workdir": "./workdir",
    "kernel_obj": "/path/to/xv6-i386/kernel",
    "kernel_src": "/path/to/xv6-i386",
    "image": "/path/to/xv6-i386/fs.img",
    "procs": 1,
    "type": "qemu",
    "vm": {
        "count": 1,
        "cpu": 1,
        "mem": 128,
        "kernel": "/path/to/xv6-i386/kernel",
        "qemu": "qemu-system-i386"
    }
}
```

## 运行Syzkaller

### 1. 编译Syzkaller
```bash
make manager fuzzer executor
```

### 2. 启动管理器
```bash
./bin/syz-manager -config=config.json
```

### 3. 查看Web界面
访问 http://127.0.0.1:56741

## XV6限制和注意事项

### 支持的系统调用
- 进程管理: fork, exec, exit, wait, getpid
- 文件操作: open, close, read, write, lseek, dup
- 文件系统: mkdir, rmdir, unlink, link, chdir, stat
- 管道: pipe
- 内存: sbrk
- 其他: kill, sleep, uptime

### 不支持的功能
- 网络相关系统调用
- 高级内存管理 (mmap/mprotect)
- 复杂IPC机制
- 信号处理（除基础kill外）
- 设备驱动接口
- 内核覆盖率收集

### 性能调优
- 使用较小的内存配置 (128MB)
- 限制并发进程数量 (procs: 1)
- 不启用网络相关测试

## 故障排除

### 常见问题

1. **QEMU启动失败**
   - 检查QEMU路径和参数
   - 确认内核文件存在且可执行

2. **编译错误**
   - 验证工具链安装
   - 检查XV6源码完整性

3. **系统调用不支持**
   - 参考XV6系统调用列表
   - 某些Linux特有功能在XV6中不可用

4. **内存不足**
   - XV6内存有限，避免大量内存分配
   - 减少测试程序复杂度

### 调试选项
```json
{
    "debug": true,
    "cover": false,
    "leak": false
}
```

## 开发和测试

### 添加新系统调用
1. 在`sys/xv6/sys.txt`中定义
2. 更新`sys/syz-extract/xv6.go`中的常量
3. 在`pkg/vminfo/xv6.go`中添加支持检查

### 自定义测试
创建XV6特定的测试用例，专注于:
- 文件系统操作
- 进程创建和管理
- 基础内存操作
- 管道通信

## 参考资料

- [XV6官方文档](https://pdos.csail.mit.edu/6.828/2020/xv6.html)
- [Syzkaller文档](https://github.com/google/syzkaller/tree/master/docs)
- [QEMU RISC-V文档](https://www.qemu.org/docs/master/system/target-riscv.html)
