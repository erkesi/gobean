# gobean

```shell
go get -u github.com/erkesi/gobean
```

> 每一个模块，都有测试用例，可以通过测试用例了解具体的使用方法

## glogs 包

```go

import github.com/erkesi/gobean/glogs

```

> 打印 gobean 项目下的包 debug 级别的日志

### 方法

#### 初始化全局 Logger 实例

> glogs.Init(logger Logger)

## ginjects 包

> 依赖注入

```go

import github.com/erkesi/gobean/ginjects

```

> 在依赖关系的顺序下，再按照实例注入的优先级和顺序，完成实例的生命周期的管理

- 封装了 [facebookarchive/inject](https://github.com/facebookarchive/inject)，[inject 如何标签使用](https://pkg.go.dev/github.com/facebookgo/inject)

- 简单增加实例的生命周期管理，Struct 类型 可以实现 Init() \ Close() 来实现初始化、销毁实例

- 使用示例：[inject_test.go](ginjects/inject_test.go)

### 方法

#### 实例的生命周期


```go

type B struct{
}

type A struct{
	B *B `inject:""`
}

// Init, 实现 ginjects.ObjectInit 接口，完成实例的初始化
func (a *A) Init() {
}

// Close, ginjects.ObjectClose 接口，完成实例的销毁
func (a *A) Close(){
}

```

#### 按照类型注入实例

- 参数 opts: 可以为 ginjects.ProvideWithPriority(priority int) , 顺序【依赖关系与优先级（从大到小）】完成实例的初始化

> ginjects.ProvideByValue(value interface{}, opts ...ProvideOptFunc)

#### 依据类型获取实例

> ginjects.ObtainByType(value interface{}) interface{}

#### 按照命名注入实例

- 参数 opts: 可以为 ginjects.ProvideWithPriority(priority int) , 按照顺序【依赖关系与优先级（从大到小）】完成实例的初始化

> ginjects.ProvideByName(name string, value interface{}, opts ...ProvideOptFunc) 

#### 按照命名获取实例

> ginjects.ObtainByName(name string) interface{}

#### 完成依赖注入以及按照顺序【依赖关系与优先级（从大到小）】完成实例的初始化（Init()）

> ginjects.Init()

#### 实例初始化的逆向顺序销毁实例（Close()）

> ginjects.Close()

#### 按照依赖注入的关系打印实例列表

> ginjects.PrintObjects()

## gevents 包

> 事件的发布与订阅

```go

import github.com/erkesi/gobean/gevents

```

> 基于 ginjects 依赖注入的能力

> 使用示例：[event_test.go](gevents/event_test.go)

### 方法

#### 注册事件处理器

> gevents.Register(executors ...Executor)


#### 设置事件默认处理器

> gevents.SetDefaultExecutor(executor Executor)

#### 发布事件

> 发布（接口：[Publisher](gevents/publisher.go)）

> Publish(ctx context.Context, event interface{}, opts ...PublishOpt) error 

> 使用默认的事件发布器

> &DefaultPublisher{}


## gextpts 包

> 扩展点能力

```go

import github.com/erkesi/gobean/gextpts

```

> 基于 ginjects 依赖注入的能力

> 执行的时候，依据扩展点接口，找到多个实现，按照优先级逐个匹配，如果匹配（Match() == true），则执行后返回。

> 使用示例：[extension_pointer_test.go](gextpts/extension_pointer_test.go)

### 方法

#### 注册扩展点实例

> gextpts.Register(et ExtensionPointer, opts ...ExtPtFunc)

#### 执行

> gextpts.Execute(ctx context.Context, f interface{}, args ...interface{}) (bool, interface{})

> gextpts.ExecuteWithErr(ctx context.Context, f interface{}, args ...interface{}) (bool, interface{}, error)

## gstatemachines 包
> 简单状态机实现
> - 定义状态转移流程
> - 定义状态接口
> - 状态机执行

```go

import github.com/erkesi/gobean/gstatemachines

```

> 使用示例：[state_machine_test.go](gstatemachines/state_machine_test.go)


## gapplications 包

> 应用启动与销毁回调函数的注册以及使用

```go

import github.com/erkesi/gobean/gapplications

```

> 使用示例：[application_test.go](gapplications/application_test.go)

### 方法

#### 注册应用启动回调函数 

- 参数 opts: 可以为 gapplications.CallbackWithPriority(priority int) , 顺序【优先级（从大到小）】执行 callback

> gapplications.AddInitCallback(callback appStateCallback, opts ...OptFunc)

#### 注册应用销毁回调函数

- 参数 opts: 可以为 gapplications.CallbackWithPriority(priority int) , 顺序【优先级（从大到小）】执行 callback

> gapplications.AddCloseCallback(callback appStateCallback, opts ...OptFunc)

#### 应用初始化，按照优先级顺序（从大到小）调启动函数 

> gapplications.Init()

#### 应用销毁，按照优先级顺序（从大到小）调销毁函数

> gapplications.Close()

## gobean/gstreamings

### 数据流处理，source -> transform -> sink

## gobean/gstreams
### 集合数据流式处理，类似 java stream
