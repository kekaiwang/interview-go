# 操作系统相关

[toc]

## 1. 基础知识

### 1.1 如何监听端口

- **`netstat` 的命令用来显示网络状态**

  - `netstat -anp`：显示系统端口使用情况
  - `netstat -nupl`：UDP类型的端口
  - `netstat -ntpl`：TCP类型的端口
  - `netstat -na|grep ESTABLISHED|wc -l`：统计已连接上的，状态为"established"
  - `netstat -l`：只显示所有监听端口
  - `netstat -lt`：只显示所有监听tcp端口
- `ps -ef | grep 80`： 查看指定端口占用

## 1.2 查看内存使用情况

- **`free` 命令**

  - `total` 表示总共有 7822MB 的物理内存(RAM)，即7.6G。
  - `used` 表示物理内存的使用量，大约是 322M。
  - `free` 表示空闲内存;
  - `shared` 表示共享内存?;
  - `buff/cache` 表示缓存和缓冲内存量; Linux 系统会将很多东西缓存起来以提高性能，这部分内存可以在必要时进行释放，给其他程序使用。
  - `available` 表示可用内存;

- **`cat /proc/meminfo`**

  - `MemTotal`, 总内存
  - `MemFree`, 空闲内存
  - `MemAvailable`, 可用内存
  - `Buffers`, 缓冲
  - `Cached`, 缓存
  - `SwapTotal`, 交换内存
  - `SwapFree`, 空闲交换内存
- **`vmstat -s`**

- **`top` 命令一般用于查看进程的CPU和内存使用情况**

## 1.3 查看磁盘使用情况

- **`df -h`**
- `df -h /usr`： 查看指定目录
- `du --max-depth=1 -h`: 查看当前目录文件夹的使用情况
- `du -sh /usr/`: 计算文件夹大小

## Nginx

Nginx由内核和模块组成，其中，内核的设计非常微小和简洁，完成的工作也非常简单，仅仅通过查找配置文件将客户端请求映射到一个locationblock（location是Nginx配置中的一个指令，用于URL匹配），而在这个location中所配置的每个指令将会启动不同的模块去完成相应的工作。

Nginx的模块从结构上分为核心模块、基础模块和第三方模块：

- 核心模块：HTTP模块、EVENT模块和MAIL模块
- 基础模块：HTTP Access模块、HTTP FastCGI模块、HTTP Proxy模块和HTTP Rewrite模块，
- 第三方模块：HTTP Upstream RequestHash模块、Notice模块和HTTP Access Key模块。

Nginx的模块从功能上分为如下三类：

- **Handlers（处理器模块）**。此类模块直接处理请求，并进行输出内容和修改headers信息等操作。Handlers处理器模块一般只能有一个。
- **Filters （过滤器模块）**。此类模块主要对其他处理器模块输出的内容进行修改操作，最后由Nginx输出。
- **Proxies（代理类模块）**。此类模块是Nginx的HTTP Upstream之类的模块，这些模块主要与后端一些服务比如FastCGI等进行交互，实现服务代理和负载均衡等功能。
