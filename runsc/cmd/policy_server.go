// Copyright 2025 The gVisor Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
    "context"

    "github.com/google/subcommands"
    "gvisor.dev/gvisor/pkg/log"
    "gvisor.dev/gvisor/pkg/sentry/policy"
    "gvisor.dev/gvisor/runsc/cmd/util"
    "gvisor.dev/gvisor/runsc/config"
    "gvisor.dev/gvisor/runsc/container"
    "gvisor.dev/gvisor/runsc/flag"
)

// PolicyServer implements subcommands.Command for the "policy-server" command.
type PolicyServer struct {
    addr string
}

// Name implements subcommands.Command.Name.
func (*PolicyServer) Name() string { return "policy-server" }

// Synopsis implements subcommands.Command.Synopsis.
func (*PolicyServer) Synopsis() string { return "start file policy HTTP server (host daemon)" }

// Usage implements subcommands.Command.Usage.
func (*PolicyServer) Usage() string {
    return "policy-server [--addr=:9090] - starts a host HTTP server to manage file access policies\n"
}

// SetFlags implements subcommands.Command.SetFlags.
func (p *PolicyServer) SetFlags(f *flag.FlagSet) {
    f.StringVar(&p.addr, "addr", ":9090", "Address to listen on, e.g. :9090 or 127.0.0.1:9090")
}

// Execute implements subcommands.Command.Execute.
func (p *PolicyServer) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
    _ = ctx
    conf := args[0].(*config.Config)

    // Ensure a global policy manager exists.
    if policy.GlobalFilePolicyManager == nil {
        policy.GlobalFilePolicyManager = policy.NewFilePolicyManager()
    }

    // 启动前先枚举一次 sandbox，帮助诊断 rootDir 是否匹配。
    initialIDs, initialErr := container.ListSandboxes(conf.RootDir)
    if initialErr != nil {
        log.Warningf("pre-start list sandboxes failed rootDir=%s: %v", conf.RootDir, initialErr)
    } else {
        if len(initialIDs) == 0 {
            log.Infof("policy-server pre-start: rootDir=%s (发现 0 个 sandbox)", conf.RootDir)
        } else {
            log.Infof("policy-server pre-start: rootDir=%s sandboxes=%v", conf.RootDir, initialIDs)
        }
    }

    // onUpdate: broadcast to all running sandboxes via control RPC，并打印结果。
    onUpdate := func(path, perm string) {
        ids, err := container.ListSandboxes(conf.RootDir)
        if err != nil {
            log.Warningf("onUpdate list sandboxes failed: %v", err)
            return
        }
        if len(ids) == 0 {
            log.Infof("onUpdate no sandboxes found rootDir=%s path=%s perm=%s (广播跳过)", conf.RootDir, path, perm)
        }
        for _, id := range ids {
            c, err := container.Load(conf.RootDir, id, container.LoadOpts{Exact: true, SkipCheck: true})
            if err != nil {
                log.Warningf("onUpdate load sandbox %v failed: %v", id, err)
                continue
            }
            if !c.IsSandboxRunning() {
                log.Infof("onUpdate sandbox %s not running, skip", c.Sandbox.ID)
                continue
            }
            if err := c.Sandbox.PolicySetFilePermission(path, perm); err != nil {
                log.Warningf("onUpdate RPC sandbox=%s path=%s perm=%s failed: %v", c.Sandbox.ID, path, perm, err)
            } else {
                log.Infof("onUpdate applied sandbox=%s path=%s perm=%s", c.Sandbox.ID, path, perm)
            }
        }
    }

    log.Infof("Starting policy server on %s (rootDir=%s)", p.addr, conf.RootDir)
    if err := policy.StartFilePolicyServer(policy.GlobalFilePolicyManager, p.addr, onUpdate); err != nil {
        util.Fatalf("policy server failed: %v", err)
        return subcommands.ExitFailure
    }
    return subcommands.ExitSuccess
}
