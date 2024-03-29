[toc]
# 网络

## 1、http相关知识

### http相关知识

#### 1.http1.0、1.1、2.0的区别

1.0是短链接，1.1是长链接

#### 1.1.1.一个 TCP 连接可以对应几个 HTTP 请求？

了解了第一个问题之后，其实这个问题已经有了答案，如果维持连接，一个 TCP 连接是可以发送多个 HTTP 请求的。

#### 1.1.2.一个 TCP 连接中 HTTP 请求发送可以一起发送么（比如一起发三个请求，再三个响应一起接收）？

HTTP/1.1 存在一个问题，单个 TCP 连接在同一时刻只能处理一个请求，意思是：两个请求的生命周期不能重叠，任意两个 HTTP 请求从开始到结束的时间在同一个 TCP 连接里不能重叠。虽然 HTTP/1.1 规范中规定了 Pipelining 来试图解决这个问题，但是这个功能在浏览器中默认是关闭的。但是，HTTP2 提供了 Multiplexing 多路传输特性，可以在一个 TCP 连接中同时完成多个 HTTP 请求。至于 Multiplexing 具体怎么实现的就是另一个问题了。我们可以看一下使用 HTTP2 的效果。

#### 1.1.3.为什么有的时候刷新页面不需要重新建立 SSL 连接？

在第一个问题的讨论中已经有答案了，TCP 连接有的时候会被浏览器和服务端维持一段时间。TCP 不需要重新建立，SSL 自然也会用之前的。

#### 1.1.4.浏览器对同一Host建立TCP连接到数量有没有限制？

Chrome 最多允许对同一个 Host 建立六个 TCP 连接。不同的浏览器有一些区别。
如果图片都是 HTTPS 连接并且在同一个域名下，那么浏览器在 SSL 握手之后会和服务器商量能不能用 HTTP2，如果能的话就使用 Multiplexing 功能在这个连接上进行多路传输。不过也未必会所有挂在这个域名的资源都会使用一个 TCP 连接去获取，但是可以确定的是 Multiplexing 很可能会被用到。
**如果发现用不了 HTTP2 呢**？或者用不了 HTTPS（现实中的 HTTP2 都是在 HTTPS 上实现的，所以也就是只能使用 HTTP/1.1）。那浏览器就会在一个 HOST 上建立多个 TCP 连接，连接数量的最大限制取决于浏览器设置，这些连接会在空闲的时候被浏览器用来发送新的请求，如果所有的连接都正在发送请求呢？那其他的请求就只能等等了。

### 2.HTTP 1.0 和 1.1、1.2 的主要变化？

#### 1.2.1.HTTP1.1 的主要变化

1. HTTP1.0 经过多年发展，在 1.1 提出了改进。首先是提出了长连接，HTTP 可以在一次 TCP 连接中不断发送请求。
2. 然后 HTTP1.1 支持只发送 header 而不发送 body。原因是先用 header 判断能否成功，再发数据，节约带宽，事实上，post 请求默认就是这样做的。
3. HTTP1.1 的 host 字段。由于虚拟主机可以支持多个域名，所以一般将域名解析后得到 host。

#### 1.2.2.HTTP2.0 的主要变化

1. 支持多路复用，同一个连接可以并发处理多个请求，方法是把 HTTP数据包拆为多个帧，并发有序的发送，根据序号在另一端进行重组，而不需要一个个 HTTP请求顺序到达；
2. 支持服务端推送，就是服务端在 HTTP 请求到达后，除了返回数据之外，还推送了额外的内容给客户端；
3. 压缩了请求头，同时基本单位是二进制帧流，这样的数据占用空间更少；
4. 适用于 HTTPS 场景，因为其在 HTTP和 TCP 中间加了一层 SSL 层。

### 3.HTTP 长连接和短连接的理解？分别应用于哪些场景？

在 HTTP/1.0 中默认使用短连接。
也就是说，客户端和服务器每进行一次 HTTP 操作，就建立一次连接，任务结束就中断连接。当客户端浏览器访问的某个 HTML 或其他类型的 Web 页中包含有其他的 Web 资源（如：JavaScript 文件、图像文件、CSS 文件等），每遇到这样一个 Web 资源，浏览器就会重新建立一个 HTTP 会话。
从 HTTP/1.1 起，默认使用长连接，用以保持连接特性。使用长连接的 HTTP 协议，会在响应头加入这行代码`Connection:keep-alive`在使用长连接的情况下，当一个网页打开完成后，客户端和服务器之间用于传输 HTTP 数据的 TCP 连接不会关闭，客户端再次访问这个服务器时，会继续使用这一条已经建立的连接。Keep-Alive 不会永久保持连接，它有一个保持时间，可以在不同的服务器软件（如：Apache）中设定这个时间。实现长连接需要客户端和服务端都支持长连接。

### 4.HTTPS 的优缺点及HTTP与HTTPS的区别？

#### 1.4.1.HTTPS优点

1. 使用 HTTPS 协议可认证用户和服务器，确保数据发送到正确的客户机和服务器；
2. HTTPS 协议是由 SSL + HTTP 协议构建的可进行加密传输、身份认证的网络协议，要比 HTTP 协议安全，可防止数据在传输过程中不被窃取、改变，确保数据的完整性；
3. HTTPS 是现行架构下最安全的解决方案，虽然不是绝对安全，但它大幅增加了中间人攻击的成本。

#### 1.4.2.HTTPS缺点

1. HTTPS 协议握手阶段比较费时，会使页面的加载时间延长近 50%，增加 10% 到 20% 的耗电；
2. HTTPS 连接缓存不如 HTTP 高效，会增加数据开销和功耗，甚至已有的安全措施也会因此而受到影响；
3. SSL 证书需要钱，功能越强大的证书费用越高，个人网站、小网站没有必要一般不会用；
4. SSL 证书通常需要绑定 IP，不能在同一 IP 上绑定多个域名，IPv4 资源不可能支撑这个消耗；
5. HTTPS 协议的加密范围也比较有限，在黑客攻击、拒绝服务攻击、服务器劫持等方面几乎起不到什么作用。最关键的，SSL 证书的信用链体系并不安全，特别是在某些国家可以控制 CA 根证书的情况下，中间人攻击一样可行。

> 443端口即网页浏览端口，主要是用于HTTPS服务，网页浏览端口，能提供加密和通过安全端口传输的另一种HTTP。

#### 1.4.3.http与https的区别

1.http时超文本传输协议，明文传输，https是具有安全性的ssl加密传输。
2.端口http默认80，https默认443。
3.http是无状态协议(无状态是指对事务没有记忆能力，缺少状态意味对后续处理需要的信息没办法提供)，https是ssl+http构建的可进行加密传输身份认证的网络协议。
4.https协议需要证书，一般需要收费。

#### 1.4.4.https的过程

1. client向server发送请求 `https://baidu.com`，然后连接到server的443端口，发送的信息主要是随机值1和客户端支持的加密算法。
2. server接收到信息之后给予client响应握手信息，包括随机值2和匹配好的协商加密算法，这个加密算法一定是client发送给server加密算法的子集。
3. 随即server给client发送第二个响应报文是数字证书。服务端必须要有一套数字证书，可以自己制作，也可以向组织申请。区别就是自己颁发的证书需要客户端验证通过，才可以继续访问，而使用受信任的公司申请的证书则不会弹出提示页面，这套证书其实就是一对公钥和私钥。传送证书，这个证书其实就是公钥，只是包含了很多信息，如证书的颁发机构，过期时间、服务端的公钥，第三方证书认证机构(CA)的签名，服务端的域名信息等内容。
4. 客户端解析证书，这部分工作是由客户端的TLS来完成的，首先会验证公钥是否有效，比如颁发机构，过期时间等等，如果发现异常，则会弹出一个警告框，提示证书存在问题。如果证书没有问题，那么就生成一个随即值（预主秘钥）。
5. 客户端认证证书通过之后，接下来是通过随机值1、随机值2和预主秘钥组装会话秘钥。然后通过证书的公钥加密会话秘钥。
6. 传送加密信息，这部分传送的是用证书加密后的会话秘钥，目的就是让服务端使用秘钥解密得到随机值1、随机值2和预主秘钥。
7. 服务端解密得到随机值1、随机值2和预主秘钥，然后组装会话秘钥，跟客户端会话秘钥相同。
8. 客户端通过会话秘钥加密一条消息发送给服务端，主要验证服务端是否正常接受客户端加密的消息。
9. 同样服务端也会通过会话秘钥加密一条消息回传给客户端，如果客户端能够正常接受的话表明SSL层连接建立完成了。

