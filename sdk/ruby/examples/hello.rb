# Example FunctionFly function in Ruby.
# Build: mruby main.rb -o hello.wasm (or use the FunctionFly bundler)
require_relative "../lib/functionfly"

class HelloFunction < FunctionFly::Function
  def handle(input, ctx)
    ctx.log("Hello from Ruby function!")
    '{"message": "Hello from Ruby!", "ok": true}'
  end
end

FunctionFly.run(HelloFunction)
