load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "policy",
    srcs = [
        "file_policy.go",
        "file_policy_server.go",
    ],
    importpath = "gvisor.dev/gvisor/pkg/sentry/policy",
    visibility = ["//visibility:public"],
    deps = [
    "//pkg/log",
    ],
)