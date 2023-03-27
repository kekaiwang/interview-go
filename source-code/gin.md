# Gin

## 一个 HTTP 请求的生命周期

### run 一个 server

Gin 是一个基于 Go 语言的 Web 框架，采用了类似于 Martini 的 API 设计，可以快速地构建 Web 应用程序。在 Gin 中，一个 HTTP 请求的处理流程一般会经过中间件的处理和路由的匹配，最终会被具体的处理函数所处理。

下面我们来结合源码，分析一下 Gin 中一个 GET 请求是如何执行的。

首先，我们需要构建一个最基本的 Gin Web 服务：

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        c.String(200, "Hello, Gin!")
    })

    r.Run(":8080")
}

```

上述代码中，我们使用了 Gin 的默认路由引擎，并在该引擎中注册了一个 GET 请求的路由处理函数，该函数会在收到一个 GET 请求时被执行，返回一个字符串 "Hello, Gin!"。最后，我们使用 r.Run(":8080") 启动了一个 HTTP 服务，监听 8080 端口。

接下来，我们来分析一下该 GET 请求的执行流程：

1. 启动服务时，调用 `gin.Default()` 方法会创建一个默认的 `Engine` 实例，并注册了默认的 Logger、Recovery 中间件和路由处理函数。

```go
func Default() *Engine {
    debugPrintWARNINGDefault()
    engine := New()
    engine.Use(Logger(), Recovery())
    engine.HandleMethodNotAllowed = true
    engine.NoMethod(func(c *Context) {
        c.AbortWithStatus(http.StatusMethodNotAllowed)
    })
    engine.NoRoute(func(c *Context) {
        c.AbortWithStatus(http.StatusNotFound)
    })
    return engine
}

```

1. 在注册路由时，调用 `r.GET("/", ...)`  方法会将该路由信息加入到 Engine 的路由表中。其中，路由信息会被保存在 Engine 的路由数组中，每个路由信息会包含请求的 HTTP 方法、请求的路径和该路由对应的处理函数等信息。

```go
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
    return group.handle(http.MethodGet, relativePath, handlers)
}

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
    absolutePath := group.calculateAbsolutePath(relativePath)
    handlers = group.combineHandlers(handlers)
    group.engine.addRoute(httpMethod, absolutePath, handlers)
    return group.returnObj()
}

