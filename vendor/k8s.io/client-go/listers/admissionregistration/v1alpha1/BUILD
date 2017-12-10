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
        "expansion_generated.go",
<<<<<<< HEAD
        "initializerconfiguration.go",
    ],
    importpath = "k8s.io/client-go/listers/admissionregistration/v1alpha1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/api/admissionregistration/v1alpha1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/errors:go_default_library",
=======
        "externaladmissionhookconfiguration.go",
        "initializerconfiguration.go",
    ],
    deps = [
        "//vendor/k8s.io/api/admissionregistration/v1alpha1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/errors:go_default_library",
<<<<<<< HEAD
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
>>>>>>> Initial dep workover
=======
>>>>>>> omg dep constraints
        "//vendor/k8s.io/apimachinery/pkg/labels:go_default_library",
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
