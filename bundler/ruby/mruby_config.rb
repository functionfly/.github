# mruby build configuration for FunctionFly WASM runtime.
# Enables core gems needed for serverless function execution.
#
# This config is used by build.sh for both host (mrbc) and cross (WASM) builds.

MRuby::Build.new do |conf|
  toolchain :gcc

  conf.gem core: "mruby-compiler"
  conf.gem core: "mruby-bin-mrbc"
  conf.gem core: "mruby-error"
  conf.gem core: "mruby-math"
  conf.gem core: "mruby-time"
  conf.gem core: "mruby-json"
  conf.gem core: "mruby-string-ext"
  conf.gem core: "mruby-numeric-ext"
  conf.gem core: "mruby-array-ext"
  conf.gem core: "mruby-hash-ext"
  conf.gem core: "mruby-range-ext"
  conf.gem core: "mruby-proc-ext"
  conf.gem core: "mruby-symbol-ext"
  conf.gem core: "mruby-kernel-ext"
  conf.gem core: "mruby-object-ext"
  conf.gem core: "mruby-fiber"
  conf.gem core: "mruby-enumerator"
  conf.gem core: "mruby-enum-lazy"
  conf.gem core: "mruby-toplevel-ext"
  conf.gem core: "mruby-method"

  conf.cc.defines += %w(MRB_NO_STDIO)
end

MRuby::CrossBuild.new("wasm") do |conf|
  toolchain :clang

  conf.gem core: "mruby-compiler"
  conf.gem core: "mruby-error"
  conf.gem core: "mruby-math"
  conf.gem core: "mruby-time"
  conf.gem core: "mruby-json"
  conf.gem core: "mruby-string-ext"
  conf.gem core: "mruby-numeric-ext"
  conf.gem core: "mruby-array-ext"
  conf.gem core: "mruby-hash-ext"
  conf.gem core: "mruby-range-ext"
  conf.gem core: "mruby-proc-ext"
  conf.gem core: "mruby-symbol-ext"
  conf.gem core: "mruby-kernel-ext"
  conf.gem core: "mruby-object-ext"
  conf.gem core: "mruby-fiber"
  conf.gem core: "mruby-enumerator"
  conf.gem core: "mruby-enum-lazy"
  conf.gem core: "mruby-toplevel-ext"
  conf.gem core: "mruby-method"

  conf.cc.defines += %w(MRB_NO_STDIO)

  conf.cc.flags += %w(-O2)
  conf.linker.flags += %w(-O2)
end