func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
    assert1(path[0] == '/', "path must begin with '/'")
    assert1(method != "", "HTTP method can not be empty")
    assert1(len(handlers) > 0, "there must be at least one handler")
    assert1(IsValidMethod(method), "invalid HTTP method")

    debugPrintRoute(method, path, handlers)

    if engine.RouterGroup == nil {
        engine.RouterGroup = &RouterGroup{
            Handlers: handlers,
            engine:   engine,
        }
    }
    engine.trees.addRoute(method, path, handlers)
}
```

当路由匹配完成后，**`gin`** 会将该请求交给对应的处理函数进行处理。下面是一个处理函数的示例代码：

```go
func handler(c *gin.Context) {
    name := c.Param("name")
    c.String(http.StatusOK, "Hello %s", name)
}
```

在这个示例代码中，**`handler`** 函数接受一个 **`gin.Context`** 类型的参数 **`c`**，并从该参数中获取了一个名为 **`name`** 的路由参数。随后，该函数使用 **`gin.Context`** 中提供的 **`String`** 方法返回了一个 **`HTTP 200`** 的响应。

当请求到达该处理函数时，**`gin`** 会创建一个新的 **`gin.Context`** 对象，并将该请求的相关信息（例如请求头、请求体、请求参数等）存储到该对象中，然后将该对象传递给对应的处理函数进行处理。

下面是 **`gin.Context`** 中存储请求信息的部分代码：

```go
type Context struct {
    // ...
    // Request related fields
    request *http.Request
    writer  ResponseWriter
    params  Params
    // ...
}
```

在该结构体中，**`request`** 字段保存了 HTTP 请求的相关信息，例如请求头、请求体等。**`writer`** 字段保存了 HTTP 响应的相关信息，例如响应头、响应体等。**`params`** 字段保存了该请求的路由参数。

在处理函数中，可以通过调用 **`gin.Context`** 中的方法获取请求信息，例如：

- **`c.Param(name string)`**：获取指定名称的路由参数
- **`c.Query(name string)`**：获取指定名称的 URL 查询参数
- **`c.PostForm(name string)`**：获取指定名称的 POST 表单参数
- **`c.Request.Header.Get(name string)`**：获取指定名称的请求头
- **`c.Request.Body`**：获取请求体

当处理函数处理完请求后，需要将响应信息写入 **`gin.Context`** 对象，并调用 **`gin.Context`** 中的 **`Write`** 方法将响应发送给客户端。下面是一个示例代码：

```go
func handler(c *gin.Context) {
    name := c.Param("name")
    c.String(http.StatusOK, "Hello %s", name)
}
```

在这个示例代码中，**`handler`** 函数将 **`Hello {name}`** 格式的字符串作为响应体，并将响应状态码设为 **`HTTP 200`**。随后，该函数调用了 **`gin.Context`** 中的 **`String`** 方法，将响应信息写入 **`gin.Context`** 对象中，并调用了 **`gin.Context`** 中的 **`Write`** 方法将响应发送给客户端。

至此，一个 **`GET`** 请求在 **`gin`** 框架中的处理流程就结束了。在该流程中，**`gin`** 通过路由匹配将请求交给对应的处理函数进行处理，并将请求信息存储到 **`gin.Context`** 对象中，然后调用处理函数处理请求，并将响应信息写入 **`gin.Context`** 对象，并通过 `Write

**`ServeHTTP`** 是 Go 语言中的一个接口，它定义了一个可以处理 HTTP 请求的方法。在 Gin 框架中，每个路由处理函数都会被包装成一个实现了 **`http.Handler`** 接口的函数，该函数实现了 **`ServeHTTP`** 方法。

```go
// ServeHTTP conforms to the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    c := engine.pool.Get().(*Context)
    c.writermem.reset(w)
    c.Request = req
    c.reset()

    engine.handleHTTPRequest(c)

    engine.pool.Put(c)
}
```

当有一个 HTTP 请求到达时，Gin 框架会首先根据请求路径和方法匹配对应的路由处理函数，然后将该函数包装成 **`http.Handler`** 接口的函数，并将该函数作为参数传递给 Go 标准库中的 **`http.Server`** 结构体的 **`Handler`** 字段。当有新的请求到达时，**`http.Server`** 就会调用该 **`Handler`** 函数进行处理。

在 Gin 中，每个路由处理函数的实现都可以被看作是一个 **`http.Handler`** 接口的实现。当有一个新的请求到达时，Gin 会调用该处理函数的 **`ServeHTTP`** 方法进行处理，同时将该请求的上下文传递给该函数。该函数通过解析请求上下文，获取请求的参数、请求体、请求头等信息，并根据这些信息来生成响应结果。最终，该函数会将生成的响应结果写回到响应体中，并通过 **`http.ResponseWriter`** 接口的方法将其返回给客户端。

在 Gin 中，**`ServeHTTP`** 方法的实现是由 **`Context`** 结构体中的 **`handleHTTPRequest`** 方法实现的。该方法通过解析请求上下文，获取请求的参数、请求体、请求头等信息，并根据这些信息来生成响应结果。最终，该方法会将生成的响应结果写回到响应体中，并通过 **`http.ResponseWriter`** 接口的方法将其返回给客户端。同时，在处理请求时，该方法还会根据路由处理函数的定义，将请求参数绑定到函数的参数上，并将函数的返回值写入到响应体中。