>TLS（Transport Layer Security，安全传输层)，TLS是建立在传输层TCP协议之上的协议，服务于应用层，它的前身是SSL（Secure Socket Layer，安全套接字层），它实现了将应用层的报文进行加密后再交由TCP进行传输的功能。

### 5.http的方法有哪些

#### 1.5.1 客户端发送的请求报文第一行为请求行，包含了方法字段。

1. GET：获取资源，当前网络中绝大部分使用的都是 GET；
2. HEAD：获取报文首部，和 GET 方法类似，但是不返回报文实体主体部分；
3. POST：传输实体，提交数据进行处理请求
4. PUT：上传文件，由于自身不带验证机制，任何人都可以上传文件，因此存在安全性问题，一般不使用该方法。
5. PATCH：对资源进行部分修改。PUT 也可以用于修改资源，但是只能完全替代原始资源，PATCH 允许部分修改。
6. OPTIONS：查询指定的 URL 支持的方法.
7. CONNECT：要求在与代理服务器通信时建立隧道。使用 SSL（Secure Sockets Layer，安全套接层）和 TLS（Transport Layer Security，传输层安全）协议把通信内容加密后经网络隧道传输。
8. TRACE：追踪路径。服务器会将通信路径返回给客户端。发送请求时，在 Max-Forwards 首部字段中填入数值，每经过一个服务器就会减 1，当数值为 0 时就停止传输。通常不会使用 TRACE，并且它容易受到 XST 攻击（Cross-Site Tracing，跨站追踪）。

#### 1.5.2 HTTP状态码

- 200：请求成功
- 300：针对请求，服务器可执行多种操作。 服务器可根据请求者 (user agent) 选择一项操作，或提供操作列表供请求者选择。
- 301：永久重定向，资源被永久转移到其他URI，会返回新的URI，今后新请求应使用新URI
- 302：临时重定向，客户端应继续使用原有URI304：所请求的资源未修改，服务器不返回任何资源
- 400：Bad Request,请求有语法问题
- 401：用户身份验证未提供或未通过403：服务器拒绝执行此请求
- 404：请求的资源不存在、
- 499: 客户关闭连接 /* 499, client has closed connection */。proxy_ignore_client_abort on; //&nbsp;&nbsp;Don’t know if this is safe.
- 500：内部服务器错误501：服务器不具备完成请求的功能。例如无法识别请求方法时可能返回。
- 502：bad Gateway503：service unavailable服务不可用
- 504: Gateway Timeout/网关超时。
  - fastcgi_connect_timeout、fastcgi_send_timeout、fastcgi_read_timeout、fastcgi_buffer_size、fastcgi_buffers、fastcgi_busy_buffers_size、fastcgi_temp_file_write_size、fastcgi_intercept_errors。特别是前三个超时时间。如果fastcgi缓冲区太小会导致fastcgi进程被挂起从而演变为504错误。

## 2、五层网络协议

OSI分层（7层）：物理层、数据链路层、   网络层、传输层、会话层、表示层、应用层。
TCP/IP分层（4层）：网络接口层、  网际层、运输层、 应用层。
五层协议（5层）：物理层、数据链路层、   网络层、运输层、 应用层。

1. 应用层应用层（application-layer）的任务是通过应用进程间的交互来完成特定网络应用。应用层协议定义的是应用进程（进程：主机中正在运行的程序）间的通信和交互的规则。对于不同的网络应用需要不同的应用层协议。在互联网中应用层协议很多，如域名系统 DNS，支持万维网应用的 HTTP 协议，支持电子邮件的 SMTP 协议等等。我们把应用层交互的数据单元称为报文。
HTTP-超文本传输协议、FTP-文件传输、SMTP-邮件传输协议、DNS-域名系统、SSH-安全外壳协议。
2. 传输层运输层（transport layer）的主要任务就是负责向两台主机进程之间的通信提供通用的数据传输服务。应用进程利用该服务传送应用层报文。“通用的”是指并不针对某一个特定的网络应用，而是多种应用可以使用同一个运输层服务。由于一台主机可同时运行多个线程，因此运输层有复用和分用的功能。所谓复用就是指多个应用层进程可同时使用下面运输层的服务，分用和复用相反，是运输层把收到的信息分别交付上面应用层中的相应进程。
TCP-传输控制协议、UDP-用户数据报文协议。
3. 网络层在计算机网络中进行通信的两个计算机之间可能会经过很多个数据链路，也可能还要经过很多通信子网。网络层的任务就是选择合适的网间路由和交换结点， 确保数据及时传送。在发送数据时，网络层把运输层产生的报文段或用户数据报封装成分组和包进行传送。
在 TCP / IP 体系结构中，由于网络层使用 IP 协议，因此分组也叫 IP 数据报，简称数据报。IP-网际协议、ARP-地址转换协议、RIP-路由信息协议
4. 数据链路层数据链路层（data link layer）通常简称为链路层。两台主机之间的数据传输，总是在一段一段的链路上传送的，这就需要使用专门的链路层的协议。在两个相邻节点之间传送数据时，数据链路层将网络层交下来的 IP 数据报组装成帧，在两个相邻节点间的链路上传送帧。每一帧包括数据和必要的控制信息（如：同步信息，地址信息，差错控制等）。在接收数据时，控制信息使接收端能够知道一个帧从哪个比特开始和到哪个比特结束。这样，数据链路层在收到一个帧后，就可从中提出数据部分，上交给网络层。控制信息还使接收端能够检测到所收到的帧中有无差错。如果发现差错，数据链路层就简单地丢弃这个出了差错的帧，以避免继续在网络中传送下去白白浪费网络资源。如果需要改正数据在链路层传输时出现差错（这就是说，数据链路层不仅要检错，而且还要纠错），那么就要采用可靠性传输协议来纠正出现的差错。这种方法会使链路层的协议复杂些。
ARQ-自动重传协议、PPP-点对点协议。
5. 物理层在物理层上所传送的数据单位是比特。物理层（physical layer）的作用是实现相邻计算机节点之间比特流的透明传送，尽可能屏蔽掉具体传输介质和物理设备的差异。使其上面的数据链路层不必考虑网络的具体传输介质是什么。“透明传送比特流”表示经实际电路传送后的比特流没有发生变化，对传送的比特流来说，这个电路好像是看不见的。中继器、集线器。
6. ARP协议(网络层)网络层的 ARP 协议完成了IP地址与物理地址的映射。ARP 请求数据包里包括源主机的 IP 地址、硬件地址、以及目的主机的 IP 地址。每台主机都会在自己的ARP缓冲区中建立一个ARP列表，以表示IP地址和MAC 地址的对应关系。当源主机需要将一个数据包要发送到目的主机时，会首先检查自己 ARP 列表中是否存在该 IP 地址对应的MAC地址：如果有，就直接将数据包发送到这个 MAC 地址；如果没有，就向本地网段发起一个 ARP 请求的广播包，查询此目的主机对应的 MAC 地址。

## 3、TCP/IP部分

### 1.TCP 的主要特点是什么？

1. TCP 是面向连接的。（就好像打电话一样，通话前需要先拨号建立连接，通话结束后要挂机释放连接）
2. 每一条 TCP 连接只能有两个端点，每一条 TCP 连接只能是点对点的（一对一）；
3. TCP 提供可靠交付的服务。通过 TCP 连接传送的数据，无差错、不丢失、不重复、并且按序到达；
4. TCP 提供全双工通信。TCP 允许通信双方的应用进程在任何时候都能发送数据。TCP 连接的两端都设有发送缓存和接收缓存，用来临时存放双方通信的数据；
5. 面向字节流。TCP 中的“流”（Stream）指的是流入进程或从进程流出的字节序列。“面向字节流”的含义是：虽然应用程序和 TCP 的交互是一次一个数据块（大小不等），但 TCP 把应用程序交下来的数据仅仅看成是一连串的无结构的字节流。

