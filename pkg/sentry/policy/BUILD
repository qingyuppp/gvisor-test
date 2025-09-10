load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "policy",
    srcs = [
        "file_policy.go",
        "file_policy_server.go",
    ],
    importpath = "qingyuppp/gvisor-test/pkg/sentry/policy",
    visibility = ["//visibility:public"],
    deps = [
        # 标准库不需要写，只有你用到的其它 go_library 规则才写，比如有用到第三方库或其它 Bazel 包时补全
    ],
)