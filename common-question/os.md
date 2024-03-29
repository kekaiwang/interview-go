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

## 如何给进程发信号？

`kill` 命令准确地说并不是 “杀死” 进程，而是给进程发送信号（signal）。
和文件一样，进程也有所有者，只有进程的所有者（或超级用户）才能使用 `kill` 命令来向它发送信号。

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

## fork() 的实现原理

`fork()` 函数是非常重要的函数，他从一个已存在的进程中创建一个新进程；新进程为子进程，而原进程称为父进程。

以当前进程作为父进程创建出一个新的子进程，并且将父进程的所有资源拷贝给子进程，这样子进程作为父进程的一个副本存在。父子进程几乎时完全相同的，但也有不同的如父子进程ID不同。

```c++
#include<unistd.h>
pid_t fork(void)
```

`pid_t` 是进程描述符，实质就是一个 `int`，**如果 fork 函数调用失败，返回一个负数，调用成功则返回两个值：0和子进程ID**。

- 该进程为父进程时，返回子进程的 pid
- 该进程为子进程时，返回0
- fork执行失败，返回 -1

fork() 系统调用通过复制一个现有进程来创建一个全新的进程。进程被存放在一个叫做任务队列的双向循环链表当中，链表当中的每一项都是类型为task_struct称为进程描述符的结构，也就是我们写过的进程PCB.

Tips：内核通过一个位置的进程标识值或PID来标识每一个进程。//最大值默认为32768，short int短整型的最大值.，他就是系统中允许同时存在的进程最大的数目。
可以到目录 /proc/sys/kernel中查看pid_max：

**当进程调用fork后，当控制转移到内核中的fork代码后，内核会做4件事情**:

1. 分配新的内存块和内核数据结构给子进程
2. 将父进程部分数据结构内容(数据空间，堆栈等）拷贝至子进程
3. 添加子进程到系统进程列表当中
4. fork返回，开始调度器调度

### 为什么fork成功调用后返回两个值?

由于在复制时复制了父进程的堆栈段，所以两个进程都停留在fork函数中，等待返回。所以fork函数会返回两次,一次是在父进程中返回，另一次是在子进程中返回，这两次的返回值不同  
其中父进程返回子进程pid，这是由于一个进程可以有多个子进程，但是却没有一个函数可以让一个进程来获得这些子进程id，那谈何给别人你创建出来的进程。而子进程返回0，这是由于子进程可以调用getppid获得其父进程进程ID,但这个父进程ID却不可能为0，因为进程ID0总是有内核交换进程所用，故返回0就可代表正常返回了。

## 内核态和用户态

**内核态**：cpu可以访问内存的所有数据，包括外围设备，例如硬盘，网卡，cpu也可以将自己从一个程序切换到另一个程序。

**用户态**：只能受限的访问内存，且不允许访问外围设备，占用cpu的能力被剥夺，cpu资源可以被其他程序获取。

### 为什么要有用户态和内核态？

由于需要限制不同的程序之间的访问能力, 防止他们获取别的程序的内存数据, 或者获取外围设备的数据, 并发送到网络, CPU划分出两个权限等级 -- 用户态和内核态。

## 进程、线程

### 僵尸进程、孤儿进程

**孤儿进程**：父进程退出，子进程还在运行的这些子进程都是孤儿进程，孤儿进程将被 init 进程(进程号为 1)所收养，并由 init 进程对它们完成状态收集工作。
**僵尸进程**：进程使用 fork 创建子进程，如果子进程退出，而父进程并没有调用 wait 或 waitpid 获取子进程的状态信息，那么子进程的进程描述符仍然保存在系统中的这些进程是僵尸进程

#### 避免僵尸进程的方法：

1. fork两次用孙子进程去完成子进程的任务；
2. 用wait()函数使父进程阻塞；
3. 使用信号量，在 signal handler 中调用waitpid，这样父进程不用阻塞。

## 死锁

### 产生条件

1. **互斥条件**：进程对所需求的资源具有排他性，若有其他进程请求该资源，请求进程只能等待。
2. **不剥夺条件**：进程在所获得的资源未释放前，不能被其他进程强行夺走，只能自己释放。
3. **请求和保持条件**：进程当前所拥有的资源在进程请求其他新资源时，由该进程继续占有。
4. **循环等待条件**：存在一种进程资源循环等待链，链中每个进程已获得的资源同时被链中下一个进程所请求。

### 避免死锁的方法

1. 死锁预防-----确保系统永远不会进入死锁状态
        产生死锁需要四个条件，那么，只要这四个条件中至少有一个条件得不到满足，就不可能发生死锁了。由于互斥条件是非共享资源所必须的，不仅不能改变，还应加以保证，所以，主要是破坏产生死锁的其他三个条件。
a、破坏“占有且等待”条件
    方法1：所有的进程在开始运行之前，必须一次性地申请其在整个运行过程中所需要的全部资源。
    优点：简单易实施且安全
    缺点：因为某项资源不满足，进程无法启动，而其他已经满足了的资源也不会得到利用，严重降低了资源的利用率，造成资源浪费。使进程经常发生饥饿现象。
    方法2：该方法是对第一种方法的改进，允许进程只获得运行初期需要的资源，便开始运行，在运行过程中逐步释放掉分配到的已经使用完毕的资源，然后再去请求新的资源。这样的话，资源利用率会得到提高，也会减少进程的饥饿问题。
b、破坏“不可抢占”条件
        当一个已经持有了一些资源的进程在提出新的资源请求没有得到满足时，它必须释放已经保持的所有资源，待以后需要使用的时候再重新申请。这就意味着进程已占有的资源会被短暂地释放或者说是被抢占了。
        该方法实现起来比较复杂，且代价也比较大。释放已经保持的资源很有可能会导致进程之前的工作实效等，反复的申请和释放资源会导致进程的执行被无限的推迟，这不仅会延长进程的周转周期，还会影响系统的吞吐量。
c、破坏“循环等待”条件
可以通过定义资源类型的线性顺序来预防，可将每个资源编号，当一个进程占有编号为i的资源时，那么它下一次申请资源只能申请编号大于i的资源。如图所示：

这样虽然避免了循环等待，但是这种歌方法是比较低效的，资源的执行速度回变慢，并且可能在没有必要的情况下拒绝资源的访问，比如说，进程c想要申请资源1，如果资源1并没有被其他进程占有，此时将它分配个进程c是没有问题的，但是为了避免产生循环等待，该申请会被拒绝，这样就降低了资源的利用率。
2. 避免死锁------在使用前进行判断，只允许不会产生死锁的进程申请资源
  死锁避免是利用额外的检验信息，在分配资源时判断是否会出现死锁，只在不会出现死锁的情况下才分配资源。
  两种避免办法：
  1、如果一个进程的请求会导致死锁，则不启动该进程。
  2、如果一个进程的增加资源请求会导致死锁，则拒绝该申请。
  避免死锁的具体实现通常利用银行家算法
3. 死锁检测与解除-------在检测到运行系统进入死锁，进行恢复。
允许系统进入死锁，如果利用死锁检测算法检测出系统已经出现了死锁，那么，此时就需要对系统采取相应的措施。常用的接触死锁的方法：
1）抢占资源：从一个或多个进程中抢占足够数量的资源分配给死锁进程，以解除死锁状态。
2）终止（或撤销）进程：终止或撤销系统中的一个或多个死锁进程，直至打破死锁状态。 