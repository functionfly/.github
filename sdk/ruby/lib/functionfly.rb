# FunctionFly Ruby SDK
require_relative "functionfly/version"
require_relative "functionfly/function"
require_relative "functionfly/context"

module FunctionFly
  def self.run(function_class)
    input = $stdin.read
    ctx = Context.new
    fn = function_class.new
    output = fn.handle(input, ctx)
    $stdout.write(output)
  rescue => e
    $stderr.puts "[functionfly] Error: #{e.message}"
    puts '{"error": {"code": "RUNTIME_ERROR", "message": "' + e.message.gsub('"', '\\"') + '"}}'
    exit 1
  end
end