```go
func (engine *Engine) handleHTTPRequest(c *Context) {
    httpMethod := c.Request.Method
    rPath := c.Request.URL.Path
    unescape := false
    if engine.UseRawPath && len(c.Request.URL.RawPath) > 0 {
        rPath = c.Request.URL.RawPath
        unescape = engine.UnescapePathValues
    }

    if engine.RemoveExtraSlash {
        rPath = cleanPath(rPath)
    }

    // Find root of the tree for the given HTTP method
    t := engine.trees
    for i, tl := 0, len(t); i < tl; i++ {
        if t[i].method != httpMethod {
            continue
        }
        root := t[i].root
        // Find route in tree
        value := root.getValue(rPath, c.params, c.skippedNodes, unescape)
        if value.params != nil {
            c.Params = *value.params
        }
        if value.handlers != nil {
            c.handlers = value.handlers
            c.fullPath = value.fullPath
            c.Next()
            c.writermem.WriteHeaderNow()
            return
        }
        if httpMethod != http.MethodConnect && rPath != "/" {
            if value.tsr && engine.RedirectTrailingSlash {
                redirectTrailingSlash(c)
                return
            }
            if engine.RedirectFixedPath && redirectFixedPath(c, root, engine.RedirectFixedPath) {
                return
            }
        }
        break
    }
    ...
}
```

在 Gin 中，**`engine.pool`** 存储了一个可用的请求对象池，这些请求对象在请求处理过程中被多次使用。当有新的请求到来时，Gin 会从池中获取一个请求对象并进行初始化，然后将其分配给处理该请求的 Goroutine。当请求处理完成后，Gin 会将该请求对象重置并放回到池中，以供下一个请求使用。

在 **`engine.handleHTTPRequest`** 方法中，首先从 **`engine.pool`** 中获取一个请求对象，然后进行初始化并处理 HTTP 请求。处理完请求后，调用 **`c.reset()`** 方法重置该请求的上下文对象，并将请求对象放回到 **`engine.pool`** 中。如果请求处理过程中发生了异常，Gin 会记录错误并将请求对象放回到 **`engine.pool`** 中，以防止请求对象被泄漏。

在 **`ServeHTTP`** 方法中，首先通过 **`engine.pool.Get()`** 方法从 **`engine`** 实例的 **`pool`** 中获取一个 **`Context`** 对象。**`pool`** 中存储的 **`Context`** 对象可以被重复利用，避免了频繁的内存分配，提高了性能。

获取 **`Context`** 对象后，将请求信息 **`c.Request`** 和响应信息 **`c.Writer`** 传递给 **`engine.handleHTTPRequest`** 方法处理。**`handleHTTPRequest`** 方法会根据请求路径和请求方法找到对应的路由处理器（handler），并将其保存在 **`c.handlers`** 切片中。

**接下来，`handleHTTPRequest` 方法会执行 `c.Next()` 方法，这个方法用来执行下一个中间件或路由处理器。**默认情况下，**`c.handlers`** 中保存的是一个或多个路由处理器。在执行路由处理器之前，**`c.Next()`** 方法会依次执行中间件函数，这些中间件函数的执行顺序由它们在 **`RouterGroup`** 中的注册顺序决定。

在执行路由处理器时，**`c.Next()`** 方法不会被调用，因为这是最后一个处理器了。在路由处理器执行完成后，**`engine.pool.Put(c)`** 方法会将 **`Context`** 对象归还到 **`pool`** 中，以供下一次使用。

总结一下，**`ServeHTTP`** 方法主要完成以下几个步骤：

1. 从 **`engine`** 实例的 **`pool`** 中获取 **`Context`** 对象；
2. 将请求信息 **`c.Request`** 和响应信息 **`c.Writer`** 传递给 **`engine.handleHTTPRequest`** 方法处理；
3. 在 **`handleHTTPRequest`** 方法中执行路由处理器，并依次执行中间件函数；
4. 在路由处理器执行完成后，将 **`Context`** 对象归还到 **`pool`** 中。

