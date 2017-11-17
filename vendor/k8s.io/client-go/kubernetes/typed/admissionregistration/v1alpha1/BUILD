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
        "admissionregistration_client.go",
        "doc.go",
<<<<<<< HEAD
        "generated_expansion.go",
        "initializerconfiguration.go",
    ],
    importpath = "k8s.io/client-go/kubernetes/typed/admissionregistration/v1alpha1",
    visibility = ["//visibility:public"],
=======
        "externaladmissionhookconfiguration.go",
        "generated_expansion.go",
        "initializerconfiguration.go",
    ],
>>>>>>> Initial dep workover
    deps = [
        "//vendor/k8s.io/api/admissionregistration/v1alpha1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/serializer:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/types:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/watch:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes/scheme:go_default_library",
        "//vendor/k8s.io/client-go/rest:go_default_library",
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
    srcs = [
        ":package-srcs",
        "//staging/src/k8s.io/client-go/kubernetes/typed/admissionregistration/v1alpha1/fake:all-srcs",
    ],
    tags = ["automanaged"],
<<<<<<< HEAD
    visibility = ["//visibility:public"],
=======
>>>>>>> Initial dep workover
)
