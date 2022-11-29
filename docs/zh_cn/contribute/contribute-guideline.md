# XuperChain开源贡献指南

中文 | [English](../../en_us/contribute/contribute-guideline.md)

---

首先，非常感谢您考虑向XuperChain提交开源贡献。

在此之前，请阅读下面的开源贡献指南，并遵守 [Issue] 和 [PR] 提交规则。
为了大家能在 GitHub 上无障碍沟通，我们强烈建议您使用英文提交 [Issue] 或 [PR]。

## 贡献信息/想法

XuperChain使用 GitHub 的 [Issue] 系统来管理`bug`和`功能提议`，您可以通过创建一个 Issue 来贡献您的信息或想法。

在您创建一个新的 Issue 之前，请首先通过 [Issue] 搜索查看是否已经存在相同或类似的 Issue。
如果您的情况跟已存在的 Issue 类似，您只需要在已存在的 Issue 中通过回复追加您的情况即可。
如果您确认不存在类似 Issue，那么别犹豫，请创建一个新的 Issue。

### Issue 要求

在创建 Issue时，请**使用**我们已经提供的**Issue模板**，并在模板中填写下述内容：

* 简要的描述 Issue 的主要问题。
* 您的系统环境信息，包括但不于:
  * XuperChain的版本
  * OS/Go/Gcc 版本等信息
* 请描述您认为应该发生的情况。
* 请描述您实际观察到的情况。
* 请提供其他可能有用的线索，帮助我们确认问题。如：
  * 日志
  * 截图
  * 最小可验证的代码样例等

> **注意：**
>
> Issue系统只面向提交bug和功能建议使用。
> 如果您是想咨询技术细节或其他问题，请首先查看我们的[文档][DocsSite]，
> 如果文档没有解决您的问题，请通过首页提供的[微信二维码][Contact]加入超级链官方微信技术群以获取更多帮助。
> **任何不符合上述要求的issue，可能会在无任何通知的情况下被关闭。**

## 贡献代码

XuperChain 通过 GitHub 标准 [PR] (Pull Request)流程接受代码贡献。
请首先fork代码库，将修复代码提交到fork的代码库，并提交一个 PR 到 XuperChain。
代码维护人员会在收到PR后，对您提交的代码进行 Code Review，并在确认没问题后合入主干代码。

请使用下述模板提交 Pull Request：

* PR中代码修改的目的是什么？
* 简要描述您解决此问题的方案。

所有的 PR 需要对应到 Issue，并在 PR 描述中附带对应的 Issue 编号。
您提交的 PR 需要首先通过 CI 流程验证以降低引入新问题的可能，
之后还需要通过至少两位代码维护人员的Review， 这可能需要一定时间，请您耐心等待。

### 代码要求

- 代码注释请遵守 Golang 代码规范
- 所有代码必须具有单元测试
- 通过所有单元测试
- 请遵循提交代码的[一些约定](pr-guideline.md)

> **注意：**
>
> 您的代码必须遵循官方的Go语言代码规范(可以使用gofmt和Golint进行检查)。
>
> 更多文档贡献的内容请见：[如何贡献代码](contribute-codes.md)


## 贡献文档

XuperChain 提供专门 [Docs代码库][DocsRepo] 用于管理文档，通过 GitHub 标准 [PR] (Pull Request)流程接受文档贡献。
合入的 文档 PR 将通过 workflow 自动部署更新到 [文档][DocsSite] 网站中，帮助社区更好的理解 XuperChain。

> 文档基于 [Sphinx] 文档写作与托管平台，支持
> - 实时预览
> - 在线编辑
> - 中文搜索
> - 自动发布
> - 标签页视图
>
> 更多文档贡献的内容请见：[如何贡献文档](contribute-documents.md)

## 签署贡献者许可协议(Contributor License Agreement, CLA)

我们非常鼓励您向超级链(XuperChain)开源的代码库提交社区贡献。
为了保证社区用户能免费使用您贡献的开源代码，您需要首先签署一份贡献者许可协议。

您可以通过 [XuperChain Contributor License Agreement] 浏览并签署超级链贡献者许可协议，
您也可以在提交PR时通过自动回复的链接签署CLA。

活跃的贡献者可能会收到邀请，加入超级链核心贡献团队，并赋予您合入Pull Request的权限。

[Contact]: ../../../README-CN.md#联系我们
[XuperChain Contributor License Agreement]: https://cla-assistant.io/xuperchain/xuperchain
[Issue]: https://github.com/xuperchain/xuperchain/issues
[PR]: https://github.com/xuperchain/xuperchain/pulls
[Markdown]: https://guides.github.com/features/mastering-markdown/
[DocsSite]: https://xuper.baidu.com/n/xuperdoc/index.html
[DocsDir]: ../../../docs
[DocsRepo]: https://github.com/xuperchain/docs
[Sphinx]: https://github.com/sphinx-doc/sphinx
