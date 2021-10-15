### Build

```bash
make build-agent
```
### Run

```bash
bin/sms-agent -rules SUBSYSTEM=block -disable-udev-event
```

### TODO

* netlink压测，保证不丢消息
* 动态配置，调整日志级别
* 使用libdevmapper操作, 避开dmsetup
* 更新检查，不允许盘内扩容
* 上报查询设备详情
* 判断文件系统类型
* 文件浏览器
* 使用sg3_utils-devel，调用C API
* 启动后先更新udev硬件元数据库，保证数据是最新的
* 补偿机制：定期全量扫描


### Feature

#### 安全
* 支持文件锁，保证每台机器上只会有一个agent在运行

