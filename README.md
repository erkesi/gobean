# gobean

## log 包，打印 gobean 项目下的包 debug 级别的日志

```shell
go get github.com/erkesi/gobean/log@latest
```

### 方法

#### 初始化全局 Logger 实例

> log.Init(logger Logger)

## inject 包，依赖注入

```shell
go get github.com/erkesi/gobean/inject@latest
```

> 在依赖关系的顺序下，再按照实例注入的优先级和顺序，完成实例的生命周期的管理

- 封装了 [facebookarchive/inject](https://github.com/facebookarchive/inject)

- 简单增加实例的生命周期管理，Struct 类型 可以实现 Init() \ Close() 来实现初始化、销毁实例

### 方法

#### 按照类型注入实例

> inject.ProvideByValue(value interface{}, opts ...ProvideFunc)

#### 依据类型获取实例

> inject.ObtainByType(value interface{}) interface{}

#### 按照命名注入实例

> inject.ProvideByName(name string, value interface{}, opts ...ProvideFunc) 

#### 按照命名获取实例

> inject.ObtainByName(name string) interface{}

#### 初始化依赖注入关系以及完成实例的初始化（Init()）

> inject.Init()

#### 按照依赖注入的关系逆向销毁实例（Close()）

> inject.Close()

#### 按照依赖注入的关系打印实例列表

> inject.PrintObjects()

## extpt 包，扩展点能力

```shell
go get github.com/erkesi/gobean/extpt@latest
```

> 基于 inject 依赖注入的能力

> 执行的时候，依据扩展点接口，找到多个实现，按照优先级逐个匹配，如果匹配（Match() == true），则执行后返回。


### 方法

#### 注册扩展点实例

> extpt.Hub.Register(et ExtensionPointer, opts ...ExtPtFunc)

#### 执行

> extpt.Execute(f interface{}, args ...interface{}) (interface{}, bool)

> extpt.ExecuteWithErr(f interface{}, args ...interface{}) (interface{}, error, bool)


## application 包，应用启动与销毁

```shell
go get github.com/erkesi/gobean/application@latest
```

> 应用启动与销毁回调函数的注册以及使用

### 方法

#### 注册应用启动回调函数 

> application.AddInitCallbackWithPriority(priority int, callback appStateCallback)

#### 注册应用销毁回调函数

> application.AddCloseCallbackWithPriority(priority int, callback appStateCallback)

#### 应用初始化，按照优先级顺序（从大到小）调启动函数 

> application.Init()

#### 应用销毁，按照优先级顺序（从大到小）调销毁函数

> application.Close()

