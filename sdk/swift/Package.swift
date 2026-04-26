// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "FunctionFly",
    products: [
        .library(name: "FunctionFly", targets: ["FunctionFly"]),
    ],
    targets: [
        .target(
            name: "FunctionFly",
            path: "Sources/FunctionFly",
            swiftSettings: [
                .unsafeFlags(["-Xfrontend", "-disable-reflection-metadata"])
            ]
        ),
    ]
)
