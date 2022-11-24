# 如何贡献文档

## 文档要求
- 所有内容都应该以 [Markdown] (GitHub风格)的形式编写，文件以`.md`为后缀
- 如果是新增文档，需将新增的文档名，添加到对应的index文件中（[SUMMARY.md](../SUMMARY.md)）

## 文档开发流程

1. 编写文档
1. 运行预览工具，并预览修改
    - [如何使用预览工具](#如何使用预览工具)
1. 提交文档修改
    - 提交修改与 PR 的步骤可以参考[代码提交指南][ContributeCode]

### 如何使用预览工具

1. 安装依赖项
   
   在此之前，请确认您的操作系统安装了 `gitbook` 及依赖项。
   
   以ubuntu系统为例，运行：
   
   ``` shell
   $ sudo apt-get update && apt-get install -y npm
   $ sudo npm install -g gitbook-cli
   ```

2. [Clone仓库](https://xuper.baidu.com/n/xuperdoc/contribution/pull_requests.html#clone)
3. 在本地运行文档站点

   进入您希望加载和构建内容的目录列表（`/docs/<LANG>`）, 运行：
   
   ``` shell
   $ cd docs/zh_cn/
   $ gitbook serve --port 8000
   ...
   Serving book on http://localhost:8000
   ```

4. 预览效果

   打开浏览器并导航到 http://localhost:8000。
   
   网站可能需要几秒钟才能成功加载，因为构建需要一定的时间


[Markdown]: https://guides.github.com/features/mastering-markdown/
[ContributeCode]: https://xuper.baidu.com/n/xuperdoc/contribution/pull_requests.html#