# XuperChain开源贡献指南

[ [English](https://github.com/xuperchain/xuperchain/blob/master/CONTRIBUTING.md) | 简体中文 ]


# 贡献准备
## 签署贡献者许可协议
我们非常鼓励您向超级链(XuperChain)开源的代码库提交社区贡献，为了保证社区用户能免费使用您贡献的开源代码，您需要首先签署一份贡献者许可协议。

您可以通过[XuperChain Contributor License Agreement](https://cla-assistant.io/xuperchain/xuperchain)浏览并签署超级链贡献者许可协议，您也可以在提交PR时通过自动回复的链接签署CLA。

# 设置本地开发环境
XuperChain 

## 安装 GO 1.13 或更高版
## (可选)安装 C++ 开发工具链
XuperChin 依赖 webassembly，如果需要使用 wasm 合约，需要安装 C++ 工具链
## (可选)安装 Docker
XuperChain 使用Docker 来实现原生合约的资源隔离以C++ 合约编译构建，如果你需要使用 C++ 合约，或者希望在Docker 中运行 native 合约，你需要安装 Docker  
// TODO 版本
## (可选) 安装 JAVA 开发环境
如果你希望为 XuperChain 的 JAVA 相关部分贡献，需要安装 JAVA 1.8+，同时需要安装 maven 
// TODO 版本
## （可选）安装 nodejs 开发环境

## (可选) 安装 Rust 开发环境

GRPC ????

# 提交第一个Pull Request



## 创建 Issue
XuperChain使用github的issue系统来管理`bug`和`功能提议`。在您创建一个新的issue之前，请首先通过issue搜索查看是否已经存在相同或类似的issue，如果您的情况跟已存在的issue类似，您只需要在已存在的issue中通过回复追加您的情况即可。如果您确认不存在类似issue，那么别犹豫，请创建一个新的issue。

在创建issue时，请**使用**我们已经提供的**issue模板**，并在模板中填写下述内容：

* 简要的描述issue的主要问题。
* 您的系统环境。 包括XuperChain的版本，OS/Go/Gcc 版本等信息。
* 请描述您认为应该发生的情况。
* 请描述您实际观察到的情况。
* 请提供日志、截图等可能有用的线索，如果可能，请提供一份最小可验证的代码样例帮助我们确认问题。

**注意：** Issue系统只面向提交bug和功能建议使用。如果您是想咨询技术细节或其他问题，请首先查看我们的wiki，如果wiki没有解决您的问题，请通过首页提供的微信二维码加入超级链官方微信技术群以获取更多帮助。**任何不符合上述要求的issue，可能会在无任何通知的情况下被关闭。**

# 提交代码
## Commit Message 规范
## Pull Request 规范
XuperChain通过github标准PR(Pull Request)流程接受代码贡献。请首先fork代码库，将修复代码提交到fork的代码库，并提交一个PR到XuperChain。代码维护人员会在收到PR后，对您提交的代码进行Code Review，并在确认没问题后合入主干代码。

请使用下述模板提交 pull request：

* PR中代码修改的目的是什么？
* 简要描述您解决此问题的方案。

所有的PR需要对应到issue，并在PR描述中附带对应的issue编号。您提交的PR需要首先通过CI流程验证以降低引入新问题的可能，之后还需要通过至少两位代码维护人员的Review，这可能需要一定时间，请您耐心等待。

**注意：** 您的代码必须遵循官方的Go语言代码规范(可以使用gofmt和golint进行检查)。

## XuperChain PR 工作流

我们欢迎任何形式的贡献
## 完善 XuperChain 文档
XuperChain 采用 ReStructuredText(TODO) 作为我们的文档写作工具,你可以在链接地址找到更多关于 ReStructuredText(TODO) 语法的教程
你也可以在测试页面(TODO)测试页面快速熟悉语法
### 改进现有内容
如果你希望对XuperChain 文档做少量修改，你可以直接在文档网站（link） 上进行修改。
### 贡献新的文档内容
如果你想要给 XuperChain 文档贡献新内容，或者进行较大规模的改动，我们推荐通过完整的 github PR 流程进行内容提交(TODO add link)。
我们提供了交互式热更新的文档写作工具，你可以通过如下步骤，实现文档修改的实时预览。

```shell script
    git clone 
    make build-image
    make server
```


在你本地验证通过后，就可以通过完整的 PR 流程(link)将变更提交到网站了

详见 docs 仓库 README.md（Link）

## 给 XuperChain 贡献新特性 
XuperChain 通过 github issue 来管理特性，如果你希望 XuperChain 贡献新特性，
## 修复 XuperChain 中的 Bug
XuperChain 通过github issue 来进行 bug 的管理。
如果你在使用XuperChain 的过程中发现了新BUG，可以提交
XuperChain 采用主干开发，分支发布（Link）模型，默认对最新的三个小版本提供(TBD,Link)提供LTS(Long Time Support) 支持，对于三个版本之外的版本，仅提供安全更新。
当你修复了 XuperChain 的 bug 时需要同时将有关修改 cherry-pick(link) 到release 分支上。
## 开发 XuperChain 插件
XuperChain 
## 贡献 XuperChain 语言 SDK
XuperChain 当前提供了 GO， JAVA，Python, JavaScript 的语言 SDK，你可以通过标准的贡献流程(link)进行完善。
如果你希望给新的语言贡献 SDK，或者实现一套新的语言SDK，你需要阅读以下内容
## 给 XuperChain 提交 Feature Request
你在使用 XuperChain 过程中的需求是 XuperChain 持续演进的动力。如果你在使用XuperChain 的过程中需要 XuperChain 使用新的功能，你可以通过 Feture Request 的方式给我们提交功能请求
你可以点击 Star 收藏仓库或者 watch 仓库，以便在有新功能更新时收到通知。
## 给 XuperChain 提交 Bug Report
XuperChain 通过 github Issue 管理 Bug Report, 如果您在使用XuperChain 的过程中发现了BUG，你可以通过，github Issue 提交

## XuperChain 应用案例


# XuperChain 版本管理规范


# 贡献准备
# 提交第一个 Pull Request
# 贡献流程
# 其他
 


import 
签署commit