### 2.UDP 的主要特点是什么？

1. UDP 是无连接的；
2. UDP 使用尽最大努力交付，即不保证可靠交付，因此主机不需要维持复杂的链接状态（这里面有许多参数）；
3. UDP 是面向报文的(面向报文的传输方式是应用层交给UDP多长的报文，UDP就照样发送，即一次发送一个报文。因此，应用程序必须选择合适大小的报文)；
4. UDP 没有拥塞控制，因此网络出现拥塞不会使源主机的发送速率降低（对实时应用很有用，如 直播，实时视频会议等）；
5. UDP 支持一对一、一对多、多对一和多对多的交互通信；
6. UDP 的首部开销小，只有 8 个字节，比 TCP 的 20 个字节的首部要短。

### 3.TCP 协议是如何保证可靠传输的？

1. 数据包校验：目的是检测数据在传输过程中的任何变化，若校验出包有错，则丢弃报文段并且不给出响应，这时 TCP 发送数据端超时后会重发数据；
2. 对失序数据包重排序：既然 TCP 报文段作为 IP 数据报来传输，而 IP 数据报的到达可能会失序，因此 TCP 报文段的到达也可能会失序。TCP 将对失序数据进行重新排序，然后才交给应用层；
3. 丢弃重复数据：对于重复数据，能够丢弃重复数据；
4. 应答机制：当 TCP 收到发自 TCP 连接另一端的数据，它将发送一个确认。这个确认不是立即发送，通常将推迟几分之一秒；
5. 超时重发：当 TCP 发出一个段后，它启动一个定时器，等待目的端确认收到这个报文段。如果不能及时收到一个确认，将重发这个报文段；
6. 流量控制：TCP 连接的每一方都有固定大小的缓冲空间。TCP 的接收端只允许另一端发送接收端缓冲区所能接纳的数据，这可以防止较快主机致使较慢主机的缓冲区溢出，这就是流量控制。TCP 使用的流量控制协议是可变大小的滑动窗口协议。

### 4.TCP 对应的应用层协议

- FTP：定义了文件传输协议，使用 21 端口。
- Telnet：它是一种用于远程登陆的端口，23 端口
- SMTP：定义了简单邮件传送协议，25 端口
- POP3：它是和 SMTP 对应，POP3用于接收邮件。110 端口。
- HTTPde：从 Web 服务器传输超文本到本地浏览器的传送协议。

### 5.UDP 对应的应用层协议

- DNS：用于域名解析服务，将域名地址转换为 IP 地址。DNS 用的是 53 号端口。
- SNMP：简单网络管理协议，使用 161 号端口，是用来管理网络设备的。由于网络设备很多，无连接的服务就体现出其优势。
- TFTP(Trival File Transfer Protocal)：简单文件传输协议，该协议在熟知端口 69 上使用 UDP 服务。

### 6.为什么 TCP 叫数据流模式？ UDP 叫数据报模式？

1. 所谓的“流模式”，是指TCP发送端发送几次数据和接收端接收几次数据是没有必然联系的，比如你通过 TCP连接给另一端发送数据，你只调用了一次 write，发送了100个字节，但是对方可以分10次收完，每次10个字节；你也可以调用10次write，每次10个字节，但是对方可以一次就收完。
原因：这是因为TCP是面向连接的，一个 socket 中收到的数据都是由同一台主机发出，且有序地到达，所以每次读取多少数据都可以。
2. 所谓的“数据报模式”，是指UDP发送端调用了几次 write，接收端必须用相同次数的 read 读完。
UDP是基于报文的，在接收的时候，每次最多只能读取一个报文，报文和报文是不会合并的，如果缓冲区小于报文长度，则多出的部分会被丢弃。原因：这是因为UDP是无连接的，只要知道接收端的 IP 和端口，任何主机都可以向接收端发送数据。 这时候，如果一次能读取超过一个报文的数据， 则会乱套。

### 7.谈谈你对停止等待协议的理解？

停止等待协议是为了实现可靠传输的，它的基本原理就是每发完一个分组就停止发送，等待对方确认。在收到确认后再发下一个分组；在停止等待协议中，若接收方收到重复分组，就丢弃该分组，但同时还要发送确认。
主要包括以下几种情况：无差错情况、出现差错情况（超时重传）、确认丢失和确认迟到。

### 8.谈谈你对滑动窗口的了解？

TCP 利用滑动窗口实现流量控制的机制。
滑动窗口（Sliding window）是一种流量控制技术。早期的网络通信中，通信双方不会考虑网络的拥挤情况直接发送数据。由于大家不知道网络拥塞状况，同时发送数据，导致中间节点阻塞掉包，谁也发不了数据，所以就有了滑动窗口机制来解决此问题。
TCP 中采用滑动窗口来进行传输控制，滑动窗口的大小意味着接收方还有多大的缓冲区可以用于接收数据。发送方可以通过滑动窗口的大小来确定应该发送多少字节的数据。当滑动窗口为 0 时，发送方一般不能再发送数据报，但有两种情况除外，一种情况是可以发送紧急数据，例如，允许用户终止在远端机上的运行进程。另一种情况是发送方可以发送一个 1 字节的数据报来通知接收方重新声明它希望接收的下一字节及发送方的滑动窗口大小。

### 9.TCP 拥塞控制的理解？使用了哪些算法？

拥塞控制和流量控制不同，前者是一个全局性的过程，而后者指点对点通信量的控制。
在某段时间，若对网络中某一资源的需求超过了该资源所能提供的可用部分，网络的性能就要变坏。这种情况就叫拥塞。
拥塞控制就是为了防止过多的数据注入到网络中，这样就可以使网络中的路由器或链路不致于过载。拥塞控制所要做的都有一个前提，就是网络能够承受现有的网络负荷。拥塞控制是一个全局性的过程，涉及到所有的主机，所有的路由器，以及与降低网络传输性能有关的所有因素。相反，流量控制往往是点对点通信量的控制，是个端到端的问题。流量控制所要做到的就是抑制发送端发送数据的速率，以便使接收端来得及接收。
为了进行拥塞控制，TCP 发送方要维持一个拥塞窗口(cwnd) 的状态变量。拥塞控制窗口的大小取决于网络的拥塞程度，并且动态变化。发送方让自己的发送窗口取为拥塞窗口和接收方的接受窗口中较小的一个。
TCP 的拥塞控制采用了四种算法，即：**慢开始、拥塞避免、快重传和快恢复**。在网络层也可以使路由器采用适当的分组丢弃策略（如：主动队列管理 AQM），以减少网络拥塞的发生。

- 慢开始：慢开始算法的思路是当主机开始发送数据时，如果立即把大量数据字节注入到网络，那么可能会引起网络阻塞，因为现在还不知道网络的符合情况。经验表明，较好的方法是先探测一下，即由小到大逐渐增大发送窗口，也就是由小到大逐渐增大拥塞窗口数值。cwnd 初始值为 1，每经过一个传播轮次，cwnd 加倍。
- 拥塞避免：拥塞避免算法的思路是让拥塞窗口 cwnd 缓慢增大，即每经过一个往返时间 RTT 就把发送方的 cwnd 加 1。
- 快重传与快恢复：在 TCP/IP 中，快速重传和快恢复（fast retransmit and recovery，FRR）是一种拥塞控制算法，它能快速恢复丢失的数据包。没有 FRR，如果数据包丢失了，TCP 将会使用定时器来要求传输暂停。在暂停的这段时间内，没有新的或复制的数据包被发送。有了 FRR，如果接收机接收到一个不按顺序的数据段，它会立即给发送机发送一个重复确认。如果发送机接收到三个重复确认，它会假定确认件指出的数据段丢失了，并立即重传这些丢失的数据段。有了 FRR，就不会因为重传时要求的暂停被耽误。当有单独的数据包丢失时，快速重传和快恢复（FRR）能最有效地工作。当有多个数据信息包在某一段很短的时间内丢失时，它则不能很有效地工作。

### 10.IP地址分类的理解

