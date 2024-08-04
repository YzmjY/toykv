# TOYKV
`toykv`是一个个人学习项目，使用GO语言，实现一个基于LSM树的单机KV数据存储引擎，主要用来学习LSM树、Compact、KV分离等特性的实现。

代码实现大量参考[badger](https://github.com/dgraph-io/badger),由于精力有限，对于事务等高级特性进行了舍弃。

# Note
学习各部分实现时的笔记记录。
| part| note|
|:-----:|:----:|
|Skiplist|[skiplist](./doc/skiplist/note.md)|
|BloomFilter|[bloomfilter](./doc/bloomfilter/note.md)|

# Example

