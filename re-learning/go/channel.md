# channel

## 设计原理

先进先出的 FIFO 队列

设计模式是：**不要通过共享内存的方式进行通信，而是通过通信的方式进行共享内存**。

## 数据结构

```go
type hchan struct {
 qcount   uint           // channel 中元素个数
 dataqsiz uint           // channel 中循环队列的长度
 buf      unsafe.Pointer // 缓冲区，也称为环形数组
 elemsize uint16         // 数据类型大小
 closed   uint32         // 是否关闭
 elemtype *_type         // 数据类型
 sendx    uint           // 发送操作处理到的位置
 recvx    uint           // 接收操作处理到的位置
 recvq    waitq          // 接收的阻塞队列
 sendq    waitq          // 发送的阻塞队列

 lock mutex
}
```
