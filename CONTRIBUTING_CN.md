# XuperUnion开源贡献指南

[ [English](https://github.com/xuperchain/xuperunion/blob/master/CONTRIBUTING.md) | 简体中文 ]


首先，非常感谢您考虑向XuperUnion提交开源贡献。

在此之前，请阅读下面的开源贡献指南，并遵守issue和PR提交规则。为了大家能在github上无障碍沟通，我们强烈建议您使用英文提交issue或PR。

## 创建 Issue
XuperUnion使用github的issue系统来管理`bug`和`功能提议`。在您创建一个新的issue之前，请首先通过issue搜索查看是否已经存在相同或类似的issue，如果您的情况跟已存在的issue类似，您只需要在已存在的issue中通过回复追加您的情况即可。如果您确认不存在类似issue，那么别犹豫，请创建一个新的issue。

在创建issue时，请**使用**我们已经提供的**issue模板**，并在模板中填写下述内容：

* 简要的描述issue的主要问题。
* 您的系统环境。 包括XuperUnion的版本，OS/Go/Gcc 版本等信息。
* 请描述您认为应该发生的情况。
* 请描述您实际观察到的情况。
* 请提供日志、截图等可能有用的线索，如果可能，请提供一份最小可验证的代码样例帮助我们确认问题。

**注意：** Issue系统只面向提交bug和功能建议使用。如果您是想咨询技术细节或其他问题，请首先查看我们的wiki，如果wiki没有解决您的问题，请通过首页提供的微信二维码加入超级链官方微信技术群以获取更多帮助。**任何不符合上述要求的issue，可能会在无任何通知的情况下被关闭。**

## 提交代码
XuperUnion通过github标准PR(Pull Request)流程接受代码贡献。请首先fork代码库，将修复代码提交到fork的代码库，并提交一个PR到XuperUnion。代码维护人员会在收到PR后，对您提交的代码进行Code Review，并在确认没问题后合入主干代码。

请使用下述模板提交pull request：

* PR中代码修改的目的是什么？
* 简要描述您解决此问题的方案。

所有的PR需要对应到issue，并在PR描述中附带对应的issue编号。您提交的PR需要首先通过CI流程验证以降低引入新问题的可能，之后还需要通过至少两位代码维护人员的Review，这可能需要一定时间，请您耐心等待。

**注意：** 您的代码必须遵循官方的Go语言代码规范(可以使用gofmt和golint进行检查)。

## 签署贡献者许可协议(Contributor License Agreement, CLA)
我们非常鼓励您向超级链(XuperChain)开源的代码库提交社区贡献，为了保证社区用户能免费使用您贡献的开源代码，您需要首先签署一份贡献者许可协议。

您可以通过[XuperChain Contributor License Agreement](https://cla-assistant.io/xuperchain/xuperunion)浏览并签署超级链贡献者许可协议，您也可以在提交PR时通过自动回复的链接签署CLA。

活跃的贡献者可能会收到邀请，加入超级链核心贡献团队，并赋予您合入Pull Request的权限。
