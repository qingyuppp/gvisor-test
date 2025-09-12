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

package control

import (
    "fmt"
    "time"
    "gvisor.dev/gvisor/pkg/log"
    "gvisor.dev/gvisor/pkg/sentry/policy"
)

// PolicySetArgs defines the payload to set a file permission policy in Sentry.
type PolicySetArgs struct {
    Path       string
    Permission string // "ro", "rw", "deny"
}

// Policy provides RPCs to manage runtime file access policy inside Sentry.
type Policy struct{}

// SetFilePermission updates GlobalFilePolicyManager at runtime.
func (*Policy) SetFilePermission(args *PolicySetArgs, _ *struct{}) error {
    if args == nil {
        return fmt.Errorf("nil args")
    }
    if policy.GlobalFilePolicyManager == nil {
        policy.GlobalFilePolicyManager = policy.NewFilePolicyManager()
    }
    perm := policy.FilePermission(args.Permission)
    switch perm {
    case policy.PermissionReadOnly, policy.PermissionReadWrite, policy.PermissionDeny:
        // ok
    default:
        return fmt.Errorf("invalid permission: %q", args.Permission)
    }
    start := time.Now()
    policy.GlobalFilePolicyManager.SetPermission(args.Path, perm)
    log.Infof("Policy RPC applied path=%s perm=%s at=%s duration=%s", args.Path, perm, start.Format(time.RFC3339Nano), time.Since(start))
    return nil
}