IP 地址是指互联网协议地址，是 IP 协议提供的一种统一的地址格式，它为互联网上的每一个网络和每一台主机分配一个逻辑地址，以此来屏蔽物理地址的差异。
IP 地址编址方案将 IP 地址空间划分为 A、B、C、D、E 五类，其中 A、B、C 是基本类，D、E 类作为多播和保留使用，为特殊地址。
每个 IP 地址包括两个标识码（ID），即网络 ID 和主机 ID。同一个物理网络上的所有主机都使用同一个网络 ID，网络上的一个主机（包括网络上工作站，服务器和路由器等）有一个主机 ID 与其对应。
A~E 类地址的特点如下：

- A 类地址：以 0 开头，第一个字节范围：0~127；
- B 类地址：以 10 开头，第一个字节范围：128~191；
- C 类地址：以 110 开头，第一个字节范围：192~223；
- D 类地址：以 1110 开头，第一个字节范围为 224~239；
- E 类地址：以 1111 开头，保留地址

#### 3.10.1.ICMP协议

IP协议并不是一个可靠的协议，它不保证数据被送达，那么，自然的，保证数据送达的工作应该由其他的模块来完成。
其中一个重要的模块就是ICMP(网络控制报文Internet Control Message Protocol)协议。ICMP不是高层协议，而是IP层的协议。
当传送IP数据包发生错误。比如主机不可达，路由不可达等等，ICMP协议将会把错误信息封包，然后传送回给主机。给主机一个处理错误的机会，这也就是为什么说建立在IP层以上的协议是可能做到安全的原因。
ping可以说是ICMP的最著名的应用，是TCP/IP协议的一部分。利用“ping”命令可以检查网络是否连通，可以很好地帮助我们分析和判定网络故障。

### 11.Pv6与IPv4的区别

1. IPv6的地址空间更大。IPv4中规定IP地址长度为32,即有2^32-1个地址；而IPv6中IP地址的长度为128,即有2^128-1个地址。夸张点说就是，如果IPV6被广泛应用以后，全世界的每一粒沙子都会有相对应的一个IP地址。
2. IPv6的路由表更小。IPv6的地址分配一开始就遵循聚类(Aggregation)的原则,这使得路由器能在路由表中用一条记录(Entry)表示一片子网,大大减小了路由器中路由表的长度,提高了路由器转发数据包的速度。
3. IPv6的组播支持以及对流的支持增强。这使得网络上的多媒体应用有了长足发展的机会，为服务质量控制提供了良好的网络平台。
4. IPv6加入了对自动配置的支持。这是对DHCP协议的改进和扩展，使得网络(尤其是局域网)的管理更加方便和快捷。
5. IPv6具有更高的安全性。在使用IPv6网络中，用户可以对网络层的数据进行加密并对IP报文进行校验，这极大地增强了网络安全。
6. IPv4 使用 ARP 来查找与 IPv4 地址相关联的物理地址（如 MAC 或链路地址）。在 IPv6 中，地址作用域是该体系结构的一部分。单点广播地址有两个已定义的作用域，包括本地链路和全局链路；而多点广播地址有 14 个作用域。为源和目标选择缺省地址时要考虑作用域。

#### 3.11.1.DHCP: 动态主机配置协议

TCP/IP协议想要运行正常的话，网络中的主机和路由器不可避免地需要配置一些信息（如接口的IP地址等）。有了这些配置信息主机/路由器才能提供/使用特定的网络服务。
DHCP 主要分为两部分： 地址的管理和配置信息的传递。

- 地址管理：地址管理处理IP地址的动态分配、向客户端提供地址租约
- 配置信息的传递：包含DHCP报文格式、状态机

### 12.什么是粘包？

在进行 Java NIO 学习时，可能会发现：如果客户端连续不断的向服务端发送数据包时，服务端接收的数据会出现两个数据包粘在一起的情况。

1. TCP 是基于字节流的，虽然应用层和 TCP 传输层之间的数据交互是大小不等的数据块，但是 TCP 把这些数据块仅仅看成一连串无结构的字节流，没有边界；
2. 从 TCP 的帧结构也可以看出，在 TCP 的首部没有表示数据长度的字段。
基于上面两点，在使用 TCP 传输数据时，才有粘包或者拆包现象发生的可能。一个数据包中包含了发送端发送的两个数据包的信息，这种现象即为粘包。接收端收到了两个数据包，但是这两个数据包要么是不完整的，要么就是多出来一块，这种情况即发生了拆包和粘包。拆包和粘包的问题导致接收端在处理的时候会非常困难，因为无法区分一个完整的数据包。

#### 3.12.1.TCP 黏包是怎么产生的？

1. 发送方产生粘包
采用 TCP 协议传输数据的客户端与服务器经常是保持一个长连接的状态（一次连接发一次数据不存在粘包），双方在连接不断开的情况下，可以一直传输数据。但当发送的数据包过于的小时，那么 TCP 协议默认的会启用 Nagle 算法，将这些较小的数据包进行合并发送（缓冲区数据发送是一个堆压的过程）；这个合并过程就是在发送缓冲区中进行的，也就是说数据发送出来它已经是粘包的状态了。
2. 接收方产生粘包
接收方采用 TCP 协议接收数据时的过程是这样的：数据到接收方，从网络模型的下方传递至传输层，传输层的 TCP 协议处理是将其放置接收缓冲区，然后由应用层来主动获取（C 语言用 recv、read 等函数）；这时会出现一个问题，就是我们在程序中调用的读取数据函数不能及时的把缓冲区中的数据拿出来，而下一个数据又到来并有一部分放入的缓冲区末尾，等我们读取数据时就是一个粘包。（放数据的速度 > 应用层拿数据速度）

### 13.封装和分用

