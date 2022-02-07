# 版本差异

[toc]

## mysql

### <= 5.6

- MySQL 5.5 及以前的版本，回滚日志是跟数据字典一起放在 ibdata 文件里的，即使长事务最终提交，回滚段被清理，文件也不会变小。

- MySQL 5.6 引入的索引下推优化。

- 每个 InnoDB 表数据存储在一个以 `.ibd` 为后缀的文件中,默认是开启状态

- MySQL 5.6 版本开始引入的 `Online DDL`

- MySQL 5.5 版本中引入了 `MDL`

- MySQL 5.6.6 版本开始，`innodb_file_per_table` 默认值就是 ON 了

- 截止到 MySQL 8.0，添加全文索引（FULLTEXT index）和空间索引 (SPATIAL index) 是 inplace 的 DDL，但不是 Online 的。

- MySQL 5.6 版本引入的一个新的排序算法，即：优先队列排序算法

- MySQL 5.6 版本引入了 GTID，**GTID 的全称是 Global Transaction Identifier，也就是全局事务 ID**

- MySQL 5.7 及之前的版本，自增值保存在内存里，并没有持久化



### 5.7 <

- MySQL 5.7 或更新版本，可以在每次执行一个比较大的操作后，通过执行 `mysql_reset_connection` 来重新初始化连接资源。这个过程不需要重连和重新做权限验证，但是会将连接恢复到刚刚创建完时的状态。

### 8.0

- 需要注意的是，MySQL 8.0 版本直接将查询缓存的整块功能删掉了，也就是说 8.0 开始彻底没有这个功能了。

    ```mysql
    select SQL_CACHE * from T where ID=10；
    ```

- MySQL 8.0 中，`innodb_flush_neighbors` 参数的默认值已经是 0 了

## redis

- 4.0 版本引入了 `unlink` 指令
- 4.0 给 `flushdb` 和 `flushall` 两个指令也带来了异步化，在指令后面增加 `async` 参数就可以将整棵大树连根拔起，扔给后台线程慢慢焚烧。
- 4.0 里引入了一个新的淘汰策略 —— **LFU 模式**，作者认为它比 LRU 更加优秀。  
LFU 的全称是 `Least Frequently Used`，表示按最近的访问频率进行淘汰，它比 LRU 更加精准地表示了一个 key 被访问的热度。

- Redis 4.0 提供了一个限流 Redis 模块，它叫 `redis-cell`。该模块也使用了漏斗算法，并提供了原子的限流指令。有了这个模块，限流问题就非常简单了。

### 5.0

- 5.0 又引入了一个新的数据结构 listpack，它是对 ziplist 结构的改进，在存储空间上会更加节省，而且结构上也比 ziplist 要精简。

- 5.0 引入了 stream
