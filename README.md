# gobean

```shell
go get -u github.com/erkesi/gobean
```

> 每一个模块，都有测试用例，可以通过测试用例了解具体的使用方法

## log 包

```go

import github.com/erkesi/gobean/log

```

> 打印 gobean 项目下的包 debug 级别的日志

### 方法

#### 初始化全局 Logger 实例

> log.Init(logger Logger)

## inject 包

> 依赖注入

```go

import github.com/erkesi/gobean/inject

```

> 在依赖关系的顺序下，再按照实例注入的优先级和顺序，完成实例的生命周期的管理

- 封装了 [facebookarchive/inject](https://github.com/facebookarchive/inject)，[inject 如何标签使用](https://pkg.go.dev/github.com/facebookgo/inject)

- 简单增加实例的生命周期管理，Struct 类型 可以实现 Init() \ Close() 来实现初始化、销毁实例

- 使用示例：[inject_test.go](inject/inject_test.go)

### 方法

#### 实例的生命周期


```go

type B struct{
}

type A struct{
	B *B `inject:""`
}

// Init, 实现 inject.ObjectInit 接口，完成实例的初始化
func (a *A) Init() {
}

// Close, inject.ObjectClose 接口，完成实例的销毁
func (a *A) Close(){
}

```

#### 按照类型注入实例

- 参数 opts: 可以为 inject.ProvideWithPriority(priority int) , 顺序【依赖关系与优先级（从大到小）】完成实例的初始化

> inject.ProvideByValue(value interface{}, opts ...ProvideOptFunc)

#### 依据类型获取实例

> inject.ObtainByType(value interface{}) interface{}

#### 按照命名注入实例

- 参数 opts: 可以为 inject.ProvideWithPriority(priority int) , 按照顺序【依赖关系与优先级（从大到小）】完成实例的初始化

> inject.ProvideByName(name string, value interface{}, opts ...ProvideOptFunc) 

#### 按照命名获取实例

> inject.ObtainByName(name string) interface{}

#### 完成依赖注入以及按照顺序【依赖关系与优先级（从大到小）】完成实例的初始化（Init()）

> inject.Init()

#### 实例初始化的逆向顺序销毁实例（Close()）

> inject.Close()

#### 按照依赖注入的关系打印实例列表

> inject.PrintObjects()

## extpt 包

> 扩展点能力

```go

import github.com/erkesi/gobean/extpt

```



> 基于 inject 依赖注入的能力

> 执行的时候，依据扩展点接口，找到多个实现，按照优先级逐个匹配，如果匹配（Match() == true），则执行后返回。

> 使用示例：[extension_pointer_test](extpt/extension_pointer_test.go)

### 方法

#### 注册扩展点实例

> extpt.Hub.Register(et ExtensionPointer, opts ...ExtPtFunc)

#### 执行

> extpt.Execute(f interface{}, args ...interface{}) (interface{}, bool)

> extpt.ExecuteWithErr(f interface{}, args ...interface{}) (interface{}, error, bool)


## application 包

> 应用启动与销毁回调函数的注册以及使用

```go

import github.com/erkesi/gobean/application

```

> 使用示例：[application_test.go](application/application_test.go)

### 方法

#### 注册应用启动回调函数 

- 参数 opts: 可以为 application.CallbackWithPriority(priority int) , 顺序【优先级（从大到小）】执行 callback

> application.AddInitCallback(callback appStateCallback, opts ...OptFunc)

#### 注册应用销毁回调函数

- 参数 opts: 可以为 application.CallbackWithPriority(priority int) , 顺序【优先级（从大到小）】执行 callback

> application.AddCloseCallback(callback appStateCallback, opts ...OptFunc)

#### 应用初始化，按照优先级顺序（从大到小）调启动函数 

> application.Init()

#### 应用销毁，按照优先级顺序（从大到小）调销毁函数

> application.Close()