1. 封装
数据在发送端从上到下经过TCP/IP协议栈，应用层->TCP/UDP->IP->链路层，当某层的一个协议数据单元（PDU）对象转换为由底层携带的数据格式表示，这个过程称为在相邻低层的封装，即上层被封装对象作为不透明数据充当底层的Payload部分，封装是层层包裹的过程。每层都有自己的消息对象（PDU）的概念，TCP层的PDU叫TCP段（segment），UDP层的PDU叫UDP数据报（Datagram），IP层的PDU叫IP数据报（Datagram），链路层的PDU叫链路层帧（Frame）。
封装的本质是将来自上层的数据看成不透明、无须解释的信息，经过本层的处理，在上层PDU的前面加上本层协议的头部，有些协议是增加尾部（链路层），头部用于在发送时复用数据，接收方基于各层封装过程中增加头部中的分解标识符执行分解。具体到TCP传输数据而言，发送端的数据要经过4次封装，应用层数据经过TCP层的时候，会增加TCP头部，产生TCP Segment，TCP头部中的端口号是该层的分解标识符；经过IP层的时候，会增加IP头部，产生IP Datagrame，IP头部中的协议类型字段是该层的分解标识符；经过链路层的时候，会增加以太网首部、尾部，产生以太网帧，帧头部中的以太网类型字段，可用于区分IPv4(0x8000)、IPv6(0x86DD)和ARP(0x0806)。
2. 分用数据到达接收端（是目的机器），会经过从下到上经过TCP/IP协议栈，链路层->IP->TCP/UDP->应用层。经过链路层会剥离以太网首尾部，根据以太网类型字段，如果是IP Datagram则交给IP层；经过IP层会清除IP头部，根据IP头部中的协议类型字段，交给TCP、UDP或者ICMP、IGMP；经过TCP/UDP层去掉TCP/UDP头部，根据端口号，最终将数据还原取出交付给应用程序。
封装发生在发送方，拆封（还原）发生在接收方。
![image](https://mail.wangkekai.cn/v2-01195084cae19c67f0e178d871480384_hd.jpg)

### 14.保活计时器

TCP还设有一个保活计时器，显然，客户端如果出现故障，服务器不能一直等下去，白白浪费资源。
服务器每收到一次客户端的请求后都会重新复位这个计时器，时间通常是设置为2小时，若两小时还没有收到客户端的任何数据，服务器就会发送一个探测报文段，以后每隔75分钟发送一次。若一连发送10个探测报文仍然没反应，服务器就认为客户端出了故障，接着就关闭连接。

### 15.交换机、路由器

1. 交换机
在计算机网络系统中，交换机是针对共享工作模式的弱点而推出的。
交换机拥有一条高带宽的背部总线和内部交换矩阵。交换机的所有的端口都挂接在这条背部总线上，当控制电路收到数据包以后，处理端口会查找内存中的地址对照表以确定目的MAC（网卡的硬件地址）的NIC（网卡）挂接在哪个端口上，通过内部交换矩阵迅速将数据包传送到目的端口。
目的MAC若不存在，交换机才广播到所有的端口，接收端口回应后交换机会“学习”新的地址，并把它添加入内部地址表中。交换机工作于OSI参考模型的第二层，即数据链路层。交换机内部的CPU会在每个端口成功连接时，通过ARP协议学习它的MAC地址，保存成一张ARP表。在今后的通讯中，发往该MAC地址的数据包将仅送往其对应的端口，而不是所有的端口。因此，交换机可用于划分数据链路层广播，即冲突域；但它不能划分网络层广播，即广播域。
2. 路由器
路由器（Router）是一种计算机网络设备，提供了路由与转发两种重要机制，可以决定数据包从来源端到目的端所经过的路由路径（host到host之间的传输路径），这个过程称为路由；将路由器输入端的数据包移送至适当的路由器输出端(在路由器内部进行)，这称为转送。路由工作在OSI模型的第三层——即网络层，例如IP协议。路由器的一个作用是连通不同的网络，另一个作用是选择信息传送的线路。 路由器与交换器的差别，路由器是属于OSI第三层的产品，交换器是OSI第二层的产品(这里特指二层交换机)。

### 16.什么是子网掩码

子网掩码是标志两个IP地址是否同属于一个子网的，也是32位二进制地址，其每一个为1代表该位是网络位，为0代表主机位。
它和IP地址一样也是使用点式十进制来表示的。如果两个IP地址在子网掩码的按位与的计算下所得结果相同，即表明它们共属于同一子网中。

### 17.网关和代理

代理连接的是两个或多个使用相同协议的应用程序，而网关连接的则是两个或多个使用不同协议的端点。
网关扮演的是“协议转换器”的角色。Web网关在一侧使用HTTP协议，在另一侧使用另一种协议。
<客户端协议>/<服务器端协议>（
HTTP/）服务器端网关：通过HTTP协议 与客户端对话，通过其他协议与服务器通信。
（/HTTP）客户端网关：通过其他协议与客户端对话，通过HTTP协议与服务器通信。

常见的网关有：

1. （HTTP/*）服务器端Web网关
2. （HTTP/HTTPS）服务器端安全网关
3. （HTTPS/HTTP）客户端安全加速器网关

## 4、三次握手和四次挥手

### 1.三次握手

三次握手（Three-way Handshake）其实就是指建立一个TCP连接时，需要客户端和服务器总共发送3个包。
进行三次握手的主要作用就是为了确认双方的接收能力和发送能力是否正常、指定自己的初始化序列号为后面的可靠性传送做准备。
实质上其实就是连接服务器指定端口，建立TCP连接，并同步连接双方的序列号和确认号，交换TCP窗口大小信息。
刚开始客户端处于 Closed 的状态，服务端处于 Listen 状态。

进行三次握手：

1. 第一次握手：客户端给服务端发一个 SYN 报文，并指明客户端的初始化序列号 ISN©。此时客户端处于 SYN_SEND 状态。首部的同步位SYN=1，初始序号seq=x，SYN=1的报文段不能携带数据，但要消耗掉一个序号。
2. 第二次握手：服务器收到客户端的 SYN 报文之后，会以自己的 SYN 报文作为应答，并且也是指定了自己的初始化序列号 ISN(s)。同时会把客户端的 ISN + 1 作为ACK 的值，表示自己已经收到了客户端的 SYN，此时服务器处于 SYN_REVD 的状态。在确认报文段中SYN=1，ACK=1，确认号ack=x+1，初始序号seq=y。
3. 第三次握手：客户端收到 SYN 报文之后，会发送一个 ACK 报文，当然，也是一样把服务器的 ISN + 1 作为 ACK 的值，表示已经收到了服务端的 SYN 报文，此时客户端处于 ESTABLISHED 状态。服务器收到 ACK 报文之后，也处于 ESTABLISHED 状态，此时，双方已建立起了连接。
确认报文段ACK=1，确认号ack=y+1，序号seq=x+1（初始为seq=x，第二个报文段所以要+1），ACK报文段可以携带数据，不携带数据则不消耗序号。发送第一个SYN的一端将执行主动打开（active open），接收这个SYN并发回下一个SYN的另一端执行被动打开（passive open）。

在socket编程中，客户端执行connect()时，将触发三次握手。
![image](https://mail.wangkekai.cn/v2-2a54823bd63e16674874aa46a67c6c72_r.jpg)

#### 4.1.1.为什么需要三次握手，两次不行吗？

弄清这个问题，我们需要先弄明白三次握手的目的是什么，能不能只用两次握手来达到同样的目的。

1. 第一次握手：客户端发送网络包，服务端收到了。这样服务端就能得出结论：客户端的发送能力、服务端的接收能力是正常的。
2. 第二次握手：服务端发包，客户端收到了。这样客户端就能得出结论：服务端的接收、发送能力，客户端的接收、发送能力是正常的。不过此时服务器并不能确认客户端的接收能力是否正常。
3. 第三次握手：客户端发包，服务端收到了。这样服务端就能得出结论：客户端的接收、发送能力正常，服务器自己的发送、接收能力也正常。因此，需要三次握手才能确认双方的接收与发送能力是否正常。

#### 4.1.2.为什么不需要四次握手？

有人可能会说 A 发出第三次握手的信息后在没有接收到 B 的请求就已经进入了连接状态，那如果 A 的这个确认包丢失或者滞留了怎么办？我们需要明白一点，完全可靠的通信协议是不存在的。在经过三次握手之后，客户端和服务端已经可以确认之前的通信状况，都收到了确认信息。所以即便再增加握手次数也不能保证后面的通信完全可靠，所以是没有必要的。

#### 4.1.3.什么是半连接队列？

服务器第一次收到客户端的 SYN 之后，就会处于 SYN_RCVD 状态，此时双方还没有完全建立其连接，服务器会把此种状态下请求连接放在一个队列里，我们把这种队列称之为半连接队列。当然还有一个全连接队列，就是已经完成三次握手，建立起连接的就会放在全连接队列中。如果队列满了就有可能会出现丢包现象。这里在补充一点关于SYN-ACK 重传次数的问题：服务器发送完SYN-ACK包，如果未收到客户确认包，服务器进行首次重传，等待一段时间仍未收到客户确认包，进行第二次重传。如果重传次数超过系统规定的最大重传次数，系统将该连接信息从半连接队列中删除。注意，每次重传等待的时间不一定相同，一般会是指数增长，例如间隔时间为 1s，2s，4s，8s…

#### 4.1.4.ISN(Initial Sequence Number)是固定的吗？

当一端为建立连接而发送它的SYN时，它为连接选择一个初始序号。ISN随时间而变化，因此每个连接都将具有不同的ISN。ISN可以看作是一个32比特的计数器，每4ms加1 。这样选择序号的目的在于防止在网络中被延迟的分组在以后又被传送，而导致某个连接的一方对它做错误的解释。
三次握手的其中一个重要功能是客户端和服务端交换 ISN(Initial Sequence Number)，以便让对方知道接下来接收数据的时候如何按序列号组装数据。如果 ISN 是固定的，攻击者很容易猜出后续的确认号，因此 ISN 是动态生成的。

### 2.四次挥手

建立一个连接需要三次握手，而终止一个连接要经过四次挥手（也有将四次挥手叫做四次握手的）。这由TCP的半关闭（half-close）造成的。
所谓的半关闭，其实就是TCP提供了连接的一端在结束它的发送后还能接收来自另一端数据的能力。TCP 的连接的拆除需要发送四个包，因此称为四次挥手(Four-way handshake)，客户端或服务器均可主动发起挥手动作。
刚开始双方都处于 ESTABLISHED 状态，假如是客户端先发起关闭请求。

四次挥手的过程如下：

1. 第一次挥手：客户端发送一个 FIN 报文，报文中会指定一个序列号。此时客户端处于 FIN_WAIT1 状态。即发出连接释放报文段（FIN=1，序号seq=u），并停止再发送数据，主动关闭TCP连接，进入FIN_WAIT1（终止等待1）状态，等待服务端的确认。
2. 第二次挥手：服务端收到 FIN 之后，会发送 ACK 报文，且把客户端的序列号值 +1 作为 ACK 报文的序列号值，表明已经收到客户端的报文了，此时服务端处于 CLOSE_WAIT 状态。即服务端收到连接释放报文段后即发出确认报文段（ACK=1，确认号ack=u+1，序号seq=v），服务端进入CLOSE_WAIT（关闭等待）状态，此时的TCP处于半关闭状态，客户端到服务端的连接释放。客户端收到服务端的确认后，进入FIN_WAIT2（终止等待2）状态，等待服务端发出的连接释放报文段。
3. 第三次挥手：如果服务端也想断开连接了，和客户端的第一次挥手一样，发给 FIN 报文，且指定一个序列号。此时服务端处于 LAST_ACK 的状态。即服务端没有要向客户端发出的数据，服务端发出连接释放报文段（FIN=1，ACK=1，序号seq=w，确认号ack=u+1），服务端进入LAST_ACK（最后确认）状态，等待客户端的确认。
4. 第四次挥手：客户端收到 FIN 之后，一样发送一个 ACK 报文作为应答，且把服务端的序列号值 +1 作为自己 ACK 报文的序列号值，此时客户端处于 TIME_WAIT 状态。需要过一阵子以确保服务端收到自己的 ACK 报文之后才会进入 CLOSED 状态，服务端收到 ACK 报文之后，就处于关闭连接了，处于 CLOSED 状态。即客户端收到服务端的连接释放报文段后，对此发出确认报文段（ACK=1，seq=u+1，ack=w+1），客户端进入TIME_WAIT（时间等待）状态。
此时TCP未释放掉，需要经过时间等待计时器设置的时间2MSL后，客户端才进入CLOSED状态。收到一个FIN只意味着在这一方向上没有数据流动。
客户端执行主动关闭并进入TIME_WAIT是正常的，服务端通常执行被动关闭，不会进入TIME_WAIT状态。
在socket编程中，任何一方执行close()操作即可产生挥手操作。![a2306c52ac3452de325ab777de24b065.jpeg](evernotecid://CD9A3445-D0EE-4B52-B7C1-F60AB57CBA78/appyinxiangcom/27150514/ENResource/p71)

#### 4.2.1.挥手为什么需要四次？

因为当服务端收到客户端的SYN连接请求报文后，可以直接发送SYN+ACK报文。其中ACK报文是用来应答的，SYN报文是用来同步的。但是关闭连接时，当服务端收到FIN报文时，很可能并不会立即关闭SOCKET，所以只能先回复一个ACK报文，告诉客户端，“你发的FIN报文我收到了”。只有等到我服务端所有的报文都发送完了，我才能发送FIN报文，因此不能一起发送。故需要四次挥手。

#### 4.2.2.四次挥手释放连接时，等待2MSL的意义?

MSL是Maximum Segment Lifetime的英文缩写，可译为“最长报文段寿命”，它是任何报文在网络上存在的最长时间，超过这个时间报文将被丢弃。为了保证客户端发送的最后一个ACK报文段能够到达服务器。因为这个ACK有可能丢失，从而导致处在LAST-ACK状态的服务器收不到对FIN-ACK的确认报文。服务器会超时重传这个FIN-ACK，接着客户端再重传一次确认，重新启动时间等待计时器。最后客户端和服务器都能正常的关闭。
假设客户端不等待2MSL，而是在发送完ACK之后直接释放关闭，一但这个ACK丢失的话，服务器就无法正常的进入关闭连接状态。RFC 793中规定MSL为2分钟，实际应用中常用的是30秒，1分钟和2分钟等

两个理由：

1. 保证客户端发送的最后一个ACK报文段能够到达服务端。
这个ACK报文段有可能丢失，使得处于LAST-ACK状态的B收不到对已发送的FIN+ACK报文段的确认，服务端超时重传FIN+ACK报文段，而客户端能在2MSL时间内收到这个重传的FIN+ACK报文段，接着客户端重传一次确认，重新启动2MSL计时器，最后客户端和服务端都进入到CLOSED状态，若客户端在TIME-WAIT状态不等待一段时间，而是发送完ACK报文段后立即释放连接，则无法收到服务端重传的FIN+ACK报文段，所以不会再发送一次确认报文段，则服务端无法正常进入到CLOSED状态。
2. 防止“已失效的连接请求报文段”出现在本连接中。
客户端在发送完最后一个ACK报文段后，再经过2MSL，就可以使本连接持续的时间内所产生的所有报文段都从网络中消失，使下一个新的连接中不会出现这种旧的连接请求报文段。

#### 4.2.3.为什么TIME_WAIT状态需要经过2MSL才能返回到CLOSE状态？

理论上，四个报文都发送完毕，就可以直接进入CLOSE状态了，但是可能网络是不可靠的，有可能最后一个ACK丢失。所以TIME_WAIT状态就是用来重发可能丢失的ACK报文。

## 5、TCP报文

![a7938e0527152b7c945c66de42f9d16a.gif](evernotecid://CD9A3445-D0EE-4B52-B7C1-F60AB57CBA78/appyinxiangcom/27150514/ENResource/p72)

1. 端口号：用来标识同一台计算机的不同的应用进程。
  1）源端口：源端口和IP地址的作用是标识报文的返回地址。
  2）目的端口：端口指明接收方计算机上的应用程序接口。TCP报头中的源端口号和目的端口号同IP数据报中的源IP与目的IP唯一确定一条TCP连接。
2. 序号和确认号：是TCP可靠传输的关键部分。序号是本报文段发送的数据组的第一个字节的序号。在TCP传送的流中，每一个字节一个序号。e.g.一个报文段的序号为300，此报文段数据部分共有100字节，则下一个报文段的序号为400。所以序号确保了TCP传输的有序性。确认号，即ACK，指明下一个期待收到的字节序号，表明该序号之前的所有数据已经正确无误的收到。确认号只有当ACK标志为1时才有效。比如建立连接时，SYN报文的ACK标志位为0。
3. 数据偏移／首部长度：4bits。由于首部可能含有可选项内容，因此TCP报头的长度是不确定的，报头不包含任何任选字段则长度为20字节，4位首部长度字段所能表示的最大值为1111，转化为10进制为15，15*32/8 = 60，故报头最大长度为60字节。首部长度也叫数据偏移，是因为首部长度实际上指示了数据区在报文段中的起始偏移值。
4. 保留：为将来定义新的用途保留，现在一般置0。
5. 控制位：URG ACK PSH RST SYN FIN，共6个，每一个标志位表示一个控制功能。
  1）URG：紧急指针标志，为1时表示紧急指针有效，为0则忽略紧急指针。
  2）ACK：确认序号标志，为1时表示确认号有效，为0表示报文中不含确认信息，忽略确认号字段。
  3）PSH：push标志，为1表示是带有push标志的数据，指示接收方在接收到该报文段以后，应尽快将这个报文段交给应用程序，而不是在缓冲区排队。
  4）RST：重置连接标志，用于重置由于主机崩溃或其他原因而出现错误的连接。或者用于拒绝非法的报文段和拒绝连接请求。
  5）SYN：同步序号，用于建立连接过程，在连接请求中，SYN=1和ACK=0表示该数据段没有使用捎带的确认域，而连接应答捎带一个确认，即SYN=1和ACK=1。
  6）FIN：finish标志，用于释放连接，为1时表示发送方已经没有数据发送了，即关闭本方数据流。
6. 窗口：滑动窗口大小，用来告知发送端接受端的缓存大小，以此控制发送端发送数据的速率，从而达到流量控制。窗口大小时一个16bit字段，因而窗口大小最大为65535。
7. 校验和：奇偶校验，此校验和是对整个的&nbsp;TCP&nbsp;报文段，包括&nbsp;TCP&nbsp;头部和&nbsp;TCP&nbsp;数据，以&nbsp;16&nbsp;位字进行计算所得。由发送端计算和存储，并由接收端进行验证。
8. 紧急指针：只有当&nbsp;URG&nbsp;标志置&nbsp;1&nbsp;时紧急指针才有效。紧急指针是一个正的偏移量，和顺序号字段中的值相加表示紧急数据最后一个字节的序号。&nbsp;TCP&nbsp;的紧急方式是发送端向另一端发送紧急数据的一种方式。
9. 选项和填充：最常见的可选字段是最长报文大小，又称为MSS（Maximum Segment Size），每个连接方通常都在通信的第一个报文段（为建立连接而设置SYN标志为1的那个段）中指明这个选项，它表示本端所能接受的最大报文段的长度。选项长度不一定是32位的整数倍，所以要加填充位，即在这个字段中加入额外的零，以保证TCP头是32的整数倍。10、数据部分：&nbsp;TCP&nbsp;报文段中的数据部分是可选的。在一个连接建立和一个连接终止时，双方交换的报文段仅有&nbsp;TCP&nbsp;首部。如果一方没有数据要发送，也使用没有任何数据的首部来确认收到的数据。在处理超时的许多情况中，也会发送不带任何数据的报文段。

## 6、UDP报文

![3a14d9a458e0b7cdeddfcd48c6bb80d5.jpeg](evernotecid://CD9A3445-D0EE-4B52-B7C1-F60AB57CBA78/appyinxiangcom/27150514/ENResource/p73)
UDP协议分为首部字段和数据字段，其中首部字段只占用8个字节，分别是个占用两个字节的源端口、目的端口、长度和检验和。

## 7、IP包

IP数据包是一种可变长分组，它由首部和数据负载两部分组成。首部长度一般为20-60字节（Byte），其中后40字节是可选的，长度不固定，前20字节格式为固定。数据负载部分的长度一般可变，整个IP数据包的最大长度为65535B。

![193855f4bef0fb2399d17a9c8087e363.jpeg](evernotecid://CD9A3445-D0EE-4B52-B7C1-F60AB57CBA78/appyinxiangcom/27150514/ENResource/p74)

1. 版本号（Version）：长度为4位（bit），IPv4的值为0100，IPv6的值为0110。
2. 首部长度：指的是IP包头长度，用4位（bit）表示，十进制值就是[0,15]，一个IP包前20个字节是必有的，后40个字节根据情况可能有可能没有。如果IP包头是20个字节，则该位应是20/4=5。
3. 服务类型（Type of Service&nbsp; TOS）：长度为8位（bit），其组成：前3位为优先级（Precedence），后4位标志位，最后1位保留未用。优先级主要用于QoS，表示从0（普通级别）到7（网络控制分组）的优先级。标志位可分别表示D（Delay更低的时延）、T（Throughput 更高的吞吐量）、R（Reliability更高的可靠性）、C（Cost 更低费用的路由）。TOS只表示用户的请求，不具有强制性，实际应用中很少用，路由器通常忽略TOS字段。
4. 总长度（Total Length）：指IP包总长度，用16位（bit）表示，即IP包最大长度可以达216=65535字节。在以太网中允许的最大包长为1500B，当超过网络允许的最大长度时需将过长的数据包分片。
5. 标识符（Identifier）：长度为16位，用于数据包在分段重组时标识其序列号。将数据分段后，打包成IP包，IP包因走的路由上不同，会产生不同的到达目地的时间，到达目地的后再根据标识符进行重新组装还原。该字段要与标志、段偏移一起使用的才能达到分段组装的目标。
6. 标志（Flags）：长度为3位，三位从左到右分别是MF、DF、未用。MF=1表示后面还有分段的数据包，MF=0表示没有更多分片（即最后一个分片）。DF=1表示路由器不能对该数据包分段，DF=0表示数据包可以被分段。
7. 偏移量（Fragment Offset）：也称段偏移，用于标识该数据段在上层初始数据报文中的偏移量。如果某个包含分段的上层报文的IP数据包在传送时丢失，则整个一系列包含分段的上层数据包的IP包都会要求重传。
8. 生存时间（TTL）：长度为8位，初始值由操作系统设置，每经过一个路由器转发后其值就减1，减至0后丢弃该包。这种机制可以避免数据包找不到目地时不断被转发，堵塞网络。
9. 协议（Protocol）：长度为8位，标识上层所使用的协议。
10. 首部校验和（Header Checksum）：长度为16位，首部检验和只对IP数据包首部进行校验，不包含数据部分。数据包每经过一个中间节点都要重新计算首部校验和，对首都进行检验。
11. 源IP地址（Source IP）：长度为32位，表示数据发送的主机IP。
12. 目的IP地址（Destination IP）：长度为32位，表示数据要接收的主机IP。
13. 选项字段（Options）：长度为0-40字节（Byte），主要有：安全和处理限制（Security）、记录路径（Record Route）、时间戳（Timestamps）、宽松源站选路（Loose Source Routing）、严格的源站选路（Strict Source Routing）等。

## 8、URL过程和DNS

### 1.在浏览器中输入 URL 地址到显示主页的过程？

1. DNS 解析：浏览器查询 DNS，获取域名对应的 IP 地址：具体过程包括浏览器搜索自身的 DNS 缓存、搜索操作系统的 DNS 缓存、读取本地的 Host 文件和向本地 DNS 服务器进行查询等。对于向本地 DNS 服务器进行查询，如果要查询的域名包含在本地配置区域资源中，则返回解析结果给客户机，完成域名解析(此解析具有权威性)；如果要查询的域名不由本地 DNS 服务器区域解析，但该服务器已缓存了此网址映射关系，则调用这个 IP 地址映射，完成域名解析（此解析不具有权威性）。如果本地域名服务器并未缓存该网址映射关系，那么将根据其设置发起递归查询或者迭代查询；
2. TCP 连接：浏览器获得域名对应的 IP 地址以后，浏览器向服务器请求建立链接，发起三次握手；
3. 发送 HTTP 请求：TCP 连接建立起来后，浏览器向服务器发送 HTTP 请求；
4. 服务器处理请求并返回 HTTP 报文：服务器接收到这个请求，并根据路径参数映射到特定的请求处理器进行处理，并将处理结果及相应的视图返回给浏览器；
5. 浏览器解析渲染页面：浏览器解析并渲染视图，若遇到对 js 文件、css 文件及图片等静态资源的引用，则重复上述步骤并向服务器请求这些资源；浏览器根据其请求到的资源、数据渲染页面，最终向用户呈现一个完整的页面。
6. 连接结束。

### 2.DNS 的解析过程？

1. 主机向本地域名服务器的查询一般都是采用递归查询。
所谓递归查询就是：如果主机所询问的本地域名服务器不知道被查询的域名的 IP 地址，那么本地域名服务器就以 DNS 客户的身份，向根域名服务器继续发出查询请求报文(即替主机继续查询)，而不是让主机自己进行下一步查询。因此，递归查询返回的查询结果或者是所要查询的 IP 地址，或者是报错，表示无法查询到所需的 IP 地址。
2. 本地域名服务器向根域名服务器的查询的迭代查询。
迭代查询的特点：当根域名服务器收到本地域名服务器发出的迭代查询请求报文时，要么给出所要查询的 IP 地址，要么告诉本地服务器：“你下一步应当向哪一个域名服务器进行查询”。然后让本地服务器进行后续的查询。根域名服务器通常是把自己知道的顶级域名服务器的 IP 地址告诉本地域名服务器，让本地域名服务器再向顶级域名服务器查询。顶级域名服务器在收到本地域名服务器的查询请求后，要么给出所要查询的 IP 地址，要么告诉本地服务器下一步应当向哪一个权限域名服务器进行查询。最后，本地域名服务器得到了所要解析的 IP 地址或报错，然后把这个结果返回给发起查询的主机。通过主机名，最终得到该主机名对应的IP地址的过程叫做域名解析（或主机名解析）。
DNS协议运行在UDP协议之上，使用端口号53。

#### 8.2.1.DNS域名系统工作原理

1. 查询 浏览器、操作系统 缓存。
2. 请求 本地域名服务器
3. 本地域名服务器未命中缓存，其请求 根域名服务器。
4. 根域名服务器返回所查询域的主域名服务器。（主域名、顶级域名，如com、cn）
5. 本地域名服务器请求主域名服务器，获取该域名的 名称服务器（域名注册商的服务器）。
6. 本地域名服务器向 名称服务器 请求 域名-IP 映射。
7. 缓存解析结果

#### 8.2.2.为什么域名解析用UDP协议?

因为UDP快啊！UDP的DNS协议只要一个请求、一个应答就好了。
而使用基于TCP的DNS协议要三次握手、发送数据以及应答、四次挥手。但是UDP协议传输内容不能超过512字节。不过客户端向DNS服务器查询域名，一般返回的内容都不超过512字节，用UDP传输即可。

#### 8.2.3.为什么区域传送用TCP协议？

因为TCP协议可靠性好啊！你要从主DNS上复制内容啊，你用不可靠的UDP？ 因为TCP协议传输的内容大啊，你用最大只能传512字节的UDP协议？万一同步的数据大于512字节，你怎么办？

#### 8.2.4.域名的结构讲清楚！

&nbsp;www.tmall.com对应的真正的域名为www.tmall.com.。
末尾的.称为根域名，因为每个域名都有根域名，因此我们通常省略。
根域名的下一级，叫做"顶级域名"，比如.com、.net；
再下一级叫做"次级域名"，比如www.tmall.com里面的.tmall，这一级域名是用户可以注册的；
再下一级是主机名（host），比如www.tmall.com里面的www，又称为"三级域名"，这是用户在自己的域里面为服务器分配的名称，是用户可以任意分配的。

#### 8.2.5.DNS怎么做域名解析的？

(1)先在本机的DNS里头查，如果有就直接返回了。
(2)本机DNS里头发现没有，就去根服务器里查。根服务器发现这个域名是属于com域，因此根域DNS服务器会返回它所管理的com域中的DNS 服务器的IP地址，意思是“虽然我不知道你要查的那个域名的地址，但你可以去com域问问看”(3)本机的DNS接到又会向com域的DNS服务器发送查询消息。com&nbsp;域中也没有www.tmall.com这个域名的信息，和刚才一样，com域服务器会返回它下面的tmall.com域的DNS服务器的IP地址。

#### 8.2.6.CNAME有什么好处？

也就是别名方便cdn配置？CDN的定义：CDN就是采用更多的缓存服务器（CDN边缘节点），布放在用户访问相对集中的地区或网络中。当用户访问网站时，利用全局负载技术，将用户的访问指向距离最近的缓存服务器上，由缓存服务器响应用户请求。

### 3.谈谈你对域名缓存的了解？

为了提高 DNS 查询效率，并减轻服务器的负荷和减少因特网上的 DNS 查询报文数量，在域名服务器中广泛使用了高速缓存，用来存放最近查询过的域名以及从何处获得域名映射信息的记录。
由于名字到地址的绑定并不经常改变，为保持高速缓存中的内容正确，域名服务器应为每项内容设置计时器并处理超过合理时间的项（例如：每个项目两天）。
当域名服务器已从缓存中删去某项信息后又被请求查询该项信息，就必须重新到授权管理该项的域名服务器绑定信息。
当权限服务器回答一个查询请求时，在响应中都指明绑定有效存在的时间值。增加此时间值可减少网络开销，而减少此时间值可提高域名解析的正确性。不仅在本地域名服务器中需要高速缓存，在主机中也需要。许多主机在启动时从本地服务器下载名字和地址的全部数据库，维护存放自己最近使用的域名的高速缓存，并且只在从缓存中找不到名字时才使用域名服务器。维护本地域名服务器数据库的主机应当定期地检查域名服务器以获取新的映射信息，而且主机必须从缓存中删除无效的项。由于域名改动并不频繁，大多数网点不需花精力就能维护数据库的一致性。

## 9、其他的知识点

### 1.Socket

Socket是应用层与TCP/IP协议族通信的中间软件抽象层，它是一组接口。
在设计模式中，Socket其实就是一个门面模式，它把复杂的TCP /IP协议族隐藏在Socket接口后面，对用户来说，一组简单的接口就是全部，让Socket去组织数据，以符合指定的协议。TCP/IP是传输层协议，主要解决数据如何在网络中传输；而Http是应用层协议，主要解决如何包装数据。Socket是对TCP/IP协议的封装，Socket本身并不是协议，而是一个调用接口（API），通过Socket，我们才能使用TCP/IP协议。
socket_create:第一个参数”AF_INET”用来指定域名;第二个参数”SOCK_STREAM”告诉函数将创建一个什么类型的Socket(在这个例子中是TCP类型)

### 2.GET 和 POST 的区别？

本质区别：
GET 只是一次 HTTP请求，POST 先发请求头再发请求体，实际上是两次请求。
GET 一般用来从服务器上获取资源，POST 一般用来更新服务器上的资源；
GET 是幂等的，即读取同一个资源，总是得到相同的数据，而 POST 不是幂等的，因为每次请求对资源的改变并不是相同的；

请求参数形式上看，
GET 请求的数据会附在 URL 之后，get请求长度最多1024kb，post最大是2M，可以在php.ini中修改post_max_size;而 POST 请求会把提交的数据则放置在是 HTTP 请求报文的请求体中；
POST 的安全性要比 GET 的安全性高。

### 3.单工通信、半双工通信、全双工通信

单工通信：通信的信道是单向的，发送端和接收端是固定的，发送端只能发送送不能接收，接收端只能接收不能发送。
半双工：可以实现双向的通信，但不能在两个方向上同时进行，必须轮流交替进行。
全双工通信：允许数据同时在两个方向上传输，即通信双方可以同时发送和接收数据。

### 4.forward 和 redirect 的区别？

Forward 和 Redirect 代表了两种请求转发方式：直接转发和间接转发。
直接转发方式（Forward）：客户端和浏览器只发出一次请求，Servlet、HTML、JSP 或其它信息资源，由第二个信息资源响应该请求，在请求对象 request 中，保存的对象对于每个信息资源是共享的。
间接转发方式（Redirect）：实际是两次 HTTP 请求，服务器端在响应第一次请求的时候，让浏览器再向另外一个 URL 发出请求，从而达到转发的目的。
举个通俗的例子：直接转发就相当于：“A 找 B 借钱，B 说没有，B 去找 C 借，借到借不到都会把消息传递给 A”；间接转发就相当于："A 找 B 借钱，B 说没有，让 A 去找 C 借"。

### 5.什么是数字签名？

为了避免数据在传输过程中被替换，比如黑客修改了你的报文内容，但是你并不知道，所以我们让发送端做一个数字签名，把数据的摘要消息进行一个加密，比如 MD5，得到一个签名，和数据一起发送。然后接收端把数据摘要进行 MD5 加密，如果和签名一样，则说明数据确实是真的。

### 6.什么是数字证书？

对称加密中，双方使用公钥进行解密。虽然数字签名可以保证数据不被替换，但是数据是由公钥加密的，如果公钥也被替换，则仍然可以伪造数据，因为用户不知道对方提供的公钥其实是假的。所以为了保证发送方的公钥是真的，CA 证书机构会负责颁发一个证书，里面的公钥保证是真的，用户请求服务器时，服务器将证书发给用户，这个证书是经由系统内置证书的备案的。

### 7.进程和线程

进程是具有一定功能的程序关于某个数据集合上的一次运行活动，进程是系统进行资源调度和分配的一个最小单位。线程是进程的实体，是CPU调度和分派的基本单位，它是比进程更小的能独立运行的基本单位。一个进程可以有多个线程（至少一个），多个线程也可以并发执行。
进程和线程的区别包括：

1. 一个程序至少有一个进程，一个进程至少有一个线程
2. 线程的划分尺度小于进程，使得多线程程序的并发性高
3. 进程在执行过程中拥有独立的内存单元，而多个线程共享内存，从而极大地提高了程序的运行效率
4. 每个独立的线程有一个程序运行的入口、顺序执行序列和程序的出口。但是线程不能够独立执行，必须依存在应用程序中，由应用程序提供多个线程执行控制
5. 多线程的意义在于一个应用程序中，有多个执行部分可以同时执行。但操作系统并没有将多个线程看做多个独立的应用，来实现进程的调度和管理以及资源分配
