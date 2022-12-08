# 如何贡献代码

本文将指导您如何进行代码开发。

## 代码要求

- 代码注释请遵守 Golang 代码规范
- 所有代码必须具有单元测试
- 通过所有单元测试
- 请遵循提交代码的[一些约定](pr-guideline.md)

## 代码开发流程

以下教程将指导您提交代码

1. [Fork](https://xuper.baidu.com/n/xuperdoc/contribution/pull_requests.html#fork)
2. 克隆 [Clone](https://xuper.baidu.com/n/xuperdoc/contribution/pull_requests.html#clone)
3. [创建本地分支](https://xuper.baidu.com/n/xuperdoc/contribution/pull_requests.html#id2)
4. 使用 pre-commit 钩子

   我们使用 [pre-commit][] 工具来管理 Git 预提交钩子。 它可以帮助我们格式化源代码，在提交（commit）前自动检查一些基本事宜。
   > 如每个文件只有一个 EOL，Git 中不要添加大文件等

   pre-commit 测试是 Travis-CI 中单元测试的一部分，不满足钩子的 PR 不能被提交到代码库。

	1. 安装并在当前目录运行 pre-commit：
	    ```bash
		$ pip install pre-commit
		$ pre-commit install
		```
	2. 使用 `gofmt` 来调整 golang源代码格式。

5. 使用 `license-eye` 工具

   [license-eye](http://github.com/apache/skywalking-eyes) 工具可以帮助我们检查和修复所有文件的证书声明， 在提交 (commit) 前证书声明都应该先完成。

   `license-eye` 检查是 Github-Action 中检测的一部分，检测不通过的 PR 不能被提交到代码库，安装使用它：

    ```bash
    $ make deps
    $ make license-check
    $ make license-fix
    ```

6. 编写代码

   通过 `git status` 查看当前状态，这会提示当前目录的一些变化，同时也可以通过 `git diff` 查看文件具体被修改的内容。

7. [构建和测试](../../../README-CN.md#构建)
8. 提交 Commit

   Git 每次提交代码，都需要写提交说明，这可以让其他人知道这次提交做了哪些改变，这可以通过`git commit` 完成。

9. 保持本地仓库最新

   在准备发起 Pull Request 之前，需要同步最新的代码。

   首先通过 [git remote][] 查看当前远程仓库的名字

    ``` shell
    $ git remote
    origin
    $ git remote -v
    origin	git@github.com:${USERNAME}/xuperchain.git (fetch)
    origin	git@github.com:${USERNAME}/xuperchain.git (push)
    ```

   这里`origin`是远程仓库名，引用了自己目录下的远程仓库。

   接下来我们创建一个命名为`upstream`的远程仓库名，引用原始仓库。

    ``` shell
    $ git remote add upstream git@github.com:xuperchain/xuperchain.git
    $ git remote
    origin
    upstream
    ```

   获取 upstream 的最新代码并更新当前分支。

    ``` shell
    $ git fetch upstream
    $ git pull upstream master
    ```

10. [Push] 到远程仓库

[pre-commit]: http://pre-commit.com/

[git remote]: https://git-scm.com/docs/git-remote

[Push]: https://xuper.baidu.com/n/xuperdoc/contribution/pull_requests.html#id3
