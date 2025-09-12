package cmd

import (
    "context"
    "flag"
    "fmt"
    "strings"
    "time"

    "github.com/google/subcommands"
    "gvisor.dev/gvisor/pkg/log"
    "gvisor.dev/gvisor/runsc/config"
    "gvisor.dev/gvisor/runsc/container"
)

// SetFilePolicy implements a direct per-sandbox file policy update without HTTP broadcast.
type SetFilePolicy struct {
    id   string
    path string
    perm string
}

func (*SetFilePolicy) Name() string     { return "set-file-policy" }
func (*SetFilePolicy) Synopsis() string { return "为指定运行中容器设置单条文件访问策略" }
func (*SetFilePolicy) Usage() string {
    return `set-file-policy --id=<container-id|name> --path=<绝对路径> --perm=<rw|ro|deny>
示例: runsc set-file-policy --id=ps-test --path=/test/test.txt --perm=deny
`
}

func (c *SetFilePolicy) SetFlags(f *flag.FlagSet) {
    f.StringVar(&c.id, "id", "", "容器 ID 或名称 (必填)")
    f.StringVar(&c.path, "path", "", "容器内文件绝对路径 (必填)")
    f.StringVar(&c.perm, "perm", "", "权限: rw|ro|deny (必填)")
}

func (c *SetFilePolicy) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus { // nolint: revive
    _ = ctx
    conf := args[0].(*config.Config)

    if c.id == "" || c.path == "" || c.perm == "" {
        fmt.Fprintf(f.Output(), "缺少必要参数，需同时提供 --id --path --perm\n")
        return subcommands.ExitUsageError
    }
    switch c.perm {
    case "rw", "ro", "deny":
    default:
        fmt.Fprintf(f.Output(), "无效权限: %s (允许: rw|ro|deny)\n", c.perm)
        return subcommands.ExitUsageError
    }

    // 解析传入 id: 允许 "name" 或 "sandbox/container" 两种形式。
    var full container.FullID
    if strings.Contains(c.id, "/") {
        parts := strings.SplitN(c.id, "/", 2)
        full = container.FullID{SandboxID: parts[0], ContainerID: parts[1]}
    } else {
        // sandboxID 与 containerID 相同（root container 情况），后续 Load 会做前缀匹配。
        full = container.FullID{SandboxID: c.id, ContainerID: c.id}
    }

    ctnr, err := container.Load(conf.RootDir, full, container.LoadOpts{Exact: false, SkipCheck: true})
    if err != nil {
        // 增强诊断：列出现有 sandbox（如果有）。
        fmt.Fprintf(f.Output(), "加载 sandbox 失败: %v\n", err)
        sandboxes, lerr := container.ListSandboxes(conf.RootDir)
        if lerr == nil {
            if len(sandboxes) == 0 {
                fmt.Fprintf(f.Output(), "当前 rootDir=%s 下没有任何运行中的 sandbox。\n", conf.RootDir)
                fmt.Fprintf(f.Output(), "请先启动一个使用 runsc 的容器，例如:\n  docker run --runtime=runsc -d --name policy-test -v /test:/test busybox sleep 600\n然后再次执行本命令。\n")
            } else {
                fmt.Fprintf(f.Output(), "可用 sandbox 列表 (ID):\n")
                for _, sb := range sandboxes {
                    fmt.Fprintf(f.Output(), "  %s\n", sb.SandboxID)
                }
            }
        } else {
            fmt.Fprintf(f.Output(), "列出 sandbox 失败: %v\n", lerr)
        }
        return subcommands.ExitFailure
    }
    if ctnr == nil || ctnr.Sandbox == nil || !ctnr.IsSandboxRunning() {
        fmt.Fprintf(f.Output(), "目标 sandbox 未运行: %s\n", c.id)
        sandboxes, lerr := container.ListSandboxes(conf.RootDir)
        if lerr == nil && len(sandboxes) > 0 {
            fmt.Fprintf(f.Output(), "当前运行中的 sandbox: \n")
            for _, sb := range sandboxes {
                fmt.Fprintf(f.Output(), "  %s\n", sb.SandboxID)
            }
        }
        return subcommands.ExitFailure
    }

    start := time.Now()
    if err := ctnr.Sandbox.PolicySetFilePermission(c.path, c.perm); err != nil {
        fmt.Fprintf(f.Output(), "下发策略失败: %v\n", err)
        return subcommands.ExitFailure
    }
    elapsed := time.Since(start)
    ts := start.Format(time.RFC3339Nano)
    log.Infof("set-file-policy ok: id=%s path=%s perm=%s at=%s duration=%s", c.id, c.path, c.perm, ts, elapsed)
    fmt.Fprintf(f.Output(), "策略已应用: %s %s => %s at=%s duration=%s\n", c.id, c.path, c.perm, ts, elapsed)
    return subcommands.ExitSuccess
}
