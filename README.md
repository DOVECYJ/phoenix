# Phoenix

这是一个模仿Elixir的Phoenix框架的Go web框架。有着和Phoenix类似的项目布局和概念，比如application，endpoint，router等。

## 安装

安装phoenix命令行工具 `phx` ：

```
go install github.com/chenyj/phoenix/cmd/phx@latest
```

`phx` 需要依赖 [templ](https://github.com/a-h/templ) , 通过下面的命令安装:

```
go install github.com/a-h/templ/cmd/templ@latest
```

## 快速开始

使用 `phx` 新建项目：`phx new hello` ，新项目目录结构如下：

**hello**
|---- _build
|---- assets
|-------- css
|-------- js
|-------- image
|---- config
|-------- application.toml
|---- lib
|-------- hello
|------------ application.go
|-------- hello\_web
|------------ components
|---------------- layout.templ
|------------ controllers
|---------------- page\_html
|-------------------- page.templ
|---------------- page\_controller.go
|------------ endpoint.go
|------------ router.go
|---- pkg
|---- priv
|-------- repo
|------------migrations
|------------migrate.go
|---- main.go
|---- README.md

- `_build` 是编译结果存放的目录，部署的时候直接将该文件夹打包即可。
- `assets` 是前端资源文件存放目录，如css，js等，项目默认使用bootstrap4。
- `config` 下存放的是项目配置文件，默认采用toml格式的配置文件。
- `lib` 是项目核心代码的存放目录，业务逻辑和web层的代码都在这里。
- `pkg` 内是本项目中使用的功能相对完整独立的包，将来可能抽离出去做为独立的库。
- `priv` 是项目私有文件，包括数据库模型迁移等，不应该被任何包引用。
- `main.go`是项目启动入口，这里会做一些公共初始化，如加载配置文件，配置日志等，以及启动application。

`lib` 目录下是项目核心代码，主要分为业务层和web层。

- `hello` 是业务逻辑层，简单来说就是放CURD代码的。
- `application.go` 是定义应用的地方，应用的定义放在业务逻辑层。
- `hello_web` 是应用的web接口层，这里只有和web相关的代码，在项目增长过程中，应该始终保持该层是"轻薄"的。
- `components` 是存放公共组件的地方，所谓组件可以简单理解为web页面模板。
- `controllers` 是MVC中的C所在的地方，在这里调用业务层，实现web功能。
- `endpoint.go` 是web层的入口，它会被application调用，从而为application提供web接口。
- `router.go` 是路由器，在这里实现controller的路由注册。

项目默认使用MySQL数据库，打开项目目录下的 `config/application.toml` 可以看到数据库配置，需要根据实际情况修改。你可以通过 `--database` 选项来修改数据库，支持以下3种：

```plaintext
--database mysql
--database pgsql
--database sqlite3
```

如果不需要数据库，也可以使用 `--no-database` 去掉数据库依赖。

此外，项目默认使用了redis做为缓存，如果不需要，可以使用 `--no-redis` 去掉redis依赖。如果你开发的是一个API项目，不需要用到HTML模板，可以使用 `--no-html` 参数去掉HTML模板。

运行项目：`go run .`

打开浏览器输入 http://localhost:8080，你会看到一个欢迎页面。

## 启动过程

应用的启动过程分为两步：

首先是 `main.go` ，完成配置文件读取，日志等基础组件的初始化，然后通过 `RunApplications` 运行应用。

第二步是应用启动，它会去配置应用所需的资源，如数据库，缓存等，然后启动web接口。application需要提供 `Start()` 和 `Stop()` 两个接口。

web接口不是项目的核心，甚至也不是应用的核心。web层只是暴露应用功能的一种方式而已，你完全可以将它替换成rpc等其他方式。如果你的应用需要其他服务，也应该在application中统一启动。

## 添加路由

找到你的项目目录下的 `lib/hello_web/router.go` ，默认有一个主页面的路由，这就是你看到的欢迎页。

```go
func route(root chi.Router) {
    root.Get("/", controllers.Index)
}
```

Phoenix使用了[Chi](https://github.com/go-chi/chi)做为路由器，除了可以使用chi本身提供的路由注册方式，Phoenix还提供了 `Resource` 、`ResourceOnly` 和 `ResourceExcept` 函数通过 `IResource` 接口快速注册REST API。

```go
type IResource interface {
    Index(http.ResponseWriter, *http.Request)  // index: show a list of object
    Edit(http.ResponseWriter, *http.Request)   // edit: show edit form
    New(http.ResponseWriter, *http.Request)    // new: show create object form
    Show(http.ResponseWriter, *http.Request)   // show: show one object detail by id
    Create(http.ResponseWriter, *http.Request) // create: save a new object
    Update(http.ResponseWriter, *http.Request) // update: save update object
    Delete(http.ResponseWriter, *http.Request) // delete: delete a object by id
}
```

假如我们有一个 `UserController` 控制器，并通过 `Resource` 注册到 "/users" 路径下：

```go
type UserController struct {
    router.IResource
}

func route(root chi.Router) {
    ...
    root.Route("/users", router.Resource(UserController{}))
}
```

那么我们会拥有以下路由：

```plaintext
GET    /users
GET    /users/{id}/edit
GET    /users/new
GET    /users/{id}
POST   /users
PATCH  /users/{id}
PUT    /users/{id}
DELETE /users/{id}

```

使用 `Resource` 方式注册路由时也可以使用中间件，如下：

```plaintext
root.Route("/users", func(r chi.Router) {
    r.Use(middleware.Logger)
    router.Resource(UserController{})(r)
})
```

如果不是一个完整的REST路由，可以使用 `ResourceOnly` 和 `ResourceExcept` 指定或排除路由。

## 添加页面

在项目目录下的 `lib/hello_web/controllers/` 中，有一个 `page_html` 目录，他是 `page_controller.go` 控制器对应的页面组件。其中 `page.templ` 是模板文件，它通过 `templ generate` 命令编译成 `page_templ.go` 。

打开 `page.templ` ，其中的 `teml Index()` 就是欢迎页组件。在文件中添加以下代码：

```go
templ Hello(name string) {
    @Layout() {
        <div align="center"><h1>Hello, { name }</h1></div>
    }
}
```

在项目目录下运行 `templ generate` 生成go代码。回到 `page_controller.go` 添加以下action：

```go
func Hello(w http.ResponseWriter, r *http.Request) {
    name := chi.URLParam(r, "name")
    render.HTML(w, pagehtml.Hello(name))
}
```

在 `lib/hello_web/router.go` 中添加以下路由：

```go
func route(root chi.Router) {
    ...
    root.Get("/{name}", controllers.Hello)
}
```

运行程序，在浏览器访问 `localhost:8080/tom` 查看效果。

## 模型迁移

模型迁移文件在 `priv/repo/migrations` 目录下，执行迁移可以使用下面的命令：

```
phx migrate
```

使用rollback可以进行回滚：

```
phx rollback
```

回滚支持通过 `--n` 选择指定回滚次数，或通过 `--v` 选项指定回调到某个版本，包括被指定的版本。

默认会使用 `config/application.toml` 配置文件中的数据配置，如需指定其他配置文件，可以使用 `--config myconfig.toml` 选项。

## TODO

- 服务注册与发现
- 监控接入

## 鸣谢

- 感谢[chi](https://github.com/go-chi/chi)提供的路由支持以及中间件。
- 感谢[viper](https://github.com/spf13/viper)提供的配置管理。
- 感谢[templ](https://github.com/a-h/templ)提供的模板支持。
- 感谢[rel](https://github.com/go-rel/rel/)提供的orm支持。