## context

### 基础介绍

Gin 框架的 Context 对象是在每次请求中创建的，用于处理请求并向客户端响应结果。在这个对象中，包含了请求的信息（例如 URL、HTTP 方法、Header、Body 等）、响应的信息（例如 HTTP 状态码、Header、Body 等）、一些便捷方法（例如解析 JSON 请求体、获取请求参数、设置 Cookie 等）以及一些中间件方法（例如设置请求时间、设置跨域 Header 等）。

下面我们来结合源码解析 Gin 框架中的 Context 对象。

Context 的定义

首先看一下 Context 对象的定义，在 gin.Context 结构体中，包含了各种请求信息和响应信息的字段，如下所示：

```
goCopy code
type Context struct {
    // 请求相关的信息
    Writer     ResponseWriter
    Request    *http.Request
    Params     Params
    ...

    // 响应相关的信息
    status     int
    ...

    // 一些方法
    Get        func(key string) (value interface{})
    Set        func(key string, value interface{})
    ...

    // 其他字段
    keys       map[string]interface{}
    errors     errorMsgs
    Handlers   HandlersChain
    ...

    // sync.Mutex，保证 Context 对象安全并发使用
    mu         sync.RWMutex
}

```

Context 的方法

接下来我们来看一下 Context 对象中的方法，包括请求处理相关的方法和响应处理相关的方法。

请求处理相关的方法

在 Context 对象中，提供了一些方便用户使用的请求处理方法，这些方法是对 Go 标准库中的 Request 对象进行了封装，方便用户使用。这些方法包括：

- Bind()：将请求中的 Body 解析为结构体。
- ShouldBind()：根据请求的 Content-Type 自动解析请求 Body，支持 JSON、XML、ProtoBuf、MsgPack 等格式。
- ShouldBindJSON()：解析请求中的 JSON Body。
- ShouldBindXML()：解析请求中的 XML Body。
- ShouldBindQuery()：解析 URL 中的 Query 参数。
- ShouldBindUri()：解析 URL 中的参数。
- ShouldBindHeader()：解析请求 Header 中的参数。
- Param()：获取 URL 中的参数。
- Query()：获取 URL 中的 Query 参数。
- PostForm()：获取表单提交的参数。
- DefaultQuery()：获取 URL 中的 Query 参数，支持设置默认值。
- GetHeader()：获取请求 Header 中的参数。
- GetRawData()：获取请求 Body 中的原始数据。

### context 优势

Gin 的 **`Context`** 相比于其他框架的 **`Context`**，具有以下优势：

1. 快速的错误处理：Gin 中的 **`Context`** 封装了快速的错误处理，可以在处理请求的过程中通过 **`c.AbortWithError()`** 方法返回错误，并终止请求的处理。
2. 参数解析：Gin 中的 **`Context`** 内置了参数解析功能，可以通过 **`c.Param()`**、**`c.Query()`**、**`c.PostForm()`** 等方法解析 URL 参数、Query 参数以及 Form 参数等。
3. 中间件支持：Gin 中的 **`Context`** 通过中间件机制可以方便地实现请求处理链式调用，例如路由认证、权限控制、日志记录等操作。
4. 渲染响应：Gin 中的 **`Context`** 内置了响应渲染功能，可以通过 **`c.JSON()`**、**`c.XML()`**、**`c.String()`** 等方法将数据渲染成 JSON、XML、文本等格式的响应。
5. 统一的错误处理：Gin 中的 **`Context`** 通过 **`c.JSON()`** 方法可以方便地统一处理错误响应，避免了在业务代码中编写大量的错误处理代码。

总之，Gin 的 **`Context`** 通过内置的方法和功能，使得开发者可以更加方便地处理请求和响应，并且减少了编写重复代码的工作量，提高了开发效率。
