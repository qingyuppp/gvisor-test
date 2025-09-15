### 编译、复制和安装可执行文件
mkdir -p bin
make copy TARGETS=runsc DESTINATION=bin/
sudo cp ./bin/runsc /usr/local/bin

### 运行容器
docker run --runtime=runsc -d --name policy-test -v /test:/test busybox sleep 600

### 下发策略 rw/deny/ro
sudo runsc --root=/var/run/docker/runtime-runc/moby set-file-policy   --id=d6255ae7f5fb44b00d1fea27efc8fbfa97d842d56da7139abf0f40f4c845b78b   --path=/test/test.txt --perm=rw

### 读写测试
docker exec policy-test sh -c 'echo AAA >> /test/test.txt'


### 方案逻辑简要总结：

1.核心目的
- 在运行中的 gVisor sandbox 内动态控制特定路径的文件访问权限（rw / ro / deny），使后续新打开操作即时受限。

2.架构组成

- 策略存储：Sentry 内全局 GlobalFilePolicyManager（内存 map[path]perm）。
- 强制点：VFS OpenAt 钩子在真正打开文件前检查策略。
- 下发通道： a) HTTP + 广播（policy-server 子命令，遍历所有 sandbox 逐个 RPC）。
b) 精简命令 set-file-policy 直接对单个 sandbox 发送控制 RPC（当前推荐测试路径）。
- 控制 RPC：Policy.SetFilePermission 更新 Sentry 内存策略。
3.执行链（set-file-policy）
Host 运行命令 → 解析并定位 sandbox（通过 state 文件）→ 发送 Policy RPC → Sentry 更新 map → 后续容器内对该路径的 open 根据 perm 决定允许或拒绝：
- rw：正常
- ro：阻止写（O_WRONLY/O_RDWR/创建/截断）
- deny：阻止任何 open
4.生效特性
- 仅影响策略应用后新的文件打开；已持有的 fd 不回溯。
- 多容器需逐个 sandbox 下发或使用 policy-server 广播。
- 内存态，不持久化，重启需重新下发。
5.关键文件
- file_policy.go：策略管理
- vfs.go：访问拦截
- policy.go：RPC 入口
- sandbox.go：RPC 客户端调用封装
- set_file_policy.go：CLI 命令
- policy_server.go & file_policy_server.go：HTTP 广播（可选路径）
6.使用要点
- 必须用正确 --root 指定实际 runsc root（从 runsc-sandbox 进程的 --root= 抓取）。
- 路径必须是容器内部视角的绝对路径。
- 失败时查看增强的命令输出确定 sandbox 是否匹配。
- 可扩展方向（未实现）
- 列出已设策略
-前缀/通配符匹配
- 持久化/加载
- 更细粒度权限（如 exec 单独控制）