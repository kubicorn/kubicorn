<<<<<<< HEAD
load("@io_bazel_rules_go//go:def.bzl", "go_library")
=======
package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)
>>>>>>> Initial dep workover

go_library(
    name = "go_default_library",
    srcs = [
<<<<<<< HEAD
        "initializerconfiguration.go",
        "interface.go",
    ],
    importpath = "k8s.io/client-go/informers/admissionregistration/v1alpha1",
    visibility = ["//visibility:public"],
=======
        "externaladmissionhookconfiguration.go",
        "initializerconfiguration.go",
        "interface.go",
    ],
>>>>>>> Initial dep workover
    deps = [
        "//vendor/k8s.io/api/admissionregistration/v1alpha1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/watch:go_default_library",
        "//vendor/k8s.io/client-go/informers/internalinterfaces:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/listers/admissionregistration/v1alpha1:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
<<<<<<< HEAD
    visibility = ["//visibility:public"],
=======
>>>>>>> Initial dep workover
)
