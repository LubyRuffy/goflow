# goflow

Fofa的本质是数据，因此数据的编排是从获取Fofa的数据作为输入，经过用户的几次数据处理，最终输出为用户接受的格式。

因此，pipeline的模式就是为了完成数据编排，设计思路如下：
-   每一个编排的过程叫做工作流workflow；
-   每一个工作流之间通过文件进行数据传递；
-   上一个工作流的输出是下一个数据流的输入；
-   数据文件统一为json格式，每一行是一个对象（后续转换为zng格式？）；
-   目前只作为任务序列式的流程，暂时不考虑流式处理；
-   目前只做单机，不考虑分布式；
-   同一个文件的不同行格式（字段）允许不同；

## Features
- 内嵌底层函数
    - FetchFile 从文件获取数据
        -   file
        -   format 格式，支持csv/json
    - FetchFofa 从fofa获取数据
        -   query
        -   size
        -   fields
    - AddField 添加字段
        -   name 字段的名称
        -   设置数据，下面二选一
            -   value 直接赋值
            -   from，根据method决定处理方式
                -   method 方法
                    -   grep 正则处理，包括子串的提取
                -   field 字段
                -   value 参数值
    - RemoveField 删除字段
        - name 字段的名称
    - PieChart 生成Pie类型的报表
      - name 字段的名称
      - value 值的名称，如果是```count()```表明去重统计
      - size 取多少条，倒叙排序
      - title 报表标题
    - HttpRequest 执行http请求
- 支持缩写模式: ```fofa("body=icon && body=link", "body,host,ip,port") & grep_add("body", "(?is)<link[^>]*?rel[^>]*?icon[^>]*?>", "icon_tag") & drop("body")```
- （未完成）每一步都支持配置是否保留文件
- （未完成）函数可以进行统一化的参数配置
- 框架支持内嵌golang注册函数的扩展
- 框架支持动态加载扩展，golang的脚本语言
- 支持simple模式，将pipeline的模式转换成完整的golang代码
- 输出到不同的目标
- 可以保持中间数据，如aggs结果；不参与主流程，只用于统计，方便后续生成报表
- 可以形成报表
- 完整的日志记录
- 支持可视化流程展示
- （未完成）支持每一个步骤输出的格式预览
- 支持WebHook配置，回调事件
  - 支持finished事件
- 支持整体打包为gzip文件
- 支持用户自定义actionId，通过params参数传递即可，用于跟踪workflow的执行进度

## simple模式

按照如下规范进行设置：
- 用管道符号进行分隔：```cmd() & cmd2() & cmd3()```
- 参数支持多种格式：
    -   字符串
        -   双引号
        -   符号“`”
    -   HEX
    -   OCT
    -   INT
    -   bool：true/false
    -   null
- 支持嵌套：```cmd(cmd1())```
- 数据源命令：
    -   fofa(query, size, fields) 从fofa获取数据
    -   load(file) 从文件加载数据
    -   gen(jsonstring) 生成一行json，调试用
    -   scan_port(hosts,ports) 扫描端口（调用nmap）
- 目标地址命令：
    -   to_mysql(table,dsn,fields) 入库mysql，table必须填写；dns可选；如果没有那么就只生成sql文件；fields可选，如果没有，那么就从数据库中进行获取，没有配置dsn的话按照全字段
    -   to_sqlite(table,dsn,fields) 入库sqlite，table必须填写；dns可选；如果没有那么就只生成sql文件；fields可选，如果没有，那么就从数据库中进行获取，没有配置dsn的话按照全字段
    -   to_excel()
- 数据操作命令：
    - cut(fields) 只保留特定字段
    - drop(fields) 删除字段，rm也可以
    - grep_add(from_field, pattern, new_field_name) 通过对已有字段的正则提取到新的字段
    - to_int(field) 格式转换为int：```./fofa --verbose pipeline 'fofa(`title="test"`, `ip,port`) & to_int(`port`)'```
    - sort(field) 排序：```./fofa --verbose pipeline 'fofa(`title="test"`, `ip,port`) & to_int(`port`) & sort(`port`)'```
    - （未完成）set(field_name, value)
    - value(field) 取出值
    - flat(field) 把数组打平，去掉空值
    - stats(field, top_size) 统计计数：```./fofa --verbose pipeline 'fofa(`title="hacked"`,`title`, 1000) & stats("title",10)'```
    - uniq(true) 相邻的去重，注意：不会先排序
    - zq(query) 调用原始的zq语句
    - chart(type, title) 生成图表，支持pie/bar
    - fork(pipelines) 原始的手动创建分支的方式
    - screenshot(url) 网页截图
    - render_dom(url) 渲染dom生成html入到数据中
    - concat_add(field+":"+field2, newfield) 拼凑字符串，生成新的字段
    - fix_url(host) 解决host到url的转换
    - http_get(urlField) 请求url，生成http_status,http_header,http_body字段
    - where(filter) 选择过滤器
    - parse_url(urlField) 解析url为一个结构体
- 报表命令：
  - chart(type) type为pie,bar这样的；这里面要求必须是进行stats处理过后的统计结果
  - pie(field, value, top, title) value可以是```count()```表明按照数据字段打平了进行聚类统计，否则说明field每一个值都不一样，value是由另一个field字段进行定义的
  - bar(field, value, top, title) value可以是```count()```表明按照数据字段打平了进行聚类统计，否则说明field每一个值都不一样，value是由另一个field字段进行定义的
- 通过 ```[ cmd1() | cmd2() ]``` 创建分支
    -   分支的数据留是分开的，比如```fofa(`port=80`,`ip,port`) & [ cut(`ip`) | cut(`port`) ]```将会生成两条数据流
- 能够追踪执行进度
    -   日志
    -   单个进度
    -   记录错误

## 设计原则：
-   每一个底层函数对于奔溃的错误直接panic就好，由上层统一进行处理；
-   异常由底层函数决定是否奔溃，还是直接warning提示就好（比如某一行文件读取失败）； 所以需要设计全局的异常日志提示？
-   不同的数据格式如何处理？比如hostinfo的元数据和根据IP进行聚合的合并数据？
-   展现形式是什么？数据存储的格式如何设计？

## 一些场景

-   用一个复杂的语句来验证任务列表：
```
'fofa("body=icon && body=link", "body,host,ip,port", 500) & grep_add("body", "(?is)<link[^>]*?rel[^>]*?icon[^>]*?>", "icon_tag") & drop("body") & flat("icon_tag") & sort() & uniq(true) & sort("count") & zq("tail 10")'
```

-   生成两个饼图并且截图：
```
'fofa("title=test","host,ip,port,country", 1000) & [flat("port") & sort() & uniq(true) & sort("count") & zq("tail 10") & chart("pie") | flat("country") & sort() & uniq(true) & sort("count") & zq("tail 10") & chart("pie") | zq("tail 10") & screenshot("host") & to_excel()]'
```

-   四个并行：
```
fofa("title=test","host,ip,port,country", 1000) & [flat("port") & sort() & uniq(true) & sort("count") & zq("tail 10") & chart("pie") | flat("country") & sort() & uniq(true) & sort("count") & zq("tail 10") & chart("pie") | zq("tail 1") & screenshot("host") & to_excel() | to_sqlite("tbl", "host,ip,port")]
```

-   goby里面的实现：
```
scan_port(`10.10.10.0/24`, `80,443,22,445,3389`) & grab() & screenshot() & crawler() & tag() & vul()
scan_port(`10.10.10.0/24`, `22,80,6379,9200`) & grab() & tag() & vul()
```