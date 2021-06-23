# XuperChain开源贡献指南

[ [English](https://github.com/xuperchain/xuperchain/blob/master/CONTRIBUTING.md) | 简体中文 ]

XuperChain 欢迎任何形式的贡献，你可以给 XuperChain 贡献代码，完善 XuperChain 文档，或者给 XuperChain 提交 Feature Request 和 Bug Report。
我们也欢迎你在社区中分享你使用 XuperChain 的经验。

# 给 XuperChain 贡献代码
## 贡献准备
在正式贡献代码之前，你需要做一些准备工作，

1. 签署贡献者许可协议
在正式贡献代码之前，你需要先签署贡献者协议。
你可以通过[XuperChain Contributor License Agreement](https://cla-assistant.io/xuperchain/xuperchain)浏览并签署超级链贡献者许可协议，也可以在提交PR时通过自动回复的链接签署CLA。

<!-- 2. [可选]对代码进行签名 -->


## 设置本地开发环境
如果你打算给 XuperChain 贡献代码，你需要设置本地开发环境
XuperChain 当前在 Linux 和 MacOS 上进行验证，当前尚未在 Windows 系统上经过验证
1. 安装基础依赖工具
   XuperChain 依赖如下工具,在正式开始前请确保你已完成如下工具的安装
   - git
   - make
   - GO 1.13 +
   - [可选] protocol buffer 

2. (可选)安装 C++ 开发工具链
    XuperChin 依赖 C++ 工具链来实现对 webassembly 的提前静态编译，如果你使用 wasm 合约，需要安装 C++ 工具链
3. (可选)安装 Docker
    XuperChain 使用 Docker 来实现原生合约的资源隔离以及 C++ 合约编译构建，如果你需要使用 C++ 合约，或者希望在 Docker 中运行 native 合约，你需要安装 Docker  
4. (可选) 安装 JAVA 开发环境
    如果你使用 XuperChain 的 JAVA SDK 进行应用开发， 或者使用 JAVA 语言进行合约开发 ，需要安装 JAVA 1.8+
    你也需要安装 maven 以完成项目的构建


## 提交第一个Pull Request

1. clone 代码到本地
   ``` bash
   git clone github.com/xuperchain/
   ```
2. 构建本地 debug 版本
   xuperchain 提供构建调试版本的能力， 调试版本保留更多的调试相关信息，以便进行有关问题的定位和调试
    ``` bash
    cd xuperchain 
    make build-debug
    ```

3. 修改本地代码
4. 增加单元测试
   你需要为你增加的功能增加单元测试，确保你提交不会被后续的提交破坏
5. [可选]本地运行单元测试，确保你的更改没有导致其他的的测试 break
   ``` bash 
   make test
   ```
    你也可以选择将单测试推迟到 github CI 阶段执行。
<!-- 6. [可选] 确保你的代码符合代码风格
   ```bash
   make lint 
   ``` -->

5. 提交 Pull Request
   完成以上内容之后，你就可以通过 Pull Request 提交你的代码了。
   你可以查看 [代码提交指南](https://xuper.baidu.com/n/xuperdoc/contribution/pull_requests.html) 获取更多关于代码提交的指南


<!-- ## Pull Request 规范 -->
<!-- XuperChain 通过 github 标准PR(Pull Request)流程接受代码贡献。请首先fork代码库，将修复代码提交到fork的代码库，并提交一个PR到XuperChain。代码维护人员会在收到PR后，对你提交的代码进行Code Review，并在确认没问题后合入主干代码。

请使用下述模板提交 pull request：

* PR中代码修改的目的是什么？
* 简要描述你解决此问题的方案。

所有的PR需要对应到issue，并在PR描述中附带对应的issue编号。你提交的PR需要首先通过CI流程验证以降低引入新问题的可能，之后还需要通过至少两位代码维护人员的Review，这可能需要一定时间，请你耐心等待。  -->




## 给 XuperChain 贡献新特性 

XuperChain 通过 github issue 来管理新 Feature 的提交。
在正式开始之前，你需要在 github 建立 [Issue](https://github.com/xuperchain/xuperchain/issues)，以进行必要的讨论。

你可以使用[Issue 模板](https://github.com/xuperchain/xuperchain/issues/new?assignees=&labels=&template=xuperchain-feature-request-template.md&title=) 来描述你的想法
我们强烈建议在进行新特性贡献之前通过 Issue 进行充分的讨论，以便获得必要的支持

## 修复 XuperChain 中的 Bug

XuperChain 通过 github issue 来进行 bug 的管理。
正式提交你的 Bug Report 之前，你需要在 [Issue 列表](https://github.com/xuperchain/xuperchain/pulls) 中搜索是否有类似的问题以及是否已经有解决方案。如果你在使用 XuperChain 的过程中发现了新 BUG，可以按照 [Bug Report 模板](https://github.com/xuperchain/xuperchain/issues/new?assignees=&labels=&template=xuperchain-bug-template.md&title=)提交 Bug Report

你可以在 github Issue 区 选择你感兴趣的 Bug 进行修复，帮助社区的其他用户解决他们遇到的问题。

<!-- 我们建议在进行修复前通过 github issue 进行充分讨论，以便获得更好的支持 -->

<!-- ## 开发 XuperChain 插件
XuperChain 通过动态内核技术，实现核心组件的可以插拔, 实现核心组件的插件化支持，你可以为XuperChain 贡献插件，优秀的插件实现将有机会被合到 XuperChain 主线代码中

## 为 XuperChain 语言 SDK 贡献
XuperChain 当前提供了 GO， JAVA，Python, JavaScript 的语言 SDK， 你可以给 XuperChain 贡献其他语言的 SDK 或者完善现有 SDK 的功能

如果你希望给新的语言贡献 SDK，或者实现一套新的语言SDK，你需要阅读以下内容
合约 SDK 设计 -->

## XuperChain 调试技巧

如果你希望给XuperChain 贡献代码，你可能会遇各种的问题，XuperChain 提供了一系列的工具来帮助开发者定位和修复遇到的各种问题
1. 开发模式
   XuperChain 提供了开发模式，以方便你在开发的过程中进行有关调试。
   在编译 xuperchain 时执行以下命令即可完成开发者模式的编译
   ```bash
      make build-debug 
   ```
   开发者模式下进行构建会关闭一些优化选项，以便更好地和调试器进行协同工作。
   <!-- 开发者模式的主要区别在于 -->
   <!-- a. 关闭内联以便和调试器更好地配合
   b. 日志级别调整为 debug 以输出更多的调试信息
   c. 使用 ixvm 以加快合约的部署
   d. 采用 single 共识避免 xxxx
   e. 打开 pprof 调试工具以
   f. 关闭原生合约健康检查，以便 attach 到合约进程 -->
2. XuperChain 日志
   XuperChain 提供了日志以方便进行问题的定位和修复，你可以适当调整日志级别，以便能够输出更多的有用的调试信息。
   XuperChain 默认日志存储在工作目录下的 logs 目录下，日志拆分成以下几个文件
   - xchain.log xchain 进程的日志
    该文件保存了xchain 节点进程的日志你可以在该文件中查看xchain 运行中的各种日志记录
   - contract.log 合约日志
    该文件保存了合约的运行日志，所有通过 ctx.Logf 输出的日志都会保存在这个文件中
   - pluginmrg.log 插件日志
    该文件保存了插件管理相关的日志，如果你在使用过程中遇到插件相关问题，你可以查看该文件
    xchain 节点日志通过 logid 来传递日志在不同组件中上下文，你可以通过logid 来查看一个请求在各个组件中的执行情况
3. 如果你是合约开发者，你可以使用 [xdev](https://github.com/xuperchain/xdev) 来进行合约的验证和测试
<!-- 4. IDE 调试
   如果你在开发过程中希望使用 IDE 进行断点调试，你可以使用 XuperChain 
   
   a. 使用 GoLand 调试 XuperChain 
   b. 使用 VSCode 调试 XuperChain 
   c. 调试原生合约 -->


   <!--  should keep in sync with docs repo README -->
# 完善 XuperChain 文档
我们欢迎任何形式的文档贡献，包括
1. 修改文档中的错别字
2. 修改文档的描述，使之更加精确
3. 修正文档中的过期内容
4. 完善已有内容
5. 贡献一篇全新的内容
6. 分享使用案例


### 完善已有内容文档
如果你需要对 XuperChain 文档进行少量的更改，你可以直接使用在线编辑提交你的变更
1. 浏览器打开 [XuperChain 文档官网](https://xuper.baidu.com/n/xuperdoc/index.html)
2. 导航到你需要修改的文档页面
3. 点击右上角"编辑此页"按钮
4. 修改内容后点击 "Commit Change" 提交修改

提交后会自动生成一个 Pull Request,待 Pull Request 合并后，你就可以在 [XuperChain 文档官网](https://xuper.baidu.com/n/xuperdoc/index.html)看到你的提交了

### 贡献复杂内容
如果你需要对已有文档做较多的调整，或者希望贡献一篇完整的内容，你可以
1. 查看并签署[贡献者协议](https://cla-assistant.io/xuperchain/xuperchain)
2. 查看 [贡献指南](https://github.com/xuperchain/xuperchain/blob/master/CONTRIBUTING.md)
3. 查看 [代码提交指南](https://xuper.baidu.com/n/xuperdoc/contribution/pull_requests.html)
4. 本地编辑文件
5. 提交完整的 Pull Request


# 参加 XuperChain 社区
## 在社区解答问题

你可以在 [XuperChain 官方论坛](https://developer.baidu.com/singleTagPage.html?tagId=49&type=QUESTION)向其他开发者展示你的观点和看法，也可以在论坛中查找你遇到的问题的解答。
我们推荐在 Issue 区中进行和 XuperChain 相关的技术以及功能需求的探讨，在论坛里你可以进行包括但是不限于 XuperChain 的区块链相关的讨论。

## 分享 XuperChain 应用案例
你可以在[XuperChain 社区](https://developer.baidu.com/article.html#/articleHomePage)发布文章,分享你使用 XuperChain 的情况。优秀的应用案例将会得到 XuperChain 官方的转载和有关的推广曝光
<!-- 参加 XuperChain SIG(特别兴趣小组) -->


## 给 XuperChain 提交 Feature Request 或者 Bug Report
如果你是 XuperChain 的非开发者用户，在使用 XuperChain 的过程中有一个"很不错的想法"，我们欢迎你通过[Feature Request](https://github.com/xuperchain/xuperchain/issues/new?assignees=&labels=&template=xuperchain-feature-request-template.md&title=)的时候分享你的想法。你的想法可能会被社区开发者以及 XuperChain 的核心开发团队实现，最终成为 XuperChain 产品的一部分。

我们也欢迎你通过 [Bug Report](https://github.com/xuperchain/xuperchain/issues/new?assignees=&labels=&template=xuperchain-bug-template.md&title=) 反馈你在使用 XuperChain 的过程中遇到的 Bug 

## 参加 XuperChain 线下活动

XuperChain 会不定期举行比赛，沙龙等各类线下活动，你可以在活动中和其他开发者分享你的使用 XuperChain 技巧，了解其他使用者的使用 XuperChain 的经验，以及与 XuperChain 的核心开发者面对面的交流。
你可以添加超级链小助手微信获取更多关于线下活动的信息


<!-- ## Commit Message 规范(Pull Request 规范) -->

